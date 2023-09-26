package main

import (
	"bytes"
	"ebpf-based-cni/defaults"
	"errors"
	"fmt"
	"strconv"
	"strings"

	// "net/http"
	"net/http"
	"net/url"
	"path"

	"net"
	"os"
	"runtime"

	"encoding/json"

	cniTypes "ebpf-based-cni/types"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	cniVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type NetConf struct {
	types.NetConf
	IPMasq bool `json:"ipMasq"`
	MTU    int  `json:"mtu"`
}

var (
	logger *log.Logger
	f      *os.File
	ospp_host_ip net.IP
	client *http.Client
)

const (
	OSPPHostName = "ospp_host"
	OSPPNetName  = "ospp_net"
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
	logger = log.New()
	// logger.SetFormatter()
	logger.SetLevel(log.DebugLevel)
	var err error
	f, err = os.OpenFile("/root/summer-ospp/cni.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Errorf("open cni log file err %v", err))
	}
	logger.SetOutput(f)
	client = &http.Client{}
}


func main() {
	skel.PluginMain(cmdAdd,
		cmdCheck,
		cmdDel,
		cniVersion.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1", "0.4.0", "1.0.0"),
		"ebpf-based CNI plugin 0.0.1")
}

func cmdAdd(args *skel.CmdArgs) (err error) {
	ep := new(cniTypes.Endpoint)
	conf := NetConf{}

	defer f.Close()
	// var osppHostVeth *netlink.Veth
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}
	// run the IPAM plugin and get back the config to apply
	result, err := execIPAMGetResult(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return fmt.Errorf("ipam get ip result err: %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	hostInterface, containerInterface,containerIP, err := setupContainerVeth(netns, args.IfName, conf.MTU, result)
	if err != nil {
		return err
	}
	ep.IPv4 = containerIP
	mac,err := net.ParseMAC(containerInterface.Mac)
	if err != nil {
		return fmt.Errorf("parse conatiner mac err: %v",err)
	}
	ep.Mac = mac
	nodeMac,err := net.ParseMAC(hostInterface.Mac)
	if err != nil {
		return fmt.Errorf("parse conatiner node mac err: %v",err)
	}
	ep.NodeMAC = nodeMac
	ep.IfName = containerInterface.Name
	
	hostVeth, err := GetVethFromLink(hostInterface.Name)
	if err != nil {
		return fmt.Errorf("get host veth err: %v",err)
	}
	ep.IfIndex = hostVeth.Index

	if conf.IPMasq {
		chain := utils.FormatChainName(conf.Name, args.ContainerID)
		comment := utils.FormatComment(conf.Name, args.ContainerID)
		for _, ipc := range result.IPs {
			if err = ip.SetupIPMasq(&ipc.Address, chain, comment); err != nil {
				return err
			}
		}
	}

	// Only override the DNS settings in the previous result if any DNS fields
	// were provided to the ptp plugin. This allows, for example, IPAM plugins
	// to specify the DNS settings instead of the ptp plugin.
	if dnsConfSet(conf.DNS) {
		result.DNS = conf.DNS
	}

	// send request attach bpf for veth ingress
	resp, err := sendVethReq(hostInterface.Name)
	if err != nil {
		logger.Error(err)
		return err
	}
	if !strings.EqualFold(resp, "") {
		logger.Info(resp)
	}
	// send request to create endpoint on local agent
	resp,err = sendEPCreate(ep)
	if err != nil {
		logger.Error(err)
		return err
	}
	if !strings.EqualFold(resp, "") {
		logger.Info(resp)
	}
	return types.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {

	conf := NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}

	if err := ipam.ExecDel(conf.IPAM.Type, args.StdinData); err != nil {
		return err
	}

	if args.Netns == "" {
		return nil
	}

	// There is a netns so try to clean up. Delete can be called multiple times
	// so don't return an error if the device is already removed.
	// If the device isn't there then don't try to clean up IP masq either.
	var ipnets []*net.IPNet
	err := ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		var err error
		ipnets, err = ip.DelLinkByNameAddr(args.IfName)
		if err != nil && err == ip.ErrLinkNotFound {
			return nil
		}
		return err
	})
	if err != nil {
		//  if NetNs is passed down by the Cloud Orchestration Engine, or if it called multiple times
		// so don't return an error if the device is already removed.
		// https://github.com/kubernetes/kubernetes/issues/43014#issuecomment-287164444
		_, ok := err.(ns.NSPathNotExistErr)
		if ok {
			return nil
		}
		return err
	}

	if len(ipnets) != 0 && conf.IPMasq {
		chain := utils.FormatChainName(conf.Name, args.ContainerID)
		comment := utils.FormatComment(conf.Name, args.ContainerID)
		for _, ipn := range ipnets {
			err = ip.TeardownIPMasq(ipn, chain, comment)
		}
	}

	return err

}

func cmdCheck(args *skel.CmdArgs) error {

	return nil
}

func execIPAMGetResult(typeIpam string, stdinData []byte) (*current.Result, error) {
	// run the IPAM plugin and get back the config to apply
	r, err := ipam.ExecAdd(typeIpam, stdinData)
	if err != nil {
		return nil, err
	}

	// Invoke ipam del if err to avoid ip leak
	defer func() {
		if err != nil {
			ipam.ExecDel(typeIpam, stdinData)
		}
	}()

	// Convert whatever the IPAM result was into the current Result type
	result, err := current.NewResultFromResult(r)
	if err != nil {
		return nil, err
	}

	if len(result.IPs) == 0 {
		return nil, errors.New("IPAM plugin returned missing IP config")
	}
	if err := ip.EnableForward(result.IPs); err != nil {
		return nil, fmt.Errorf("could not enable IP forwarding: %v", err)
	}
	return result, nil
}
func setupContainerVeth(netns ns.NetNS, ifName string, mtu int, pr *current.Result) (*current.Interface, *current.Interface,net.IP, error) {
	// The IPAM result will be something like IP=192.168.3.5/24, GW=192.168.3.1.
	// What we want is really a point-to-point link but veth does not support IFF_POINTTOPOINT.
	// Next best thing would be to let it ARP but set interface to 192.168.3.5/32 and
	// add a route like "192.168.3.0/24 via 192.168.3.1 dev $ifName".
	// Unfortunately that won't work as the GW will be outside the interface's subnet.

	// Our solution is to configure the interface with 192.168.3.5/24, then delete the
	// "192.168.3.0/24 dev $ifName" route that was automatically added. Then we add
	// "192.168.3.1/32 dev $ifName" and "192.168.3.0/24 via 192.168.3.1 dev $ifName".
	// In other words we force all traffic to ARP via the gateway except for GW itself.

	hostInterface := &current.Interface{}
	containerInterface := &current.Interface{}
	var containerIP net.IP
	osppVeth, err := GetVethFromLink(OSPPHostName)
	if err != nil {
		return nil, nil,nil, err
	}
	addrs, err := netlink.AddrList(osppVeth, netlink.FAMILY_V4)
	if err != nil {
		return nil, nil, nil,err
	}

	for _, addr := range addrs {
		// logger.Printf("find addr is %s/%d", addr.IPNet.IP, addr.IPNet.Mask)
		ospp_host_ip = addr.IPNet.IP
		break
	}

	err = netns.Do(func(hostNS ns.NetNS) error {
		hostVeth, contVeth0, err := ip.SetupVeth(ifName, mtu, "", hostNS)
		if err != nil {
			return err
		}
		hostInterface.Name = hostVeth.Name
		hostInterface.Mac = hostVeth.HardwareAddr.String()
		containerInterface.Name = contVeth0.Name
		containerInterface.Mac = contVeth0.HardwareAddr.String()
		containerInterface.Sandbox = netns.Path()

		for _, ipc := range pr.IPs {
			// All addresses apply to the container veth interface
			ipc.Interface = current.Int(1)
		}

		pr.Interfaces = []*current.Interface{hostInterface, containerInterface}

		contVeth, err := net.InterfaceByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to look up %q: %v", ifName, err)
		}

		if err = ipam.ConfigureIface(ifName, pr); err != nil {
			return err
		}
		// logger.Println(osppVeth.IP)

		for _, ipc := range pr.IPs {
			// Delete the route that was automatically added
			route := netlink.Route{
				LinkIndex: contVeth.Index,
				Dst: &net.IPNet{
					IP:   ipc.Address.IP.Mask(ipc.Address.Mask),
					Mask: ipc.Address.Mask,
				},
				Scope: netlink.SCOPE_NOWHERE,
			}

			if err := netlink.RouteDel(&route); err != nil {
				return fmt.Errorf("failed to delete route %v: %v", route, err)
			}
			containerIP = ipc.Address.IP

			addrBits := 32
			if ipc.Address.IP.To4() == nil {
				addrBits = 128
			}
			/*
			   default via cilium_host_ip dev eth0 mtu 1500
			   cilium_host_ip dev eth0 scope link
			*/
			for _, r := range []netlink.Route{
				{
					LinkIndex: contVeth.Index,
					Dst: &net.IPNet{
						IP:   ospp_host_ip,
						Mask: net.CIDRMask(addrBits, addrBits),
					},
					Scope: netlink.SCOPE_LINK,
					// Src:   ipc.Address.IP,
				},
				/* create default route */
				{
					LinkIndex: contVeth.Index,
					Scope: netlink.SCOPE_UNIVERSE,
					Gw:    ospp_host_ip,
					MTU: 1500,
					// Src:   ipc.Address.IP,
				},
			} {
				if err := netlink.RouteAdd(&r); err != nil {
					return fmt.Errorf("failed to add route %v: %v", r, err)
				}
			}
		}
		/* add arp entry */
		/* cilium_host_ip ether	veth11	mac_addr C eth0*/
		arpEntry := &netlink.Neigh{
			LinkIndex:    contVeth.Index,
			State:        netlink.NUD_PERMANENT,
			Family:       netlink.FAMILY_V4,
			Flags:        netlink.NTF_SELF,
			IP:           ospp_host_ip,
			HardwareAddr: hostVeth.HardwareAddr,
		}
		if err := netlink.NeighAdd(arpEntry); err != nil {
			return fmt.Errorf("add arp entry err %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil,nil, err
	}
	return hostInterface, containerInterface, containerIP,nil
}

func dnsConfSet(dnsConf types.DNS) bool {
	return dnsConf.Nameservers != nil ||
		dnsConf.Search != nil ||
		dnsConf.Options != nil ||
		dnsConf.Domain != ""
}

func GetVethFromLink(linkName string) (*netlink.Veth, error) {
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		return nil, err
	}
	return link.(*netlink.Veth), nil
}

/* http call is here */

func sendVethReq(vethName string) (string, error) {

	// 将vethName添加到路径中
	requestPath := path.Join("veth", vethName)

	fullURL, err := url.Parse(defaults.BaseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %v", err)
	}
	fullURL.Path = requestPath
	// 发起GET请求
	resp, err := http.Get(fullURL.String())
	if err != nil {
		return "", fmt.Errorf("error sending GET request: %v", err)
	}
	defer resp.Body.Close()
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	return string(buf[:n]), nil
}

func sendEPCreate(ep *cniTypes.Endpoint) (string,error) {
	
	body, _:= json.Marshal(ep)
    req, err := http.NewRequest("POST", defaults.BaseURL+"/ep", bytes.NewBuffer(body))
	if err != nil {
		return "",err
	}
    req.Header.Set("Content-Type", "application/json")
    resp, err := client.Do(req)
    if err != nil {
        return "",err
    }
    defer resp.Body.Close()
    return strconv.Itoa(resp.StatusCode),nil
}

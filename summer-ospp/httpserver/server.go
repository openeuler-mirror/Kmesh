package main

import (
	"ebpf-based-cni/httpserver/bpf"
	"ebpf-based-cni/httpserver/handlers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/vishvananda/netlink"
)

type NetConf struct {
	types.NetConf
	IPMasq bool `json:"ipMasq"`
	MTU    int  `json:"mtu"`
	OsppHost string `json:"ospp_host"`
}
const (
	VethHostName = "ospp_host"
	VethNetName  = "ospp_net"
)



func main() {
    app := fiber.New()

	err_ospp := initAgent();
	if err_ospp != nil {
		log.Fatal(fmt.Errorf("failed to init ospp agent: %v", err_ospp))
	}
	
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello, World 👋!")
    })

	app.Get("/veth/:vethName", func(c *fiber.Ctx) error {
		vethName := c.Params("vethName")
		if err := bpf.LoadTCIngressProgToVeth(vethName);err != nil {
			return c.SendString(err.Error())
		}
		if err := bpf.LoadTCEgressProgToVeth(vethName); err != nil {
			return c.SendString(fmt.Sprintf("load tc egress err %v",err.Error()))
		}
		return c.SendString("add tc prog to veth ok");
	})

	app.Post("/ep",handlers.CreateEP)

    go func ()  {
		log.Fatal(app.Listen(":3000"))
	}()
	handleExitSignal()
}

func initAgent() error {
	file, err := os.Open("/etc/cni/net.d/network.conf")
	if err != nil {
		return fmt.Errorf("parse network conf err")
	}
	defer file.Close()
	data,err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("can not read json file %v",err)
	}
	conf := NetConf{}
	if err := json.Unmarshal(data, &conf); err != nil {
		log.Fatal(fmt.Errorf("failed to load netconf: %v", err))
	}
	log.Println(conf.Name)
	hostVeth, _ := netlink.LinkByName(VethHostName)
	netVeth, _ := netlink.LinkByName(VethNetName)
	if hostVeth != nil && netVeth != nil {
		return nil
	}
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: VethHostName,
			MTU:  1500,
		},
		PeerName: VethNetName,
	}
	if err := netlink.LinkAdd(veth); err != nil {
		return err
	}

	veth1, err := netlink.LinkByName(VethHostName)
	if err != nil {
		netlink.LinkDel(veth)
		return err
	}

	veth2, err := netlink.LinkByName(VethNetName)
	if err != nil {
		netlink.LinkDel(veth)
		return err
	}

	if err := netlink.LinkSetUp(veth1.(*netlink.Veth)); err != nil {
		return fmt.Errorf("failed to setup osppHostVeth: %v", err)
	}

	if err := netlink.LinkSetUp(veth2.(*netlink.Veth)); err != nil {
		return fmt.Errorf("failed to setup osppNetVeth: %v", err)
	}
	ip_str := conf.OsppHost
	ip32 := fmt.Sprintf("%s/%s", ip_str, "32")

	ipAddr, ipNet, err := net.ParseCIDR(ip32)
	if err != nil {
		return err
	}
	ipNet.IP = ipAddr
	link, err := netlink.LinkByName(VethHostName)
	if err != nil {
		return err
	}
	if err := netlink.AddrAdd(link, &netlink.Addr{IPNet: ipNet});err != nil {
		return fmt.Errorf("VethHostName setup ip addr error: %v",err)
	}

	return nil
}


func handleExitSignal() {
	exitSignals := make(chan os.Signal, 1)
	signal.Notify(exitSignals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-exitSignals
	fmt.Println("Received signal:", sig)
	releaseResource()
	os.Exit(0)
}

func releaseResource() {
	fmt.Println("close bpf map")
	fmt.Println("del ospp_host ospp_net")
	fmt.Println("del ospp cni related host route")
}
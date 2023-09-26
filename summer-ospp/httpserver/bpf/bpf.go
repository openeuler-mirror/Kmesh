//go:generate go run github.com/cilium/ebpf/cmd/bpf2go OSPP ../../bpf/cni.ebpf.c

package bpf

import (
	"ebpf-based-cni/defaults"
	"errors"
	"fmt"
	"os"

	// "os"
	// "os/signal"
	// "syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/rlimit"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

var (
	logger *log.Logger
	f *os.File
)

func init() {
	// stopper := make(chan os.Signal, 1)
	// signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		fmt.Printf("rlimit error %v", err)
	}

	logger = log.New()
	// logger.SetFormatter()
	logger.SetLevel(log.DebugLevel)
	var err error
	f, err = os.OpenFile("/root/summer-ospp/ebpf.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Errorf("open ebpf log file err %v", err))
	}
	logger.SetOutput(f)

}

var c *ebpf.Collection

func LoadProgsAndMaps() error {
	defer f.Close()
	cs, err := LoadOSPP()
	if err != nil {
		return fmt.Errorf("load prog obj err is %w", err)
	}

	opts := ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{PinPath: defaults.GetTCGlobalsPath()},
		Programs: ebpf.ProgramOptions{
			// LogLevel: ebpf.LogLevelBranch,
		},
	}
	c, err = ebpf.NewCollectionWithOptions(cs, opts)
	if err != nil {
		var verr *ebpf.VerifierError
		if errors.As(err, &verr) {
			fmt.Printf("%+v\n", verr)
			logger.Printf("%+v\n", verr)
		}
		return fmt.Errorf("load prog and map to kernel err is %w", err)
	}
	return nil
}

func LoadTCIngressProgToVeth(vethName string) error {
	program, ok := c.Programs["tc_ingress"]
	if !ok {
		return fmt.Errorf("tc_ingress prog not found")
	}
	if err := LoadMyProgToKernelLink(program,vethName,true); err != nil {
		return fmt.Errorf("load err veth %w",err)
	}
	return nil
}

func LoadTCEgressProgToVeth(vethName string) error {
	program, ok := c.Programs["tc_egress"]
	if !ok {
		return fmt.Errorf("tc_egress prog not found")
	}
	if err := LoadMyProgToKernelLink(program,vethName,false); err != nil {
		return fmt.Errorf("load err veth %w",err)
	}
	return nil
}

func LoadTCIngressProgToHostNetDev(devName string) error {
	program, ok := c.Programs["tc_ingress_host"]
	if !ok {
		return fmt.Errorf("prog not found")
	}
	if err := LoadMyProgToKernelLink(program,devName,true); err != nil {
		return fmt.Errorf("load ebpf prog to host network dev err  %w",err)
	}
	return nil
}

func LoadMyProgToKernelLink(program *ebpf.Program,linkName string,ingress bool) error {

	link, err := netlink.LinkByName(linkName)
	if err != nil {
		return fmt.Errorf("link not found %w", err)
	}

	if err = AttchProgToTCDev(link, program, program.String(),ingress); err != nil {
		return fmt.Errorf("attach to dev tc filter failed %w", err)
	}

	return nil
}

func AttchProgToTCDev(link netlink.Link, prog *ebpf.Program, name string,ingress bool) error {

	// netlink.QdiscAdd()
	err := replaceQdisc(link)
	if err != nil {
		return fmt.Errorf("add clsact qdisc fail: %w", err)
	}
	var filter *netlink.BpfFilter
	if ingress {
		filter = &netlink.BpfFilter{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: link.Attrs().Index,
				Parent:    netlink.HANDLE_MIN_INGRESS,
				Handle:    1,
				Protocol:  unix.ETH_P_ALL,
				Priority:  1,
			},
			Fd:           prog.FD(),
			Name:         fmt.Sprintf("%s-%s", name, link.Attrs().Name),
			DirectAction: true,
		}
	}else {
		filter = &netlink.BpfFilter{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: link.Attrs().Index,
				Parent:    netlink.HANDLE_MIN_EGRESS,
				Handle:    1,
				Protocol:  unix.ETH_P_ALL,
				Priority:  1,
			},
			Fd:           prog.FD(),
			Name:         fmt.Sprintf("%s-%s", name, link.Attrs().Name),
			DirectAction: true,
		}
	}
	filter.ClassId = netlink.HANDLE_CLSACT
	if err := netlink.FilterReplace(filter); err != nil {
		return fmt.Errorf("replace tc filter : %w", err)
	}

	return nil
}

func replaceQdisc(link netlink.Link) error {
	attrs := netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(0xffff, 0),
		Parent:    netlink.HANDLE_CLSACT,
	}

	qdisc := &netlink.GenericQdisc{
		QdiscAttrs: attrs,
		QdiscType:  "clsact",
	}

	return netlink.QdiscReplace(qdisc)
}

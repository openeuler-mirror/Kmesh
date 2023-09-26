package test

import (
	"context"
	"ebpf-based-cni/defaults"
	"fmt"
	"os"
	"testing"

	gocni "github.com/containerd/go-cni"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func TestCmdAdd(t *testing.T) {

	id := "container1"
	netns := "/var/run/netns/container1"
	CreateContainer(id,netns)

	id = "container2"
	netns = "/var/run/netns/container2"
	CreateContainer(id,netns)

	id = "container3"
	netns = "/var/run/netns/container3"
	CreateContainer(id,netns)

	id = "container5"
	netns = "/var/run/netns/container5"
	CreateContainer(id,netns)
	
}

func TestCmdDel(t *testing.T) {
	id := "container1"
	netns := "/var/run/netns/container1"
	DeleteContainer(id,netns)

	id = "container2"
	netns = "/var/run/netns/container2"
	DeleteContainer(id,netns)

	id = "container3"
	netns = "/var/run/netns/container3"
	DeleteContainer(id,netns)

	id = "container5"
	netns = "/var/run/netns/container5"
	DeleteContainer(id,netns)

	ResetReservedIP()
	
}

func ResetReservedIP()  {
	var err error
	link, err := netlink.LinkByName(defaults.HostNetDev)
	if err != nil {
		fmt.Println(err)
	}
	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		fmt.Println(err)
	}
	if (addrs[0].IP.String() == "192.168.43.154") {
		str := "10.1.1.9"
		err = os.WriteFile(defaults.IPAMPATHFILE,[]byte(str),0644)
		if err != nil {
			fmt.Println(err)
		}
	}else if (addrs[0].IP.String() == "192.168.43.158") {
		str := "10.1.2.9"
		err = os.WriteFile(defaults.IPAMPATHFILE,[]byte(str),0644)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func CreateContainer(id string,netns string)  {
	// CNI allows multiple CNI configurations and the network interface
	// will be named by eth0, eth1, ..., ethN.
	ifPrefixName := "eth"
	defaultIfName := "eth0"

	// Initializes library
	l, err := gocni.New(
		// one for loopback network interface
		gocni.WithMinNetworkCount(2),
		gocni.WithPluginConfDir("/etc/cni/net.d"),
		gocni.WithPluginDir([]string{"/opt/cni/bin"}),
		// Sets the prefix for network interfaces, eth by default
		gocni.WithInterfacePrefix(ifPrefixName))
	if err != nil {
		log.Fatalf("failed to initialize cni library: %v", err)
	}

	// Load the cni configuration
	if err := l.Load(gocni.WithLoNetwork, gocni.WithDefaultConf); err != nil {
		log.Fatalf("failed to load cni configuration: %v", err)
	}

	// Setup network for namespace.
	labels := map[string]string{
		"K8S_POD_NAMESPACE":          "namespace1",
		"K8S_POD_NAME":               "pod3",
		"K8S_POD_INFRA_CONTAINER_ID": id,
		// Plugin tolerates all Args embedded by unknown labels, like
		// K8S_POD_NAMESPACE/NAME/INFRA_CONTAINER_ID...
		"IgnoreUnknown": "1",
	}

	ctx := context.Background()

	// Teardown network
	/* defer func() {
		if err := l.Remove(ctx, id, netns, gocni.WithLabels(labels)); err != nil {
			log.Fatalf("failed to teardown network: %v", err)
		}
	}() */

	// Setup network
	result, err := l.Setup(ctx, id, netns, gocni.WithLabels(labels))
	if err != nil {
		log.Fatalf("failed to setup network for namespace: %v", err)
	}

	// Get IP of the default interface
	IP := result.Interfaces[defaultIfName].IPConfigs[0].IP.String()
	fmt.Printf("IP of the default interface %s:%s", defaultIfName, IP)
}

func DeleteContainer(id string,netns string)  {
	// CNI allows multiple CNI configurations and the network interface
	// will be named by eth0, eth1, ..., ethN.
	ifPrefixName := "eth"

	// Initializes library
	l, err := gocni.New(
		// one for loopback network interface
		gocni.WithMinNetworkCount(2),
		gocni.WithPluginConfDir("/etc/cni/net.d"),
		gocni.WithPluginDir([]string{"/opt/cni/bin"}),
		// Sets the prefix for network interfaces, eth by default
		gocni.WithInterfacePrefix(ifPrefixName))
	if err != nil {
		log.Fatalf("failed to initialize cni library: %v", err)
	}

	// Load the cni configuration
	if err := l.Load(gocni.WithLoNetwork, gocni.WithDefaultConf); err != nil {
		log.Fatalf("failed to load cni configuration: %v", err)
	}

	// Setup network for namespace.
	labels := map[string]string{
		"K8S_POD_NAMESPACE":          "namespace1",
		"K8S_POD_NAME":               "pod3",
		"K8S_POD_INFRA_CONTAINER_ID": id,
		// Plugin tolerates all Args embedded by unknown labels, like
		// K8S_POD_NAMESPACE/NAME/INFRA_CONTAINER_ID...
		"IgnoreUnknown": "1",
	}

	ctx := context.Background()

	// Teardown network
	defer func() {
		if err := l.Remove(ctx, id, netns, gocni.WithLabels(labels)); err != nil {
			log.Fatalf("failed to teardown network: %v", err)
		}
	}()

	// Get IP of the default interface
	fmt.Printf("delete pod %s:%s", labels["K8S_POD_NAMESPACE"], labels["K8S_POD_NAME"])
}

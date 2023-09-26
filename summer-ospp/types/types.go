package types

import "net"

type Endpoint struct {
	
	// ID of the endpoint, unique in the scope of the node
	ID uint16 `json:"id"`

	// containerName is the name given to the endpoint by the container runtime.
	// Mutable, must be read with the endpoint lock!
	ContainerName string `json:"container_name"`

	// ifName is the name of the host facing interface (veth pair) which
	// connects into the endpoint
	IfName string `json:"if_name"`

	// ifIndex is the interface index of the host face interface (veth pair)
	IfIndex int `json:"if_index"`

	// mac is the MAC address of the endpoint
	Mac net.HardwareAddr `json:"mac"`
	
	// IPv4 is the IPv4 address of the endpoint
	IPv4 net.IP `json:"ipv4"`

	// nodeMAC is the MAC of the node (agent). The MAC is different for every endpoint.
	NodeMAC net.HardwareAddr `json:"node_mac"`
}

type BPFEndpointKey struct {
	IP uint32
}

type BPFEndpoint struct {
	IfIndex uint32
	MAC     uint64
	NodeMAC uint64
}

type BPFLBKey struct {
	IP uint32
	Slot uint32
}
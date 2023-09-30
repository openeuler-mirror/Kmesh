package endpoint

import "net"

type Endpoint struct {
	
	// ID of the endpoint, unique in the scope of the node
	ID uint16

	// containerName is the name given to the endpoint by the container runtime.
	// Mutable, must be read with the endpoint lock!
	containerName string

	// ifName is the name of the host facing interface (veth pair) which
	// connects into the endpoint
	ifName string

	// ifIndex is the interface index of the host face interface (veth pair)
	ifIndex int

	// mac is the MAC address of the endpoint
	mac net.HardwareAddr // Container MAC address.

	
	// IPv4 is the IPv4 address of the endpoint
	IPv4 net.IP

	// nodeMAC is the MAC of the node (agent). The MAC is different for every endpoint.
	nodeMAC net.HardwareAddr
}
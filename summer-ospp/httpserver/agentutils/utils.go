package agentutils

import "net"

func IpToUint32(ip net.IP) uint32 {
	ipBytes := ip.To4()
	ipUint32 := uint32(ipBytes[3])<<24 | uint32(ipBytes[2])<<16 | uint32(ipBytes[1])<<8 | uint32(ipBytes[0])
	return ipUint32
}


func MacToUint64(mac net.HardwareAddr) uint64 {
	if len(mac) != 6 {
		return 0
	}
	var macUint64 uint64
	for i := 5; i >= 0; i-- {

		macUint64 = (macUint64 << 8) | uint64(mac[i])
	}

	return macUint64
}
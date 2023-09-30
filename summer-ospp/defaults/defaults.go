package defaults

import "path/filepath"

const (
	BPFFSRoot = "/sys/fs/bpf"
	TCGlobalsPath = ""
	BaseURL = "http://127.0.0.1:3000" 
	HostNetDev = "ens33"
	IPAMPATHFILE = "/var/lib/cni/networks/mynet/last_reserved_ip.0"
)


func GetTCGlobalsPath() string{
	return filepath.Join(BPFFSRoot, TCGlobalsPath)
}

package bpf

import (
	"ebpf-based-cni/defaults"
	"ebpf-based-cni/httpserver/agentutils"
	cniTypes "ebpf-based-cni/types"
	"log"
	"net"

	"github.com/cilium/ebpf"
)

var (
	LocalPodMap *ebpf.Map
	SvcMap *ebpf.Map
)

func init()  {
	var err error
	if err := LoadProgsAndMaps();err != nil {
		log.Fatal(err)
	}
	err = LoadTCIngressProgToHostNetDev(defaults.HostNetDev)
	if err != nil {
		log.Fatal(err)
	}
	LocalPodMap, err = ebpf.LoadPinnedMap(defaults.BPFFSRoot+"/local_pod_map", &ebpf.LoadPinOptions{})
	if err != nil {
		log.Fatalf("load bpf local_pod_map faild : %v",err)
	}
	SvcMap, err = ebpf.LoadPinnedMap(defaults.BPFFSRoot+"/svc_map", &ebpf.LoadPinOptions{})
	if err != nil {
		log.Fatalf("load bpf svc_map faild : %v",err)
	}
	LoadLBInfoToMap()
}

func LoadLBInfoToMap()  {
	vip := net.ParseIP("10.1.64.64")
	ep1 := net.ParseIP("10.1.1.10")
	ep2 := net.ParseIP("10.1.1.11")
	ep3 := net.ParseIP("10.1.1.12")
	ep4 := net.ParseIP("10.1.2.10")
	ep5 := net.ParseIP("10.1.2.11")
	var err error
	lbkey := cniTypes.BPFLBKey{}
	lbkey.IP = agentutils.IpToUint32(vip)
	lbkey.Slot = 0;
	err = SvcMap.Update(lbkey,agentutils.IpToUint32(ep1),ebpf.UpdateAny);
	if err != nil {
		log.Fatalf("load lb info ep1 faild : %v",err)
	}
	lbkey.Slot = 1;
	err = SvcMap.Update(lbkey,agentutils.IpToUint32(ep2),ebpf.UpdateAny);
	if err != nil {
		log.Fatalf("load lb info ep2 faild : %v",err)
	}
	if err != nil {
		log.Fatalf("load lb info ep3 faild : %v",err)
	}
	lbkey.Slot = 2;
	err = SvcMap.Update(lbkey,agentutils.IpToUint32(ep3),ebpf.UpdateAny);
	if err != nil {
		log.Fatalf("load lb info ep4 faild : %v",err)
	}
	lbkey.Slot = 3;
	err = SvcMap.Update(lbkey,agentutils.IpToUint32(ep4),ebpf.UpdateAny);
	if err != nil {
		log.Fatalf("load lb info ep4 faild : %v",err)
	}
	lbkey.Slot = 4;
	err = SvcMap.Update(lbkey,agentutils.IpToUint32(ep5),ebpf.UpdateAny);
	if err != nil {
		log.Fatalf("load lb info ep4 faild : %v",err)
	}
}
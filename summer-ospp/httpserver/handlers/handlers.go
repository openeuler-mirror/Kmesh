package handlers

import (
	"ebpf-based-cni/httpserver/bpf"
	cniTypes "ebpf-based-cni/types"
	"log"
	"ebpf-based-cni/httpserver/agentutils"
	"github.com/cilium/ebpf"
	"github.com/gofiber/fiber/v2"
) 



func init()  {
	/* var err error
	local_pod_map, err = ebpf.LoadPinnedMap(defaults.BPFFSRoot+"/local_pod_map", &ebpf.LoadPinOptions{})
	if err != nil {
		log.Fatalf("load bpf local_pod_map faild : %v",err)
	} */
}


func CreateEP(c *fiber.Ctx) error {

	// update ep info to bpf map
	ep := new(cniTypes.Endpoint)
	err := c.BodyParser(ep)
	if err != nil {
		log.Printf("body parse err %v\n",err)
	}
	log.Println(ep.Mac.String())
	epKey := cniTypes.BPFEndpointKey{
	}
	epKey.IP = agentutils.IpToUint32(ep.IPv4)
	epValue := cniTypes.BPFEndpoint{}
	epValue.IfIndex = uint32(ep.IfIndex)
	epValue.MAC = agentutils.MacToUint64(ep.Mac)
	epValue.NodeMAC = agentutils.MacToUint64(ep.NodeMAC)
	err = bpf.LocalPodMap.Update(epKey, epValue, ebpf.UpdateAny)
	if err != nil {
		log.Fatalf("update bpf map error is :%v",err)
	}

	return nil
}

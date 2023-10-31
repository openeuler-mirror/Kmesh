package maps

// #cgo pkg-config: api-v2-c
// #include "deserialization_to_bpf_map.h"
// #include "listener/listener.pb-c.h"
import "C"
import (
	"fmt"
	"unsafe"
	
	core_v2 "openeuler.io/mesh/api/v2/core"
)

type BreakCountKeyAndValue struct {
	Address   core_v2.SocketAddress
	BreakCount int
}

func MapOfBreakCountUpdate(key *core_v2.SocketAddress, value *int) error {
	var err error

	log.Debugf("MapOfBreakCountUpdate [%s], [%d]", key.String(), *value)

	cKey, err := socketAddressToClang(key) 
	if err != nil {
		return fmt.Errorf("map of break count lookup %s", err)
	}
	defer socketAddressFreeClang(cKey)

	overCount := C.int(*value)

	ret := C.deserial_update_map_of_break_count_elem(unsafe.Pointer(cKey), unsafe.Pointer(&overCount))
	if ret != 0 {
		return fmt.Errorf("map of break count update deserial_update_map_of_break_count_elem failed")
	}
	return nil
}

func MapOfBreakCountDelete(key *core_v2.SocketAddress) error {
	log.Debugf("MapOfBreakCountDelete [%s]", key.String())

	cKey, err := socketAddressToClang(key)
	if err != nil {
		return fmt.Errorf("MapOfBreakCountLookup %s", err)
	}
	defer socketAddressFreeClang(cKey)
	ret := C.deserial_delete_map_of_break_count_elem(unsafe.Pointer(cKey))
	if ret != 0 {
		return fmt.Errorf("map of break count update deserial_delete_map_of_break_count_elem failed")
	}
	return nil
}

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

type IsBreakKeyAndValue struct {
	BreakAddress core_v2.SocketAddress
	BreakValue   int
}

func MapOfIsBreakUpdate(key *core_v2.SocketAddress, value *int) error {
	var err error

	log.Debugf("MapOfIsBreakUpdate [%v], [%d]", *key, *value)

	cKeyPtr, err := socketAddressToClang(key)
	if err != nil {
		return fmt.Errorf("MapOfIsBreakLookup %s", err)
	}
	defer socketAddressFreeClang(cKeyPtr)

	breakValue := C.int(*value)

	ret := C.deserial_update_map_of_is_break_elem(unsafe.Pointer(cKeyPtr), unsafe.Pointer(&breakValue))
	if ret != 0 {
		return fmt.Errorf("MapOfIsBreakUpdate deserial_update_map_of_is_break_elem failed")
	}
	return nil
}

func MapOfIsBreakDelete(key *core_v2.SocketAddress) error {
	log.Debugf("MapOfIsBreakDelete [%v]", *key)

	cKeyPtr, err := socketAddressToClang(key)
	if err != nil {
		return fmt.Errorf("MapOfIsBreakLookup %s", err)
	}
	defer socketAddressFreeClang(cKeyPtr)

	ret := C.deserial_delete_map_of_is_break_elem(unsafe.Pointer(cKeyPtr))
	if ret != 0 {
		return fmt.Errorf("MapOfIsBreakDelete deserial_delete_map_of_is_break_elem failed")
	}
	return nil
}

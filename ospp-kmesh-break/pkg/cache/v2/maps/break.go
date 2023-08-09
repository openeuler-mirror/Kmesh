package maps

// #cgo pkg-config: api-v2-c
// #include "deserialization_to_bpf_map.h"
// #include "listener/listener.pb-c.h"
import "C"
import (
	"fmt"
	"unsafe"

)

const MaxKeyLength = 50

type BreakKeyAndValue struct {
	BreakKey   [MaxKeyLength]byte
	BreakValue int
}

func MapOfBreakUpdate(key string, value *int) error {
	if len(key) > MaxKeyLength {
		return fmt.Errorf("key length exceeds maximum allowed length")
	}

	var err error

	log.Debugf("map of break update [%s], [%d]", key, *value)

	// Convert Go string to [MaxKeyLength]byte
	var cKey [MaxKeyLength]byte
	copy(cKey[:], key)

	cKeyPtr, err := keyToClang(&cKey)
	if err != nil {
		return fmt.Errorf("map of break lookup %s", err)
	}
	defer keyFreeClang(cKeyPtr)

	// Assuming value is directly compatible with C.int
	breakValue := C.int(*value)

	ret := C.deserial_update_map_of_break_elem(unsafe.Pointer(cKeyPtr), unsafe.Pointer(&breakValue))
	if ret != 0 {
		return fmt.Errorf("map of break update deserial_update_map_of_break_elem failed")
	}
	return nil
}

func MapOfBreakDelete(key string) error {
	if len(key) > MaxKeyLength {
		return fmt.Errorf("key length exceeds maximum allowed length")
	}

	log.Debugf("map of break delete [%s]", key)

	// Convert Go string to [MaxKeyLength]byte
	var cKey [MaxKeyLength]byte
	copy(cKey[:], key)

	cKeyPtr, err := keyToClang(&cKey)
	if err != nil {
		return fmt.Errorf("map of break lookup %s", err)
	}
	defer keyFreeClang(cKeyPtr)

	ret := C.deserial_delete_map_of_break_elem(unsafe.Pointer(cKeyPtr))
	if ret != 0 {
		return fmt.Errorf("map of break update deserial_delete_map_of_break_elem failed")
	}
	return nil
}

package nvml

/*
#include "nvmlbridge.h"
*/
import "C"

import (
	"testing"
)

type TestDevice struct {
	Device
}

var intproptestfunctions = map[string]C.uint{
	"Index": C.uint(0),
}

func (gpu *TestDevice) intProperty(property string) (uint, error) {
	var t *testing.T
	// t.Logf("entering mocked intProperty")
	ret, ok := intproptestfunctions[property]
	if ok == false {
		t.Errorf("Could not find table entry for %s", property)
	}

	return uint(ret), nil
}

func testIndex(t *testing.T) {
	var gpu TestDevice
	idx, err := gpu.intProperty("Index")
	if err != nil {
		t.Errorf("gpu.Index() returned error: %s idx: %d", err, idx)
	}
}

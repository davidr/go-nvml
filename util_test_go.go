package nvml

/*
#include <string.h>
*/
import "C"

import (
	"testing"
)

// TestCTextBufferHandling tests some rudimentary functionality we wrote to
// handle passing text back and forth between Go and C.
func testCStringHandling(t *testing.T) {
	var tests = []struct {
		cs *C.char
		sz uint
		gs string
	}{
		{C.CString("test"), 4, "test"},
		{C.CString("testalongerstring"), 4, "test"},
		{C.CString("testalongerstring"), 100, "testalongerstring"},
	}

	for _, ts := range tests {
		cbuf := genCStringBuffer(ts.sz + 1)
		C.strncpy(cbuf, ts.cs, (C.size_t)(ts.sz))

		gs := strndup(cbuf, ts.sz)
		if gs != ts.gs {
			t.Errorf("converted cstring %s != strndup returned %s", ts.gs, gs)
		}

	}

}

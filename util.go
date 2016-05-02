package nvml

/*
#include "nvmlbridge.h"
*/
import "C"

import (
	"errors"
)

// NVMLInit initializes the NVML session.
func NVMLInit() error {
	var result C.nvmlReturn_t

	result = C.nvmlInit()
	if result != C.NVML_SUCCESS {
		return errors.New("nvmlInit returned error")
	}

	return nil
}

// lots of the nvml functions require an allocated *char into which to place
// strings. genCStringBuffer() allocates this buffer and returns it.
//
// IMPORTANT: These buffers need to be freed! It is strongly advised to put:
//
//            defer C.free(unsafe.Pointer(buffer))
//
// immediately after the allocation of this buffer!
//
func genCStringBuffer(size uint) *C.char {
	buf := make([]byte, size)
	return C.CString(string(buf))
}

// strndup replicates the functionality of strndup from string.h in go to
// convert *C.char to string, respecting null-termination in the original.
// C.GoStringN doesn't stop at null characters.
//
// h/t: https://utcc.utoronto.ca/~cks/space/blog/programming/GoCGoStringFunctions
//
func strndup(cs *C.char, len uint) string {
	return C.GoStringN(cs, C.int(C.strnlen(cs, C.size_t(len))))
}

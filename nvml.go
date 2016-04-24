package nvml

/*
// #cgo CPPFLAGS: -I/path/to/install
// #cgo LDFLAGS: -L/path/to/install -l nvidia-ml

#include <nvml.h>
#include <stdlib.h>
#include <stddef.h>
#include <string.h>


// Not every function can be genericized in this way because of all the custom structs,
// but there are several nvmlGet functions we want that take a nvmlDevice_t, *char, and
// a length as arguments. These are trivial to pass as function pointers along with their,
// arguments, so we might as well save some effort.
//
typedef int (*gettextProperty) (nvmlDevice_t device , char *buf, uint length);
int bridge_get_text_property(gettextProperty f,
                             nvmlDevice_t device,
                             char *buf,
                             uint length)
{
    nvmlReturn_t ret;

    ret = f(device, buf, length);

    if (ret == NVML_SUCCESS) {
        return(EXIT_SUCCESS);
    } else {
        return(EXIT_FAILURE);
    }
}

*/
import "C"

import (
	"errors"
	"log"
	"unsafe"
)

type Device struct {
	nvmldevice C.nvmlDevice_t
	pcibus     string
	name       string
	uuid       string
}

func NewDevice(cdevice C.nvmlDevice_t) (*Device, error) {

	device := Device{
		nvmldevice: cdevice,
	}

	// Populate the device with some basic properties
	uuid, err := device.UUID()
	if err != nil {
		return nil, errors.New("Cannot retrieve property")
	}
	device.uuid = uuid

	name, err := device.Name()
	if err != nil {
		return nil, errors.New("Cannot retrieve property")
	}
	device.name = name

	return &device, nil
}

func (gpu *Device) PowerState() (int, error) {
	var pstate C.nvmlPstates_t
	var result C.nvmlReturn_t

	result = C.nvmlDeviceGetPowerState(gpu.nvmldevice, &pstate)
	if result != C.NVML_SUCCESS {
		return -1, errors.New("GetPowerState returned error")
	}

	return int(pstate), nil
}

func (gpu *Device) PowerUsage() (uint, error) {
	var result C.nvmlReturn_t
	var cusage C.uint

	result = C.nvmlDeviceGetPowerUsage(gpu.nvmldevice, &cusage)
	if result != C.NVML_SUCCESS {
		return 0, errors.New("GetPowerState returned error")
	}

	return uint(cusage), nil
}

func (gpu *Device) Temp() (uint, error) {
	var result C.nvmlReturn_t
	var ctemp C.uint

	result = C.nvmlDeviceGetTemperature(gpu.nvmldevice, C.NVML_TEMPERATURE_GPU, &ctemp)
	if result != C.NVML_SUCCESS {
		return 0, errors.New("GetPowerState returned error")
	}

	return uint(ctemp), nil
}

type CTextPropFunc struct {
	f      C.gettextProperty
	length C.uint
}

var textpropfunctions = map[string]*CTextPropFunc{
	"Name":                {C.gettextProperty(C.nvmlDeviceGetName), C.NVML_DEVICE_NAME_BUFFER_SIZE},
	"Serial":              {C.gettextProperty(C.nvmlDeviceGetSerial), C.NVML_DEVICE_SERIAL_BUFFER_SIZE},
	"UUID":                {C.gettextProperty(C.nvmlDeviceGetUUID), C.NVML_DEVICE_UUID_BUFFER_SIZE},
	"InforomImageVersion": {C.gettextProperty(C.nvmlDeviceGetInforomImageVersion), C.NVML_DEVICE_INFOROM_VERSION_BUFFER_SIZE},
	"VbiosVersion":        {C.gettextProperty(C.nvmlDeviceGetVbiosVersion), C.NVML_DEVICE_VBIOS_VERSION_BUFFER_SIZE},
}

func (gpu *Device) InforomImageVersion() (string, error) {
	return gpu.textProperty("InforomImageVersion")
}

func (gpu *Device) VbiosVersion() (string, error) {
	return gpu.textProperty("VbiosVersion")
}

// Return the product name of the device, e.g. "Tesla K40m"
func (gpu *Device) Name() (string, error) {
	return gpu.textProperty("Name")
}

// Return the UUID of the device
func (gpu *Device) UUID() (string, error) {
	return gpu.textProperty("UUID")
}

// Return the serial number of the device
func (gpu *Device) Serial() (string, error) {
	return gpu.textProperty("Serial")
}

// textProperty takes a propertyname as input and then runs the corresponding
// function in the textpropfunctions map, returning the result as a Go string.
//
// textProperty takes care of allocating (and freeing) the text buffers of
// proper size.
func (gpu *Device) textProperty(property string) (string, error) {
	var result C.int
	var propvalue string

	tpf, ok := textpropfunctions[property]
	if ok == false {
		return "", errors.New("property not found")
	}

	var buf *C.char = genCStringBuffer(uint(tpf.length))
	defer C.free(unsafe.Pointer(buf))

	result = C.bridge_get_text_property(tpf.f, gpu.nvmldevice, buf, tpf.length)

	if result != C.EXIT_SUCCESS {
		return propvalue, errors.New("gettextProperty bridge returned error")
	}

	propvalue = strndup(buf, int(tpf.length))
	if len(propvalue) > 0 {
		return propvalue, nil
	} else {
		return "", errors.New("textProperty returned empty string")
	}
}

// Go correspondent of the C.nvmlMemory_t struct. Memory in bytes
type NvmlMemory struct {
	free  uint64
	total uint64
	used  uint64
}

// MemoryInfo returns a NvmlMemory struct populated with the amount of memory used,
// free, and in total on the device, in bytes.
func (gpu *Device) MemoryInfo() (NvmlMemory, error) {
	var result C.nvmlReturn_t
	var cmeminfo C.nvmlMemory_t
	var meminfo NvmlMemory

	result = C.nvmlDeviceGetMemoryInfo(gpu.nvmldevice, &cmeminfo)
	if result != C.NVML_SUCCESS {
		return meminfo, errors.New("GetPowerState returned error")
	}

	meminfo.free = uint64(cmeminfo.free)
	meminfo.total = uint64(cmeminfo.total)
	meminfo.used = uint64(cmeminfo.used)

	return meminfo, nil
}

// Return a string representation of the nvml
func (gpu *Device) Error(cerror C.nvmlReturn_t) error {
	var cerrorstring *C.char

	// No need to process anything further if the nvml call succeeded
	if cerror == C.NVML_SUCCESS {
		return nil
	}

	cerrorstring = C.nvmlErrorString(cerror)
	if cerrorstring == nil {
		// I'm not sure how this could happen, but it's easy to check for
		return errors.New("Error not found in nvml.h")
	}

	return errors.New(C.GoString(cerrorstring))
}

func nvmlInit() error {
	var result C.nvmlReturn_t

	result = C.nvmlInit()
	if result != C.NVML_SUCCESS {
		return errors.New("nvmlInit returned error")
	}

	return nil
}

func nvmlDeviceGetCount() (int, error) {
	var count C.uint

	result := C.nvmlDeviceGetCount(&count)
	if result != C.NVML_SUCCESS {
		return -1, errors.New("nvmlDeviceGetCount failed")
	}

	return int(count), nil
}

func GetAllGPUs() ([]Device, error) {
	var devices []Device
	cdevices, err := getAllDevices()
	if err != nil {
		return devices, err
	}

	for _, cdevice := range cdevices {
		device, err := NewDevice(cdevice)
		if err != nil {
			break
		}

		devices = append(devices, *device)
	}

	return devices, nil
}

// getAllDevices returns an array of nvmlDevice_t structs representing all GPU
// devices in the system.
func getAllDevices() ([]C.nvmlDevice_t, error) {
	var devices []C.nvmlDevice_t

	device_count, err := nvmlDeviceGetCount()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < device_count; i++ {
		var device C.nvmlDevice_t
		result := C.nvmlDeviceGetHandleByIndex(C.uint(i), &device)
		if result != C.NVML_SUCCESS {
			return devices, errors.New("nvmlDeviceGetHandleByIndex returns error")
		}

		devices = append(devices, device)
	}

	if len(devices) > 0 {
		return devices, errors.New("No devices found")
	}

	return devices, nil
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
func strndup(cstring *C.char, length int) string {
	clength := C.int(C.strnlen(cstring, C.size_t(length)))
	gostring := C.GoStringN(cstring, clength)
	return gostring
}

func init() {
	err := nvmlInit()
	if err != nil {
		log.Fatal("Could not initialize NVML interface: %s\n", err)
	}
}

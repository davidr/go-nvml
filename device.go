package nvml

/*
#cgo CPPFLAGS: -I/usr/local/cuda/include -I/usr/local/cuda-8.0/targets/x86_64-linux/include
#cgo LDFLAGS: -L/usr/lib/nvidia-375 -L/usr/lib/nvidia -l nvidia-ml

#include "nvmlbridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"log"
	"unsafe"
)

type Device struct {
	nvmldevice C.nvmlDevice_t
	index      uint
	pcibus     string
	name       string
	uuid       string
}

// NewDevice is a contstructor function for Device structs. Given an nvmlDevice_t
// object as input, it populates some static property fields and returns a Device
func NewDevice(cdevice C.nvmlDevice_t) (*Device, error) {
	device := Device{
		nvmldevice: cdevice,
	}

	uuid, err := device.UUID()
	if err != nil {
		return nil, errors.New("Cannot retrieve UUID property")
	}
	device.uuid = uuid

	name, err := device.Name()
	if err != nil {
		return nil, errors.New("Cannot retrieve Name property")
	}
	device.name = name

	index, err := device.Index()
	if err != nil {
		return nil, errors.New("Cannot retrieve Index property")
	}
	device.index = index

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

// Temp returns the current temperature of the card in degrees Celsius
func (gpu *Device) Temp() (uint, error) {
	var result C.nvmlReturn_t
	var ctemp C.uint

	result = C.nvmlDeviceGetTemperature(gpu.nvmldevice, C.NVML_TEMPERATURE_GPU, &ctemp)
	if result != C.NVML_SUCCESS {
		return 0, errors.New("GetPowerState returned error")
	}

	return uint(ctemp), nil
}

const (
	NVML_CLOCK_GRAPHICS = iota
	NVML_CLOCK_SM
	NVML_CLOCK_MEM
	NVML_CLOCK_COUNT
)

// ClockInfo returns the {graphics, shader model, or memory} clock in Mhz.
// Pass in the relevant constant NVML_CLOCK_{GRAPHICS,SM,MEM}.
func (gpu *Device) ClockInfo(cclock_type C.nvmlClockType_t) (uint, error) {
	var result C.nvmlReturn_t
	var cclock C.uint

	result = C.nvmlDeviceGetClockInfo(gpu.nvmldevice, cclock_type, &cclock)
	if result != C.NVML_SUCCESS {
		return 0, errors.New("GetClockInfo returned error")
	}

	return uint(cclock), nil
}

// MaxClockInfo returns the {graphics, shader model, or memory} clock in Mhz.
// Pass in the relevant constant NVML_CLOCK_{GRAPHICS,SM,MEM}.
func (gpu *Device) MaxClockInfo(cclock_type C.nvmlClockType_t) (uint, error) {
	var result C.nvmlReturn_t
	var cclock C.uint

	result = C.nvmlDeviceGetMaxClockInfo(gpu.nvmldevice, cclock_type, &cclock)
	if result != C.NVML_SUCCESS {
		return 0, errors.New("GetMaxClockInfo returned error")
	}

	return uint(cclock), nil
}

type cIntPropFunc struct {
	f C.getintProperty
}

var intpropfunctions = map[string]*cIntPropFunc{
	"Index":                        {C.getintProperty(C.nvmlDeviceGetIndex)},
	"MinorNumber":                  {C.getintProperty(C.nvmlDeviceGetMinorNumber)},
	"InforomConfigurationChecksum": {C.getintProperty(C.nvmlDeviceGetInforomConfigurationChecksum)},
	"MaxPCIeLinkGeneration":        {C.getintProperty(C.nvmlDeviceGetMaxPcieLinkGeneration)},
	"MaxPCIeLinkWidth":             {C.getintProperty(C.nvmlDeviceGetMaxPcieLinkWidth)},
	"CurrPCIeLinkGeneration":       {C.getintProperty(C.nvmlDeviceGetCurrPcieLinkGeneration)},
	"CurrPCIeLinkWidth":            {C.getintProperty(C.nvmlDeviceGetCurrPcieLinkWidth)},
	"PCIeReplayCounter":            {C.getintProperty(C.nvmlDeviceGetPcieReplayCounter)},
	"FanSpeed":                     {C.getintProperty(C.nvmlDeviceGetFanSpeed)},
	"PowerManagementLimit":         {C.getintProperty(C.nvmlDeviceGetPowerManagementLimit)},
	"PowerManagementDefaultLimit":  {C.getintProperty(C.nvmlDeviceGetPowerManagementDefaultLimit)},
	"PowerUsage":                   {C.getintProperty(C.nvmlDeviceGetPowerUsage)},
	"EnforcedPowerLimit":           {C.getintProperty(C.nvmlDeviceGetEnforcedPowerLimit)},
	"BoardId":                      {C.getintProperty(C.nvmlDeviceGetBoardId)},
	"MultiGpuBoard":                {C.getintProperty(C.nvmlDeviceGetMultiGpuBoard)},
}

func (gpu *Device) intProperty(property string) (uint, error) {
	var cuintproperty C.uint

	ipf, ok := intpropfunctions[property]
	if ok == false {
		return 0, errors.New("property not found")
	}

	result := C.bridge_get_int_property(ipf.f, gpu.nvmldevice, &cuintproperty)
	if result != C.EXIT_SUCCESS {
		return 0, errors.New("getintProperty bridge returned error")
	}

	return uint(cuintproperty), nil
}

// Index returns the NVML index of the device.
func (gpu *Device) Index() (uint, error) {
	return gpu.intProperty("Index")
}

// MinorNumber returns the minor number of the device. The minor number
// is the integer such that the device node file for the GPU will be
// /dev/nvidia[Device.MinorNumber]
func (gpu *Device) MinorNumber() (uint, error) {
	return gpu.intProperty("MinorNumber")
}

// InforomConfigurationChecksum returns the checksum of the configuration
// stored in the device's inforom. (Can be used to verify identical configuration
// between devices.)
func (gpu *Device) InforomConfigurationChecksum() (uint, error) {
	return gpu.intProperty("InforomConfigurationChecksum")
}

// MaxPCIeLinkGeneration returns the maximum PCIe link generation possible with this
// device and system.
func (gpu *Device) MaxPCIeLinkGeneration() (uint, error) {
	return gpu.intProperty("MaxPCIeLinkGeneration")
}

// MaxPCIeLinkWidth returns the maximum PCIe link width possible with this device
// and system
func (gpu *Device) MaxPCIeLinkWidth() (uint, error) {
	return gpu.intProperty("MaxPCIeLinkWidth")
}

// CurrPCIeLinkGeneration returns the current PCIe link generation number
func (gpu *Device) CurrPCIeLinkGeneration() (uint, error) {
	return gpu.intProperty("CurrPCIeLinkGeneration")
}

// CurrPCIeLinkWidth returns the current PCIe link width
func (gpu *Device) CurrPCIeLinkWidth() (uint, error) {
	return gpu.intProperty("CurrPCIeLinkWidth")
}

// PCIeReplayCounter returns the replay counter and rollover info.
func (gpu *Device) PCIeReplayCounter() (uint, error) {
	return gpu.intProperty("PCIeReplayCounter")
}

// FanSpeed returns the current fan speed of the device, on devices that
// have fans.
func (gpu *Device) FanSpeed() (uint, error) {
	return gpu.intProperty("FanSpeed")
}

// PowerManagementLimit returns the power management limit for the device, in mW
func (gpu *Device) PowerManagementLimit() (uint, error) {
	return gpu.intProperty("PowerManagementLimit")
}

// PowerManagementDefaultLimit returns the upper limit for the amount of power
// the card is allowed to draw, in mW.
func (gpu *Device) PowerManagementDefaultLimit() (uint, error) {
	return gpu.intProperty("PowerManagementDefaultLimit")
}

// PowerUsage returns the current power usage of the device, in mW.
func (gpu *Device) PowerUsage() (uint, error) {
	return gpu.intProperty("PowerUsage")
}

// EnforcedPowerLimit returns the effective power limit that the driver enforces after
// taking into account all limiters.
func (gpu *Device) EnforcedPowerLimit() (uint, error) {
	return gpu.intProperty("EnforcedPowerLimit")
}

// BoardID returns the device boardId, which will be identical for GPUs connected to
// the same PLX
func (gpu *Device) BoardId() (uint, error) {
	return gpu.intProperty("BoardId")
}

// MultiGpuBoard
func (gpu *Device) MultiGpuBoard() (bool, error) {
	p, err := gpu.intProperty("MultiGpuBoard")
	if int(p) == 0 {
		return true, err
	} else {
		return false, err
	}
}

type cTextPropFunc struct {
	f      C.gettextProperty
	length C.uint
}

var textpropfunctions = map[string]*cTextPropFunc{
	"Name":                {C.gettextProperty(C.nvmlDeviceGetName), C.NVML_DEVICE_NAME_BUFFER_SIZE},
	"Serial":              {C.gettextProperty(C.nvmlDeviceGetSerial), C.NVML_DEVICE_SERIAL_BUFFER_SIZE},
	"UUID":                {C.gettextProperty(C.nvmlDeviceGetUUID), C.NVML_DEVICE_UUID_BUFFER_SIZE},
	"InforomImageVersion": {C.gettextProperty(C.nvmlDeviceGetInforomImageVersion), C.NVML_DEVICE_INFOROM_VERSION_BUFFER_SIZE},
	"VbiosVersion":        {C.gettextProperty(C.nvmlDeviceGetVbiosVersion), C.NVML_DEVICE_VBIOS_VERSION_BUFFER_SIZE},
}

// textProperty takes a propertyname as input and then runs the corresponding
// function in the textpropfunctions map, returning the result as a Go string.
//
// textProperty takes care of allocating (and freeing) the text buffers of
// proper size.
func (gpu *Device) textProperty(property string) (string, error) {
	var propvalue string

	// If there isn't a valid entry for this property in the map, there's no reason
	// to process any further
	tpf, ok := textpropfunctions[property]
	if ok == false {
		return "", errors.New("property not found")
	}

	var buf *C.char = genCStringBuffer(uint(tpf.length))
	defer C.free(unsafe.Pointer(buf))

	result := C.bridge_get_text_property(tpf.f, gpu.nvmldevice, buf, tpf.length)
	if result != C.EXIT_SUCCESS {
		return propvalue, errors.New("gettextProperty bridge returned error")
	}

	propvalue = strndup(buf, uint(tpf.length))
	if len(propvalue) > 0 {
		return propvalue, nil
	} else {
		return "", errors.New("textProperty returned empty string")
	}
}

// InforomImageVersion returns the global inforom image version
func (gpu *Device) InforomImageVersion() (string, error) {
	return gpu.textProperty("InforomImageVersion")
}

// VbiosVersion returns the VBIOS version of the device
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

// Go correspondent of the C.nvmlMemory_t struct. Memory in bytes
type NVMLMemory struct {
	Free  uint64
	Total uint64
	Used  uint64
}

// MemoryInfo returns a NVMLMemory struct populated with the amount of memory used,
// free, and in total on the device, in bytes.
func (gpu *Device) MemoryInfo() (NVMLMemory, error) {
	var result C.nvmlReturn_t
	var cmeminfo C.nvmlMemory_t
	var meminfo NVMLMemory

	result = C.nvmlDeviceGetMemoryInfo(gpu.nvmldevice, &cmeminfo)
	if result != C.NVML_SUCCESS {
		return meminfo, errors.New("GetMemoryInfo returned error")
	}

	meminfo.Free = uint64(cmeminfo.free)
	meminfo.Total = uint64(cmeminfo.total)
	meminfo.Used = uint64(cmeminfo.used)

	return meminfo, nil
}

// Go correspondent of the C.nvmlUtilization_t struct. gpu/memory in percent
type NVMLUtilization struct {
	Gpu    uint
	Memory uint
}

// UtilizationRates returns a NVMLUtilization struct with gpu/memory
// utilization in percent
func (gpu *Device) UtilizationRates() (NVMLUtilization, error) {
	var result C.nvmlReturn_t
	var cutil C.nvmlUtilization_t
	var util NVMLUtilization

	result = C.nvmlDeviceGetUtilizationRates(gpu.nvmldevice, &cutil)
	if result != C.NVML_SUCCESS {
		return util, errors.New("GetUtilizationRates returned error")
	}

	util.Gpu = uint(cutil.gpu)
	util.Memory = uint(cutil.memory)

	return util, nil
}

type NVMLProcessInfo struct {
	Pid           uint
	UsedGpuMemory uint64
}

// GraphicsRunningProcesses only tells you if there is a process using the GPU
// in that case it will return (nil, nil), otherwise you'll get an error
func (gpu *Device) GraphicsRunningProcesses() ([]NVMLProcessInfo, error) {
	var result C.nvmlReturn_t
	var cinfoCount C.uint = 0

	result = C.nvmlDeviceGetGraphicsRunningProcesses(gpu.nvmldevice, &cinfoCount, nil)
	if result != C.NVML_SUCCESS {
		if result == C.NVML_ERROR_INSUFFICIENT_SIZE {
			return nil, errors.New("GetGraphicsRunningProcesses insufficient size")
		} else {
			return nil, errors.New("GetGraphicsRunningProcesses returned error")
		}
	}

	return nil, nil
}

// EccErrors will return an error if there have been memory errors on the hardware
func (gpu *Device) EccErrors() (err error) {
	var result C.nvmlReturn_t
	var cCount C.ulonglong = 0

	result = C.nvmlDeviceGetTotalEccErrors(gpu.nvmldevice, C.NVML_MEMORY_ERROR_TYPE_UNCORRECTED, C.NVML_AGGREGATE_ECC, &cCount)
	if result != C.NVML_SUCCESS {
		switch result {
		case C.NVML_ERROR_NOT_SUPPORTED:
			return fmt.Errorf("nvmlDeviceGetMemoryErrorCounter is not supported on this hardware")
		default:
			return fmt.Errorf("nvmlDeviceGetMemoryErrorCounter returned error (%d)", result)
		}
	}
	if cCount != 0 {
		return fmt.Errorf("nvmlDeviceGetMemoryErrorCounter detected errors (%d)", cCount)
	}

	return nil
}

// Return a proper golang error of representation of the nvmlReturn_t error
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

func nvmlDeviceGetCount() (int, error) {
	var count C.uint

	result := C.nvmlDeviceGetCount(&count)
	if result != C.NVML_SUCCESS {
		return -1, errors.New("nvmlDeviceGetCount failed")
	}

	return int(count), nil
}

// GetAllGPUs will return a slice of type Device for all NVML devices present on
// the host system
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

	if len(devices) == 0 {
		return devices, errors.New("No devices found")
	}

	return devices, nil
}

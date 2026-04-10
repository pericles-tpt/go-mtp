package main

// #cgo pkg-config: libmtp
// #include <libmtp.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type RawDevice C.LIBMTP_raw_device_t
type MTPDevice C.LIBMTP_mtpdevice_t
type File C.LIBMTP_file_t

type RawDeviceList []*RawDevice

func ReleaseDevice(m *MTPDevice) {
	C.LIBMTP_Release_Device((*C.LIBMTP_mtpdevice_t)(m))
}

func GetRawDevices() (RawDeviceList, error) {
	var devs *C.LIBMTP_raw_device_t
	var numDevs C.int
	r := C.LIBMTP_Detect_Raw_Devices(&devs, &numDevs)
	if r != 0 {
		fmt.Println("GetRawDevices might be an error?: ", r)
	}

	slice := unsafe.Slice(&devs, int(numDevs))
	rdl := make(RawDeviceList, numDevs)
	for i, d := range slice {
		var nd *RawDevice = nil
		if d != nil {
			nd = (*RawDevice)(d)
		}
		rdl[i] = nd
	}
	return rdl, nil
}

// TODO:
// func FreeRawDevices(dl RawDeviceList) {
// 	C.free(dl)
// }
// TODO: Release Device

func OpenRawDevice(d *RawDevice) *MTPDevice {
	fmt.Println("d: ", d)
	mtpDevice := C.LIBMTP_Open_Raw_Device_Uncached((*C.LIBMTP_raw_device_t)(d))
	fmt.Println("mtp device: ", mtpDevice)
	return (*MTPDevice)(mtpDevice)
}

func GetStorage(md *MTPDevice) {
	// 0 -> not sorted, maybe add sorting in future
	err := C.LIBMTP_Get_Storage((*C.LIBMTP_mtpdevice_t)(md), 0)
	// TODO: Handle error
	if err != 0 {
		fmt.Println("Maybe error?: ", err)
	}
}

func GetFilesAndFolders(md *MTPDevice, storage, parent uint32) ([]*File, error) {
	var (
		ret = []*File{}
	)
	if md == nil {
		return ret, errors.New("nil device provided")
	}

	fl := C.LIBMTP_Get_Files_And_Folders((*C.LIBMTP_mtpdevice_t)(md), C.uint32_t(storage), C.uint32_t(parent))
	if fl == nil {
		return ret, errors.New("no files found")
	}
	defer C.LIBMTP_destroy_file_t(fl)

	for p := fl; p != nil; p = p.next {
		ret = append(ret, (*File)(p))
	}
	return ret, nil
}

func (m *MTPDevice) GetSerialNumber() string {
	sno := C.LIBMTP_Get_Serialnumber((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(sno)
}

func (m *MTPDevice) GetManufacturerName() string {
	sno := C.LIBMTP_Get_Manufacturername((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(sno)
}

func (m *MTPDevice) GetModelName() string {
	sno := C.LIBMTP_Get_Modelname((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(sno)
}

func (m *MTPDevice) GetDeviceVersion() string {
	sno := C.LIBMTP_Get_Deviceversion((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(sno)
}

func (m *MTPDevice) GetFriendlyName() string {
	sno := C.LIBMTP_Get_Friendlyname((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(sno)
}

func (m *MTPDevice) GetBatteryLevel() (uint8, uint8, error) {
	var (
		max  = C.uint8_t(0)
		curr = C.uint8_t(0)
	)
	errno := C.LIBMTP_Get_Batterylevel((*C.LIBMTP_mtpdevice_t)(m), &max, &curr)
	if errno != 0 {
		return 0, 0, fmt.Errorf("failed got error code: %d, battery level likely unsupported", errno)
	}
	return uint8(max), uint8(curr), nil
}

// GetDeviceBySerialNumber, returns a CACHED MTPDevice
func GetDeviceBySerialNumber(sno string) *MTPDevice {
	cstr := C.CString(sno)
	dev := C.LIBMTP_Get_Device_By_SerialNumber(cstr)
	defer C.free(unsafe.Pointer(cstr))
	return (*MTPDevice)(dev)
}

func main() {
	dl, err := GetRawDevices()
	if err != nil {
		panic(err)
	}
	// defer FreeRawDevices(dl)

	if len(dl) > 0 {
		mtpDevice := OpenRawDevice(dl[0])
		fmt.Println("mtp device: ", mtpDevice)
		if mtpDevice == nil {
			return
		}
		defer ReleaseDevice(mtpDevice)

		sno := mtpDevice.GetSerialNumber()
		fmt.Println("Sno: ", sno)
		fmt.Println("Manu: ", mtpDevice.GetManufacturerName())
		fmt.Println("ModName: ", mtpDevice.GetModelName())
		fmt.Println("DevVer: ", mtpDevice.GetDeviceVersion())
		fmt.Println("Friendly: ", mtpDevice.GetFriendlyName())
		max, curr, err := mtpDevice.GetBatteryLevel()
		if err != nil {
			fmt.Println("Failed to get battery level: ", err)
		}
		fmt.Println(curr, max)

		// mtpDevice = GetDeviceBySerialNumber(sno)
		// defer ReleaseDevice(mtpDevice)
		// fmt.Println("Should be the same as the first mtp device: ", mtpDevice)

		// GetStorage(mtpDevice)
		// sid := uint32(mtpDevice.storage.id)

		// files, err := GetFilesAndFolders(mtpDevice, sid, 0)
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println("Files are: ", files)

		// for _, f := range files {
		// 	s := C.GoString(f.filename)
		// 	fmt.Println("fn: ", s)
		// }

	}
}

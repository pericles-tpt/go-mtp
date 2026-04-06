package main

// #cgo pkg-config: libmtp
// #include <libmtp.h>
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

func Init() {
	C.LIBMTP_Init()
}

func ReleaseMTPDevice(d *MTPDevice) {
	C.LIBMTP_Release_Device((*C.LIBMTP_mtpdevice_t)(d))
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

func main() {
	Init()
	dl, err := GetRawDevices()
	if err != nil {
		panic(err)
	}
	// defer FreeRawDevices(dl)

	if len(dl) > 0 {
		mtpDevice := OpenRawDevice(dl[0])
		fmt.Println("mtp device: ", mtpDevice)
		if mtpDevice != nil {
			defer ReleaseMTPDevice(mtpDevice)
		}

		GetStorage(mtpDevice)
		sid := uint32(mtpDevice.storage.id)

		files, err := GetFilesAndFolders(mtpDevice, sid, 0)
		if err != nil {
			panic(err)
		}
		fmt.Println("Files are: ", files)

		for _, f := range files {
			s := C.GoString(f.filename)
			fmt.Println("fn: ", s)
		}

	}
}

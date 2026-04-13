package gomtp

// #cgo pkg-config: libmtp
// #include <libmtp.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"time"
	"unsafe"
)

type RawDevice C.LIBMTP_raw_device_t
type MTPDevice C.LIBMTP_mtpdevice_t

type File struct {
	ItemId    uint32
	ParentId  uint32
	StorageId uint32

	Filename         string
	Filesize         uint64
	ModificationDate time.Time

	// TODO: Implement go enum for FileType
	FileType C.LIBMTP_filetype_t
}

type DeviceCapability C.LIBMTP_devicecap_t

var (
	GetPartialObject  DeviceCapability = C.LIBMTP_DEVICECAP_GetPartialObject
	SendPartialObject DeviceCapability = C.LIBMTP_DEVICECAP_SendPartialObject
	EditObject        DeviceCapability = C.LIBMTP_DEVICECAP_EditObjects
	MoveObject        DeviceCapability = C.LIBMTP_DEVICECAP_MoveObject
	CopyObject        DeviceCapability = C.LIBMTP_DEVICECAP_CopyObject
)

// SECTION: GET / OPEN DEVICES

// GetRawDevices, gets a list of all RawDevice's connected to the host
func GetRawDevices() ([]*RawDevice, error) {
	var (
		rdl []*RawDevice

		devs    *C.LIBMTP_raw_device_t
		numDevs C.int
	)
	errNo := C.LIBMTP_Detect_Raw_Devices(&devs, &numDevs)
	if errNo != 0 {
		return rdl, fmt.Errorf("failed got error code: %d, maybe no MTP devices detected", errNo)
	}

	slice := unsafe.Slice(&devs, int(numDevs))
	rdl = make([]*RawDevice, numDevs)
	for i, d := range slice {
		rdl[i] = (*RawDevice)(d)
	}
	return rdl, nil
}

// OpenRawDevice, opens a raw device as an `MTPDevice`
//
// NOTE: It's the callers responsibility to ensure `ReleaseDevice` is called for each `OpenRawDevice`
func OpenRawDevice(d *RawDevice) *MTPDevice {
	mtpDevice := C.LIBMTP_Open_Raw_Device_Uncached((*C.LIBMTP_raw_device_t)(d))
	return (*MTPDevice)(mtpDevice)
}

// ReleaseDevice, releases an opened mtp device
func ReleaseDevice(m *MTPDevice) {
	C.LIBMTP_Release_Device((*C.LIBMTP_mtpdevice_t)(m))
}

// GetStorage, updates the storage id's on a device, creates a linked list of them and puts the head into `MTPDevice`
func (md *MTPDevice) GetStorage() error {
	// TODO: Last param is to specify sorting, maybe make it configurable in future?
	errNo := C.LIBMTP_Get_Storage((*C.LIBMTP_mtpdevice_t)(md), 0)
	if errNo != 0 {
		return fmt.Errorf("failed got error code: %d", errNo)
	}
	return nil
}

// SECTION: Get device information

func (m *MTPDevice) GetSerialNumber() string {
	s := C.LIBMTP_Get_Serialnumber((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(s)
}

func (m *MTPDevice) GetManufacturerName() string {
	s := C.LIBMTP_Get_Manufacturername((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(s)
}

func (m *MTPDevice) GetModelName() string {
	s := C.LIBMTP_Get_Modelname((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(s)
}

func (m *MTPDevice) GetDeviceVersion() string {
	s := C.LIBMTP_Get_Deviceversion((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(s)
}

func (m *MTPDevice) GetFriendlyName() string {
	s := C.LIBMTP_Get_Friendlyname((*C.LIBMTP_mtpdevice_t)(m))
	return C.GoString(s)
}

// GetBatteryLevel, returns that `max` and `curr` battery level (if supported)
func (m *MTPDevice) GetBatteryLevel() (uint8, uint8, error) {
	var (
		max  = C.uint8_t(0)
		curr = C.uint8_t(0)
	)
	errNo := C.LIBMTP_Get_Batterylevel((*C.LIBMTP_mtpdevice_t)(m), &max, &curr)
	if errNo != 0 {
		return 0, 0, fmt.Errorf("failed got error code: %d, battery level likely unsupported", errNo)
	}
	return uint8(max), uint8(curr), nil
}

// CheckCapability, checks if the device is capable of a `DeviceCapability`
func (m *MTPDevice) CheckCapability(cap DeviceCapability) bool {
	status := C.LIBMTP_Check_Capability((*C.LIBMTP_mtpdevice_t)(m), (C.LIBMTP_devicecap_t)(cap))
	return status != 0
}

// SECTION: File and folder access

// GetFilesAndFolders, gets a list of files and folders from `storage` and `parent` (folder) ids
func GetFilesAndFolders(md *MTPDevice, storage, parent uint32) ([]File, error) {
	var (
		ret = []File{}
	)
	if md == nil {
		return ret, errors.New("nil device provided")
	}

	fl := C.LIBMTP_Get_Files_And_Folders((*C.LIBMTP_mtpdevice_t)(md), C.uint32_t(storage), C.uint32_t(parent))
	if fl == nil {
		return ret, errors.New("no files or folders found")
	}

	for p := fl; p != nil; p = p.next {
		gf := libmtpToGoFileStruct(p)
		ret = append(ret, gf)
		C.LIBMTP_destroy_file_t(p)
	}
	return ret, nil
}

func libmtpToGoFileStruct(libmtpFile *C.LIBMTP_file_t) File {
	return File{
		ItemId:    uint32(libmtpFile.item_id),
		ParentId:  uint32(libmtpFile.parent_id),
		StorageId: uint32(libmtpFile.storage_id),

		Filename:         C.GoString(libmtpFile.filename),
		Filesize:         uint64(libmtpFile.filesize),
		ModificationDate: time.Unix(int64(libmtpFile.modificationdate), 0),

		FileType: libmtpFile.filetype,
	}
}

// GetFileToFile, copies a file off the device to a local file at `path`
func (m *MTPDevice) GetFileToFile(id uint32, path string) error {
	// TODO: Maybe add support for last two argument in the future:
	// - "callback" (progress indicator function)
	// - "data" (user-defined pointer to pass data to the progress updated)
	// SOURCE: https://github.com/libmtp/libmtp/blob/41786891edcb4b57cf22f2721b164e1416d1feb5/src/libmtp.c#L5331
	cPath := C.CString(path)
	errNo := C.LIBMTP_Get_File_To_File((*C.LIBMTP_mtpdevice_t)(m), C.uint32_t(id), cPath, nil, nil)
	C.free(unsafe.Pointer(cPath))
	if errNo != 0 {
		return fmt.Errorf("failed got error code: %d", errNo)
	}
	return nil
}

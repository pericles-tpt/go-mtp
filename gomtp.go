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

type MTPDevice struct {
	ptr *C.LIBMTP_mtpdevice_t

	ObjectBitSize uint8
	Storage       []Storage

	MaximumBatteryLevel    uint8
	DefaultMusicFolder     uint32
	DefaultPlaylistFolder  uint32
	DefaultPictureFolder   uint32
	DefaultVideoFolder     uint32
	DefaultOrganizerFolder uint32
	DefaultZencastFolder   uint32
	DefaultAlbumFolder     uint32
	DefaultTextFolder      uint32

	Cached int

	// TODO: (maybe) params, usbinfo, errorstack, cd, extensions
	// SOURCE: https://github.com/libmtp/libmtp/blob/41786891edcb4b57cf22f2721b164e1416d1feb5/src/libmtp.h.in#L635
}

type Storage struct {
	Id                 uint32
	StorageType        uint16
	FilesystemType     uint16
	AccessCapability   uint16
	MaxCapacity        uint64
	FreeSpaceInBytes   uint64
	FreeSpaceInObjects uint64
	StorageDescription string
	VolumeIdentifier   string
}

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
	libmtpDevice := C.LIBMTP_Open_Raw_Device_Uncached((*C.LIBMTP_raw_device_t)(d))
	mtpDevice := libmtpToGoMTPDeviceStruct(libmtpDevice)
	return &mtpDevice
}

func libmtpToGoMTPDeviceStruct(m *C.LIBMTP_mtpdevice_t) MTPDevice {
	dev := MTPDevice{
		ptr: m,

		ObjectBitSize: (uint8)(m.object_bitsize),

		MaximumBatteryLevel:    (uint8)(m.maximum_battery_level),
		DefaultMusicFolder:     (uint32)(m.default_music_folder),
		DefaultPlaylistFolder:  (uint32)(m.default_playlist_folder),
		DefaultPictureFolder:   (uint32)(m.default_picture_folder),
		DefaultVideoFolder:     (uint32)(m.default_video_folder),
		DefaultOrganizerFolder: (uint32)(m.default_organizer_folder),
		DefaultZencastFolder:   (uint32)(m.default_zencast_folder),
		DefaultAlbumFolder:     (uint32)(m.default_album_folder),
		DefaultTextFolder:      (uint32)(m.default_text_folder),

		Cached: (int)(m.cached),
	}

	var numStorage int
	for f := m.storage; f != nil; f = f.next {
		numStorage++
	}
	dev.Storage = make([]Storage, 0, numStorage)
	for f := m.storage; f != nil; f = f.next {
		dev.Storage = append(dev.Storage, Storage{
			Id:                 (uint32)(f.id),
			StorageType:        (uint16)(f.StorageType),
			FilesystemType:     (uint16)(f.FilesystemType),
			AccessCapability:   (uint16)(f.FreeSpaceInBytes),
			MaxCapacity:        (uint64)(f.MaxCapacity),
			FreeSpaceInBytes:   (uint64)(f.FreeSpaceInBytes),
			FreeSpaceInObjects: (uint64)(f.next.FreeSpaceInObjects),
			StorageDescription: C.GoString(f.StorageDescription),
			VolumeIdentifier:   C.GoString(f.VolumeIdentifier),
		})
	}

	return dev
}

// ReleaseDevice, releases an opened mtp device
func (m *MTPDevice) ReleaseDevice() {
	C.LIBMTP_Release_Device(m.ptr)
}

// GetStorage, updates the storage id's on a device, creates a linked list of them and puts the head into `MTPDevice`
func (m *MTPDevice) GetStorage() error {
	// TODO: Last param is to specify sorting, maybe make it configurable in future?
	errNo := C.LIBMTP_Get_Storage((*C.LIBMTP_mtpdevice_t)(m.ptr), 0)
	if errNo != 0 {
		return fmt.Errorf("failed got error code: %d", errNo)
	}
	return nil
}

// SECTION: Get device information

func (m *MTPDevice) GetSerialNumber() string {
	s := C.LIBMTP_Get_Serialnumber((*C.LIBMTP_mtpdevice_t)(m.ptr))
	return C.GoString(s)
}

func (m *MTPDevice) GetManufacturerName() string {
	s := C.LIBMTP_Get_Manufacturername((*C.LIBMTP_mtpdevice_t)(m.ptr))
	return C.GoString(s)
}

func (m *MTPDevice) GetModelName() string {
	s := C.LIBMTP_Get_Modelname((*C.LIBMTP_mtpdevice_t)(m.ptr))
	return C.GoString(s)
}

func (m *MTPDevice) GetDeviceVersion() string {
	s := C.LIBMTP_Get_Deviceversion((*C.LIBMTP_mtpdevice_t)(m.ptr))
	return C.GoString(s)
}

func (m *MTPDevice) GetFriendlyName() string {
	s := C.LIBMTP_Get_Friendlyname((*C.LIBMTP_mtpdevice_t)(m.ptr))
	return C.GoString(s)
}

// GetBatteryLevel, returns that `max` and `curr` battery level (if supported)
func (m *MTPDevice) GetBatteryLevel() (uint8, uint8, error) {
	var (
		max  = C.uint8_t(0)
		curr = C.uint8_t(0)
	)
	errNo := C.LIBMTP_Get_Batterylevel((*C.LIBMTP_mtpdevice_t)(m.ptr), &max, &curr)
	if errNo != 0 {
		return 0, 0, fmt.Errorf("failed got error code: %d, battery level likely unsupported", errNo)
	}
	return uint8(max), uint8(curr), nil
}

// CheckCapability, checks if the device is capable of a `DeviceCapability`
func (m *MTPDevice) CheckCapability(cap DeviceCapability) bool {
	status := C.LIBMTP_Check_Capability((*C.LIBMTP_mtpdevice_t)(m.ptr), (C.LIBMTP_devicecap_t)(cap))
	return status != 0
}

// SECTION: File and folder access

// GetFilesAndFolders, gets a list of files and folders from `storage` and `parent` (folder) ids
func (m *MTPDevice) GetFilesAndFolders(storage, parent uint32) ([]File, []File, error) {
	var (
		files   = []File{}
		folders = []File{}
	)
	if m == nil {
		return files, folders, errors.New("nil device provided")
	}

	fl := C.LIBMTP_Get_Files_And_Folders((*C.LIBMTP_mtpdevice_t)(m.ptr), C.uint32_t(storage), C.uint32_t(parent))
	if fl == nil {
		return files, folders, errors.New("no files or folders found")
	}

	for p := fl; p != nil; p = p.next {
		gf := libmtpToGoFileStruct(p)
		if p.filetype == C.LIBMTP_FILETYPE_FOLDER {
			folders = append(folders, gf)
		} else {
			files = append(files, gf)
		}
		C.LIBMTP_destroy_file_t(p)
	}
	return files, folders, nil
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
	errNo := C.LIBMTP_Get_File_To_File((*C.LIBMTP_mtpdevice_t)(m.ptr), C.uint32_t(id), cPath, nil, nil)
	C.free(unsafe.Pointer(cPath))
	if errNo != 0 {
		return fmt.Errorf("failed got error code: %d", errNo)
	}
	return nil
}

package main

import "fmt"

func GetDeviceBySerialNumber(targetSerialNo string) (*MTPDevice, error) {
	devList, err := GetRawDevices()
	if err != nil {
		return nil, err
	}

	for _, maybeTarget := range devList {
		mtpDevice := OpenRawDevice(maybeTarget)
		if mtpDevice == nil {
			continue
		}

		sno := mtpDevice.GetSerialNumber()
		if sno == targetSerialNo {
			return mtpDevice, nil
		}
		ReleaseDevice(mtpDevice)
	}

	return nil, fmt.Errorf("no devices found matching serial number: %s", targetSerialNo)
}

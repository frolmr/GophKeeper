package domain

import (
	"os"

	"github.com/denisbrodbeck/machineid"
)

// DeviceOnServer represents device information as stored on the server
type DeviceOnServer struct {
	ID        string // Unique device identifier
	Name      string // Human-readable device name
	Confirmed bool   // Whether device is approved for use
}

// Device represents the current device's identity
type Device struct {
	ID   string // Protected machine identifier
	Name string // Hostname of the device
}

// ThisDevice generates identity information for the current device.
// Combines a protected machine ID with the hostname.
//
// Returns:
//
//	*Device - Current device information
//	error   - If machine ID or hostname cannot be determined
//
// Security:
// - Uses machineid.ProtectedID to create a stable but app-specific identifier
// - Hostname is used only for display purposes
func ThisDevice() (*Device, error) {
	protectedID, err := machineid.ProtectedID(AppName)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &Device{
		ID:   protectedID,
		Name: hostname,
	}, nil
}

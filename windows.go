// +build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetDriveTypeW  = kernel32.NewProc("GetDriveTypeW")
	procGetVolumeInformationW = kernel32.NewProc("GetVolumeInformationW")
)

const (
	DRIVE_UNKNOWN     = 0
	DRIVE_NO_ROOT_DIR = 1
	DRIVE_REMOVABLE   = 2
	DRIVE_FIXED       = 3
	DRIVE_REMOTE      = 4
	DRIVE_CDROM       = 5
	DRIVE_RAMDISK     = 6
)

// GetDriveType returns the type of drive for the given drive letter
func GetDriveType(driveLetter string) (uint32, error) {
	rootPath := fmt.Sprintf("%s:\\", driveLetter)
	rootPathPtr, err := syscall.UTF16PtrFromString(rootPath)
	if err != nil {
		return 0, err
	}
	
	ret, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(rootPathPtr)))
	return uint32(ret), nil
}

// IsPhysicalDrive returns true if the drive is a physical drive that supports device path access
func IsPhysicalDrive(driveLetter string) bool {
	driveType, err := GetDriveType(driveLetter)
	if err != nil {
		return false
	}
	
	// Only FIXED and CDROM drives typically support device path access
	// REMOVABLE might work but can be problematic with virtual drives
	// REMOTE and RAMDISK definitely won't work with device paths
	switch driveType {
	case DRIVE_FIXED, DRIVE_CDROM:
		return true
	case DRIVE_REMOTE, DRIVE_RAMDISK:
		return false
	case DRIVE_REMOVABLE:
		// For removable drives, we need to check if it's truly physical
		// Try to detect if it's a virtual mount
		return !isVirtualMount(driveLetter)
	default:
		return false
	}
}

// isVirtualMount checks if a drive is a virtual mount (like mounted ISO)
func isVirtualMount(driveLetter string) bool {
	// Try to open the device path - if it fails, it's likely virtual
	devicePath := fmt.Sprintf("\\\\.\\%s:", driveLetter)
	file, err := os.Open(devicePath)
	if err != nil {
		return true
	}
	defer file.Close()
	
	// Try to stat the device - if this fails on 32-bit, it's likely virtual or problematic
	_, err = file.Stat()
	if err != nil {
		return true
	}
	
	return false
}

// GetDriveTypeString returns a human-readable string for the drive type
func GetDriveTypeString(driveType uint32) string {
	switch driveType {
	case DRIVE_UNKNOWN:
		return "Unknown"
	case DRIVE_NO_ROOT_DIR:
		return "Invalid"
	case DRIVE_REMOVABLE:
		return "Removable"
	case DRIVE_FIXED:
		return "Fixed"
	case DRIVE_REMOTE:
		return "Network"
	case DRIVE_CDROM:
		return "CD-ROM"
	case DRIVE_RAMDISK:
		return "RAM Disk"
	default:
		return "Unknown"
	}
}

// GetDriveSizeFromFilesystem attempts to get the drive size by reading files
// This is a fallback for when device path access doesn't work
func GetDriveSizeFromFilesystem(drivePath string) (int64, error) {
	// For virtual drives, we can't reliably get the size
	// Return an error so the caller knows to skip size-dependent operations
	return 0, fmt.Errorf("cannot determine size for virtual/network drives")
}

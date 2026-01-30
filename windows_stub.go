// +build !windows

package main

// IsPhysicalDrive always returns false on non-Windows platforms
func IsPhysicalDrive(driveLetter string) bool {
	return false
}

// GetDriveType is not supported on non-Windows platforms
func GetDriveType(driveLetter string) (uint32, error) {
	return 0, nil
}

// GetDriveTypeString returns empty string on non-Windows
func GetDriveTypeString(driveType uint32) string {
	return ""
}

// GetDriveSizeFromFilesystem is not needed on non-Windows
func GetDriveSizeFromFilesystem(drivePath string) (int64, error) {
	return 0, nil
}

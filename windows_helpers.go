//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

// getDriveLetters returns a list of CD-ROM drive letters on Windows
func getDriveLetters() []string {
	var drives []string

	// Load kernel32.dll to access Windows API
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDriveType := kernel32.NewProc("GetDriveTypeW")

	// Check all possible drive letters A-Z
	for drive := 'A'; drive <= 'Z'; drive++ {
		drivePath := fmt.Sprintf("%c:\\", drive)
		drivePathUTF16, err := syscall.UTF16PtrFromString(drivePath)
		if err != nil {
			continue
		}

		// GetDriveType returns:
		// 0 = DRIVE_UNKNOWN
		// 1 = DRIVE_NO_ROOT_DIR
		// 2 = DRIVE_REMOVABLE
		// 3 = DRIVE_FIXED
		// 4 = DRIVE_REMOTE
		// 5 = DRIVE_CDROM
		// 6 = DRIVE_RAMDISK
		driveType, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(drivePathUTF16)))

		// Only include CD-ROM drives (type 5)
		if driveType == 5 {
			drives = append(drives, fmt.Sprintf("%c:", drive))
		}
	}

	return drives
}

// getCurrentDrive returns the drive letter where the executable is located
func getCurrentDrive() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}

	// Get the drive letter from the path
	vol := filepath.VolumeName(exePath)
	if len(vol) >= 2 && vol[1] == ':' {
		return vol[:2] // Return just "C:" format
	}

	return ""
}

// isCheckisomd5Available checks if checkisomd5.exe is available in PATH or exe directory
func isCheckisomd5Available() bool {
	// First check if it's in PATH
	_, err := exec.LookPath("checkisomd5.exe")
	if err == nil {
		return true
	}

	// Check in the same directory as the executable
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	exeDir := filepath.Dir(exePath)
	checkisomd5Path := filepath.Join(exeDir, "checkisomd5.exe")

	_, err = os.Stat(checkisomd5Path)
	return err == nil
}

// isDriveReady checks if a drive has media inserted using Windows API
func isDriveReady(drive string) bool {
	// Ensure drive ends with backslash for Windows API
	drivePath := drive
	if len(drivePath) == 2 && drivePath[1] == ':' {
		drivePath += "\\"
	}

	drivePathUTF16, err := syscall.UTF16PtrFromString(drivePath)
	if err != nil {
		return false
	}

	// Load kernel32.dll to access Windows API
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getVolumeInformation := kernel32.NewProc("GetVolumeInformationW")

	// Try to get volume information
	// If this succeeds, the drive has media
	ret, _, _ := getVolumeInformation.Call(
		uintptr(unsafe.Pointer(drivePathUTF16)),
		0,  // lpVolumeNameBuffer
		0,  // nVolumeNameSize
		0,  // lpVolumeSerialNumber
		0,  // lpMaximumComponentLength
		0,  // lpFileSystemFlags
		0,  // lpFileSystemNameBuffer
		0,  // nFileSystemNameSize
	)

	return ret != 0
}

// runCheckisomd5 runs the external checkisomd5.exe tool
func runCheckisomd5(config *Config) error {
	// Find checkisomd5.exe
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	
	checkisoPath := ""
	// Try exe directory first
	localPath := filepath.Join(exeDir, "checkisomd5.exe")
	if _, err := os.Stat(localPath); err == nil {
		checkisoPath = localPath
	} else {
		// Try PATH
		if path, err := exec.LookPath("checkisomd5.exe"); err == nil {
			checkisoPath = path
		} else {
			return fmt.Errorf("checkisomd5.exe not found")
		}
	}
	
	// Run checkisomd5.exe with -v (verbose) flag
	cmd := exec.Command(checkisoPath, "-v", config.Path)
	
	// Capture combined output
	output, err := cmd.CombinedOutput()
	
	// Print the output
	fmt.Print(string(output))
	
	// Check exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// checkisomd5 returns non-zero on failure
			if exitErr.ExitCode() != 0 {
				hasErrors = true
			}
		}
		return err
	}
	
	return nil
}

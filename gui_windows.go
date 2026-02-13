//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var (
	modkernel32     = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole = modkernel32.NewProc("AttachConsole")
	procGetConsoleWindow = modkernel32.NewProc("GetConsoleWindow")
)

const (
	ATTACH_PARENT_PROCESS = ^uint32(0) // (DWORD)-1
)

// hasConsole checks if the process has a console attached
func hasConsole() bool {
	hwnd, _, _ := procGetConsoleWindow.Call()
	return hwnd != 0
}

// attachParentConsole attempts to attach to the parent console
func attachParentConsole() bool {
	ret, _, _ := procAttachConsole.Call(uintptr(ATTACH_PARENT_PROCESS))
	return ret != 0
}

// getDriveLetters returns a list of available drive letters on Windows
func getDriveLetters() []string {
	var drives []string
	
	// Get logical drives bitmask
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getLogicalDrives := kernel32.NewProc("GetLogicalDrives")
	
	ret, _, _ := getLogicalDrives.Call()
	if ret == 0 {
		return drives
	}
	
	// Check each bit for drive letters A-Z
	for i := 0; i < 26; i++ {
		if ret&(1<<uint(i)) != 0 {
			drive := string(rune('A' + i))
			// Check if it's a CD-ROM drive using GetDriveType
			driveType := getDriveType(drive + ":\\")
			// DRIVE_CDROM = 5
			if driveType == 5 {
				drives = append(drives, drive+":")
			}
		}
	}
	
	return drives
}

// getDriveType returns the drive type for a given path
func getDriveType(path string) uint32 {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDriveTypeW := kernel32.NewProc("GetDriveTypeW")
	
	pathPtr, _ := syscall.UTF16PtrFromString(path)
	ret, _, _ := getDriveTypeW.Call(uintptr(unsafe.Pointer(pathPtr)))
	
	return uint32(ret)
}

// getCurrentDrive returns the drive letter the executable is running from
func getCurrentDrive() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	
	// Get the drive letter from the executable path
	if len(exePath) >= 2 && exePath[1] == ':' {
		drive := strings.ToUpper(string(exePath[0])) + ":"
		// Check if it's a CD-ROM drive
		driveType := getDriveType(drive + "\\")
		if driveType == 5 {
			return drive
		}
	}
	
	return ""
}

// runGUI starts the GUI mode
func runGUI() {
	var mainWindow *walk.MainWindow
	var driveComboBox *walk.ComboBox
	var resultTextEdit *walk.TextEdit
	var verifyButton *walk.PushButton
	
	drives := getDriveLetters()
	if len(drives) == 0 {
		walk.MsgBox(nil, "Error", "No CD-ROM drives found on this system.", walk.MsgBoxIconError)
		return
	}
	
	// Get current drive if running from a drive
	currentDrive := getCurrentDrive()
	defaultIndex := 0
	if currentDrive != "" {
		for i, drive := range drives {
			if drive == currentDrive {
				defaultIndex = i
				break
			}
		}
	}
	
	err := MainWindow{
		AssignTo: &mainWindow,
		Title:    fmt.Sprintf("chkiso - ISO/Drive Verification Tool v%s", VERSION),
		MinSize:  Size{Width: 600, Height: 400},
		Size:     Size{Width: 700, Height: 500},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 3},
				Children: []Widget{
					Label{
						Text: "Select Drive:",
					},
					ComboBox{
						AssignTo:      &driveComboBox,
						Model:         drives,
						CurrentIndex:  defaultIndex,
						MinSize:       Size{Width: 100},
					},
					PushButton{
						AssignTo: &verifyButton,
						Text:     "Verify",
						OnClicked: func() {
							verifyDrive(driveComboBox, resultTextEdit, verifyButton)
						},
					},
				},
			},
			TextEdit{
				AssignTo: &resultTextEdit,
				ReadOnly: true,
				VScroll:  true,
				Font:     Font{Family: "Courier New", PointSize: 9},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Close",
						OnClicked: func() {
							mainWindow.Close()
						},
					},
				},
			},
		},
	}.Create()
	
	if err != nil {
		walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to create window: %v", err), walk.MsgBoxIconError)
		return
	}
	
	mainWindow.Run()
}

// verifyDrive performs the verification for the selected drive
func verifyDrive(driveCombo *walk.ComboBox, resultText *walk.TextEdit, verifyBtn *walk.PushButton) {
	// Get selected drive
	selectedIndex := driveCombo.CurrentIndex()
	if selectedIndex < 0 {
		resultText.SetText("Error: No drive selected")
		return
	}
	
	model := driveCombo.Model()
	drives, ok := model.([]string)
	if !ok || selectedIndex >= len(drives) {
		resultText.SetText("Error: Invalid drive selection")
		return
	}
	
	selectedDrive := drives[selectedIndex]
	
	// Disable button during verification
	verifyBtn.SetEnabled(false)
	
	resultText.SetText(fmt.Sprintf("Verifying drive %s...\n\nPlease wait, this may take a few minutes...\n\n", selectedDrive))
	
	// Run verification in a goroutine to prevent UI freezing
	go func() {
		defer func() {
			// Re-enable button when done
			verifyBtn.Synchronize(func() {
				verifyBtn.SetEnabled(true)
			})
		}()
		
		// Create a config for the verification
		config := &Config{
			Path:     selectedDrive,
			NoVerify: false,
			MD5Check: false,
		}
		
		// Validate path
		if err := validatePath(config); err != nil {
			resultText.Synchronize(func() {
				resultText.AppendText(fmt.Sprintf("Error: %v\n", err))
			})
			return
		}
		
		// Capture output
		output := &strings.Builder{}
		
		// Run verification (we'll capture the output)
		output.WriteString(fmt.Sprintf("=== Verifying Drive %s ===\n\n", selectedDrive))
		
		// Display SHA256 Hash
		output.WriteString("--- SHA256 Hash (Informational) ---\n")
		calculatedHash, err := getSha256FromPath(config)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error calculating hash: %v\n", err))
		} else {
			output.WriteString(fmt.Sprintf("SHA256: %s\n", strings.ToLower(calculatedHash)))
		}
		output.WriteString("\n")
		
		// Verify contents
		output.WriteString("--- Verifying Contents ---\n")
		mountPath := fmt.Sprintf("%s\\", selectedDrive)
		output.WriteString(fmt.Sprintf("Verifying contents of physical drive at: %s\n", mountPath))
		output.WriteString(fmt.Sprintf("Searching for checksum files (*.sha, sha256sum.txt, SHA256SUMS) in %s...\n", mountPath))
		
		// Find checksum files
		checksumFiles, err := findChecksumFiles(mountPath)
		if err != nil {
			output.WriteString(fmt.Sprintf("Warning: Error finding checksum files: %v\n", err))
		} else if len(checksumFiles) == 0 {
			output.WriteString("Warning: Could not find any checksum files (*.sha, sha256sum.txt, SHA256SUMS) on the media.\n")
		} else {
			output.WriteString(fmt.Sprintf("\nFound %d checksum file(s):\n", len(checksumFiles)))
			for i, cf := range checksumFiles {
				relPath, err := filepath.Rel(mountPath, cf)
				if err != nil {
					relPath = cf
				}
				output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, relPath))
			}
			output.WriteString("\n")
			
			totalFiles := 0
			failedFiles := 0
			
			for _, checksumFile := range checksumFiles {
				output.WriteString(fmt.Sprintf("Processing checksum file: %s\n", filepath.Base(checksumFile)))
				baseDir := filepath.Dir(checksumFile)
				
				// Process checksum file
				files, failed := processChecksumFile(checksumFile, baseDir, output)
				totalFiles += files
				failedFiles += failed
				output.WriteString("\n")
			}
			
			output.WriteString("--- Verification Summary ---\n")
			output.WriteString(fmt.Sprintf("Checksum files processed: %d\n", len(checksumFiles)))
			output.WriteString(fmt.Sprintf("Total files verified: %d\n", totalFiles))
			
			if failedFiles == 0 && totalFiles > 0 {
				output.WriteString(fmt.Sprintf("Success: All %d files verified successfully.\n", totalFiles))
			} else if totalFiles == 0 {
				output.WriteString("No files were verified.\n")
			} else {
				output.WriteString(fmt.Sprintf("Failure: %d out of %d files failed verification.\n", failedFiles, totalFiles))
			}
		}
		
		// Update the result text
		resultText.Synchronize(func() {
			resultText.SetText(output.String())
		})
	}()
}

// processChecksumFile processes a single checksum file and returns (totalFiles, failedFiles)
func processChecksumFile(checksumFile, baseDir string, output *strings.Builder) (int, int) {
	totalFiles := 0
	failedFiles := 0
	
	// Read file content
	content, err := os.ReadFile(checksumFile)
	if err != nil {
		output.WriteString(fmt.Sprintf("Warning: Could not read checksum file: %v\n", err))
		return totalFiles, failedFiles
	}
	
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Match SHA256 hash pattern
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		expectedHash := strings.ToLower(parts[0])
		if len(expectedHash) != 64 {
			continue
		}
		
		// Get filename (everything after the hash, removing optional asterisk)
		fileName := strings.TrimSpace(strings.TrimPrefix(line, parts[0]))
		fileName = strings.TrimPrefix(fileName, "*")
		fileName = strings.TrimSpace(fileName)
		
		if fileName == "" {
			continue
		}
		
		totalFiles++
		
		// Validate that the file path doesn't escape the base directory
		filePathOnMedia := filepath.Join(baseDir, fileName)
		cleanPath := filepath.Clean(filePathOnMedia)
		if !strings.HasPrefix(cleanPath, filepath.Clean(baseDir)) {
			output.WriteString(fmt.Sprintf("Warning: Skipping potentially unsafe path: %s (referenced in %s)\n", fileName, filepath.Base(checksumFile)))
			failedFiles++
			continue
		}
		
		if _, err := os.Stat(filePathOnMedia); os.IsNotExist(err) {
			output.WriteString(fmt.Sprintf("Warning: File not found on media: %s (referenced in %s)\n", fileName, filepath.Base(checksumFile)))
			failedFiles++
			continue
		}
		
		output.WriteString(fmt.Sprintf("Verifying: %s", fileName))
		calculatedHash, err := getSha256Hash(filePathOnMedia)
		if err != nil {
			output.WriteString(fmt.Sprintf(" -> ERROR: %v\n", err))
			failedFiles++
			continue
		}
		
		calculatedHash = strings.ToLower(calculatedHash)
		if calculatedHash == expectedHash {
			output.WriteString(" -> OK\n")
		} else {
			output.WriteString(" -> FAILED\n")
			failedFiles++
		}
	}
	
	return totalFiles, failedFiles
}

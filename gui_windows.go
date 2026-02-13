//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
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

// isDriveReady checks if a drive has media inserted and is ready
func isDriveReady(driveLetter string) bool {
	// Try to access the drive root
	drivePath := driveLetter
	if !strings.HasSuffix(drivePath, ":\\") {
		drivePath = driveLetter + ":\\"
	}
	
	// Try to read the root directory
	_, err := os.ReadDir(drivePath)
	return err == nil
}

// isCheckisomd5Available checks if checkisomd5.exe is available in PATH or current directory
func isCheckisomd5Available() bool {
	// Check in PATH
	_, err := exec.LookPath("checkisomd5.exe")
	if err == nil {
		return true
	}
	
	// Check in same directory as executable
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	
	exeDir := filepath.Dir(exePath)
	checkisomd5Path := filepath.Join(exeDir, "checkisomd5.exe")
	
	_, err = os.Stat(checkisomd5Path)
	return err == nil
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
	logDebug("runGUI() called")
	
	var mainWindow *walk.MainWindow
	var driveComboBox *walk.ComboBox
	var resultTextEdit *walk.TextEdit
	var verifyButton *walk.PushButton
	var md5CheckBox *walk.CheckBox
	
	logDebug("Getting drive letters...")
	drives := getDriveLetters()
	logDebug("Found %d CD-ROM drives: %v", len(drives), drives)
	
	// Get current drive if running from a drive
	currentDrive := getCurrentDrive()
	logDebug("Current drive: %s", currentDrive)
	defaultIndex := 0
	if currentDrive != "" && len(drives) > 0 {
		for i, drive := range drives {
			if drive == currentDrive {
				defaultIndex = i
				break
			}
		}
	}
	
	// If no drives found, add a placeholder message so the window still shows
	if len(drives) == 0 {
		drives = []string{"<No CD-ROM drives found>"}
		defaultIndex = 0
		logDebug("No CD-ROM drives found, using placeholder")
	}
	
	// Check if checkisomd5.exe is available
	logDebug("Checking for checkisomd5.exe...")
	md5Available := isCheckisomd5Available()
	logDebug("checkisomd5.exe available: %v", md5Available)
	
	// Build the children widgets dynamically
	var children []Widget
	
	// Add drive selection row
	children = append(children, Composite{
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
					md5Check := false
					if md5CheckBox != nil {
						md5Check = md5CheckBox.Checked()
					}
					verifyDriveWithOptions(driveComboBox, resultTextEdit, verifyButton, mainWindow, md5Check)
				},
			},
		},
	})
	
	// Add browse button and MD5 checkbox row
	var browseRowChildren []Widget
	browseRowChildren = append(browseRowChildren, PushButton{
		Text: "Browse for ISO file...",
		OnClicked: func() {
			md5Check := false
			if md5CheckBox != nil {
				md5Check = md5CheckBox.Checked()
			}
			browseForISOWithOptions(resultTextEdit, verifyButton, mainWindow, md5Check)
		},
	})
	
	// Add MD5 checkbox if checkisomd5.exe is available
	if md5Available {
		browseRowChildren = append(browseRowChildren, CheckBox{
			AssignTo: &md5CheckBox,
			Text:     "Verify implanted MD5 (checkisomd5)",
		})
	}
	
	browseRowChildren = append(browseRowChildren, HSpacer{})
	
	children = append(children, Composite{
		Layout: HBox{},
		Children: browseRowChildren,
	})
	
	// Add text area
	children = append(children, TextEdit{
		AssignTo: &resultTextEdit,
		ReadOnly: true,
		VScroll:  true,
		Font:     Font{Family: "Courier New", PointSize: 9},
	})
	
	// Add close button
	children = append(children, Composite{
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
	})
	
	logDebug("Building GUI window with %d children widgets", len(children))
	logDebug("Creating MainWindow...")
	
	err := MainWindow{
		AssignTo: &mainWindow,
		Title:    fmt.Sprintf("chkiso - ISO/Drive Verification Tool v%s", VERSION),
		MinSize:  Size{Width: 600, Height: 400},
		Size:     Size{Width: 700, Height: 500},
		Layout:   VBox{},
		OnDropFiles: func(files []string) {
			md5Check := false
			if md5CheckBox != nil {
				md5Check = md5CheckBox.Checked()
			}
			handleDroppedFilesWithOptions(files, resultTextEdit, verifyButton, mainWindow, md5Check)
		},
		Children: children,
	}.Create()
	
	if err != nil {
		logDebug("ERROR: Failed to create window: %v", err)
		
		// Get the log file path to show user
		logPath := ""
		if logFile != nil {
			logPath = logFile.Name()
		}
		
		errorMsg := fmt.Sprintf("Failed to create window: %v\n\n", err)
		errorMsg += "This error can occur due to:\n"
		errorMsg += "• Windows Common Controls issues\n"
		errorMsg += "• Too many GUI elements\n"
		errorMsg += "• System resource constraints\n\n"
		
		if logPath != "" {
			errorMsg += fmt.Sprintf("Debug log saved to:\n%s\n\n", logPath)
			errorMsg += "Please check the log file for more details."
		}
		
		walk.MsgBox(nil, "Error Creating GUI", errorMsg, walk.MsgBoxIconError)
		return
	}
	
	logDebug("MainWindow created successfully")
	
	// If no drives were found, show a helpful error message in the text area
	if len(drives) == 1 && drives[0] == "<No CD-ROM drives found>" {
		logDebug("Setting initial message for no drives found")
		resultTextEdit.SetText("No CD-ROM drives detected on this system.\n\n" +
			"To verify an ISO file:\n" +
			"  • Click 'Browse for ISO file...' button below, or\n" +
			"  • Drag and drop an ISO file onto this window\n\n" +
			"To verify a CD/DVD drive:\n" +
			"  1. Insert a bootable CD/DVD into a drive\n" +
			"  2. Or mount an ISO file using Windows Explorer (right-click → Mount)\n" +
			"  3. Then relaunch this application\n\n" +
			"Command-line usage is also available:\n" +
			"  chkiso.exe path\\to\\image.iso")
		verifyButton.SetEnabled(false)
	} else {
		logDebug("Setting initial ready message")
		// Show helpful hint about drag and drop
		resultTextEdit.SetText("Ready to verify.\n\n" +
			"Select a drive from the dropdown and click 'Verify',\n" +
			"or click 'Browse for ISO file...',\n" +
			"or drag and drop an ISO file onto this window.")
	}
	
	logDebug("Starting GUI event loop (mainWindow.Run)")
	mainWindow.Run()
	logDebug("GUI event loop ended")
}

// handleDroppedFiles processes files dropped onto the window
func handleDroppedFiles(files []string, resultText *walk.TextEdit, verifyBtn *walk.PushButton, owner walk.Form) {
	if len(files) == 0 {
		return
	}
	
	// Only process the first file
	filePath := files[0]
	
	// Check if it's an ISO file
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".iso" {
		resultText.SetText(fmt.Sprintf("Error: Only ISO files are supported.\n\nYou dropped: %s\n\nPlease drop an ISO file (.iso extension) onto this window.", filepath.Base(filePath)))
		return
	}
	
	// Verify the dropped ISO file
	verifyISOFile(filePath, resultText, verifyBtn)
}

// browseForISO opens a file dialog to select an ISO file for verification
func browseForISO(resultText *walk.TextEdit, verifyBtn *walk.PushButton, owner walk.Form) {
	dlg := new(walk.FileDialog)
	dlg.Title = "Select ISO file to verify"
	dlg.Filter = "ISO Files (*.iso)|*.iso|All Files (*.*)|*.*"
	
	accepted, err := dlg.ShowOpen(owner)
	if err != nil {
		resultText.SetText(fmt.Sprintf("Error opening file dialog: %v", err))
		return
	}
	
	if !accepted {
		// User cancelled
		return
	}
	
	isoPath := dlg.FilePath
	if isoPath == "" {
		return
	}
	
	// Verify the selected ISO file
	verifyISOFile(isoPath, resultText, verifyBtn)
}

// verifyISOFile performs verification on an ISO file
func verifyISOFile(isoPath string, resultText *walk.TextEdit, verifyBtn *walk.PushButton) {
	// Disable button during verification
	verifyBtn.SetEnabled(false)
	
	resultText.SetText(fmt.Sprintf("Verifying ISO file: %s\n\nPlease wait, this may take a few minutes...\n\n", filepath.Base(isoPath)))
	
	// Run verification in a goroutine
	go func() {
		defer func() {
			verifyBtn.Synchronize(func() {
				verifyBtn.SetEnabled(true)
			})
		}()
		
		// Create config for the ISO file
		config := &Config{
			Path:     isoPath,
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
		
		output := &strings.Builder{}
		output.WriteString(fmt.Sprintf("=== Verifying ISO File ===\n"))
		output.WriteString(fmt.Sprintf("File: %s\n\n", filepath.Base(isoPath)))
		
		// Display SHA256 Hash
		output.WriteString("--- SHA256 Hash ---\n")
		calculatedHash, err := getSha256FromPath(config)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error calculating hash: %v\n", err))
		} else {
			output.WriteString(fmt.Sprintf("SHA256: %s\n", strings.ToLower(calculatedHash)))
		}
		output.WriteString("\n")
		
		// Try MD5 check
		output.WriteString("--- Checking for Implanted MD5 ---\n")
		md5Result, err := checkImplantedMD5(config)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error: %v\n", err))
		} else if md5Result == nil {
			output.WriteString("No implanted MD5 signature found.\n")
		} else {
			output.WriteString(fmt.Sprintf("Verification Method: %s\n", md5Result.VerificationMethod))
			output.WriteString(fmt.Sprintf("Stored MD5:          %s\n", md5Result.StoredMD5))
			output.WriteString(fmt.Sprintf("Calculated MD5:      %s\n", md5Result.CalculatedMD5))
			if md5Result.IsIntegrityOK {
				output.WriteString("Result: SUCCESS - Implanted MD5 is valid.\n")
			} else {
				output.WriteString("Result: FAILURE - Implanted MD5 does not match.\n")
			}
		}
		output.WriteString("\n")
		
		output.WriteString("--- Summary ---\n")
		output.WriteString("ISO file verification complete.\n")
		output.WriteString("\nNote: Content verification requires the ISO to be mounted.\n")
		output.WriteString("To verify file contents, mount the ISO and select the drive from the dropdown.")
		
		resultText.Synchronize(func() {
			resultText.SetText(output.String())
		})
	}()
}

// verifyDrive performs the verification for the selected drive
func verifyDrive(driveCombo *walk.ComboBox, resultText *walk.TextEdit, verifyBtn *walk.PushButton, owner walk.Form) {
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
	
	// Check if this is the placeholder message for no drives
	if selectedDrive == "<No CD-ROM drives found>" {
		resultText.SetText("Error: No CD-ROM drives available to verify.\n\n" +
			"Click 'Browse for ISO file...' to verify an ISO file from your hard drive.")
		return
	}
	
	// Check if the drive is empty (no media inserted)
	if !isDriveReady(selectedDrive) {
		resultText.SetText(fmt.Sprintf("Drive %s is detected but empty.\n\n", selectedDrive) +
			"Please insert a bootable CD/DVD into the drive and try again.\n\n" +
			"Alternatively:\n" +
			"  • Click 'Browse for ISO file...' to verify an ISO file from your hard drive\n" +
			"  • Mount an ISO file using Windows Explorer (right-click → Mount)\n" +
			"  • Then relaunch this application to verify the mounted drive")
		return
	}
	
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

// Wrapper functions that add md5Check parameter

// verifyDriveWithOptions is a wrapper that adds MD5 check option
func verifyDriveWithOptions(driveCombo *walk.ComboBox, resultText *walk.TextEdit, verifyBtn *walk.PushButton, owner walk.Form, md5Check bool) {
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
	
	// Check if this is the placeholder message for no drives
	if selectedDrive == "<No CD-ROM drives found>" {
		resultText.SetText("Error: No CD-ROM drives available to verify.\n\n" +
			"Click 'Browse for ISO file...' to verify an ISO file from your hard drive.")
		return
	}
	
	// Check if the drive is empty (no media inserted)
	if !isDriveReady(selectedDrive) {
		resultText.SetText(fmt.Sprintf("Drive %s is detected but empty.\n\n", selectedDrive) +
			"Please insert a bootable CD/DVD into the drive and try again.\n\n" +
			"Alternatively:\n" +
			"  • Click 'Browse for ISO file...' to verify an ISO file from your hard drive\n" +
			"  • Mount an ISO file using Windows Explorer (right-click → Mount)\n" +
			"  • Then relaunch this application to verify the mounted drive")
		return
	}
	
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
			MD5Check: md5Check,
		}
		
		// Validate path
		if err := validatePath(config); err != nil {
			resultText.Synchronize(func() {
				resultText.AppendText(fmt.Sprintf("Error: %v\n", err))
			})
			return
		}
		
		output := &strings.Builder{}
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
		
		// Check implanted MD5 if requested
		if md5Check {
			output.WriteString("--- Verifying Implanted MD5 ---\n")
			md5Result, err := checkImplantedMD5(config)
			if err != nil {
				output.WriteString(fmt.Sprintf("Error: %v\n", err))
			} else if md5Result == nil {
				output.WriteString("No implanted MD5 signature found.\n")
			} else {
				output.WriteString(fmt.Sprintf("Verification Method: %s\n", md5Result.VerificationMethod))
				output.WriteString(fmt.Sprintf("Stored MD5:          %s\n", md5Result.StoredMD5))
				output.WriteString(fmt.Sprintf("Calculated MD5:      %s\n", md5Result.CalculatedMD5))
				if md5Result.IsIntegrityOK {
					output.WriteString("Result: SUCCESS - Implanted MD5 is valid.\n")
				} else {
					output.WriteString("Result: FAILURE - Implanted MD5 does not match.\n")
				}
			}
			output.WriteString("\n")
		}
		
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
		
		resultText.Synchronize(func() {
			resultText.SetText(output.String())
		})
	}()
}

// browseForISOWithOptions is a wrapper that adds MD5 check option
func browseForISOWithOptions(resultText *walk.TextEdit, verifyBtn *walk.PushButton, owner walk.Form, md5Check bool) {
	dlg := new(walk.FileDialog)
	dlg.Title = "Select ISO file to verify"
	dlg.Filter = "ISO Files (*.iso)|*.iso|All Files (*.*)|*.*"
	
	accepted, err := dlg.ShowOpen(owner)
	if err != nil {
		resultText.SetText(fmt.Sprintf("Error opening file dialog: %v", err))
		return
	}
	
	if !accepted {
		// User cancelled
		return
	}
	
	isoPath := dlg.FilePath
	if isoPath == "" {
		return
	}
	
	// Verify the selected ISO file
	verifyISOFileWithOptions(isoPath, resultText, verifyBtn, md5Check)
}

// handleDroppedFilesWithOptions is a wrapper that adds MD5 check option
func handleDroppedFilesWithOptions(files []string, resultText *walk.TextEdit, verifyBtn *walk.PushButton, owner walk.Form, md5Check bool) {
	if len(files) == 0 {
		return
	}
	
	// Only process the first file
	filePath := files[0]
	
	// Check if it's an ISO file
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".iso" {
		resultText.SetText(fmt.Sprintf("Error: Only ISO files are supported.\n\nYou dropped: %s\n\nPlease drop an ISO file (.iso extension) onto this window.", filepath.Base(filePath)))
		return
	}
	
	// Verify the dropped ISO file
	verifyISOFileWithOptions(filePath, resultText, verifyBtn, md5Check)
}

// verifyISOFileWithOptions performs verification on an ISO file with MD5 option
func verifyISOFileWithOptions(isoPath string, resultText *walk.TextEdit, verifyBtn *walk.PushButton, md5Check bool) {
	// Disable button during verification
	verifyBtn.SetEnabled(false)
	
	resultText.SetText(fmt.Sprintf("Verifying ISO file: %s\n\nPlease wait, this may take a few minutes...\n\n", filepath.Base(isoPath)))
	
	// Run verification in a goroutine
	go func() {
		defer func() {
			verifyBtn.Synchronize(func() {
				verifyBtn.SetEnabled(true)
			})
		}()
		
		// Create config for the ISO file
		config := &Config{
			Path:     isoPath,
			NoVerify: false,
			MD5Check: md5Check,
		}
		
		// Validate path
		if err := validatePath(config); err != nil {
			resultText.Synchronize(func() {
				resultText.AppendText(fmt.Sprintf("Error: %v\n", err))
			})
			return
		}
		
		output := &strings.Builder{}
		output.WriteString(fmt.Sprintf("=== Verifying ISO File ===\n"))
		output.WriteString(fmt.Sprintf("File: %s\n\n", filepath.Base(isoPath)))
		
		// Display SHA256 Hash
		output.WriteString("--- SHA256 Hash ---\n")
		calculatedHash, err := getSha256FromPath(config)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error calculating hash: %v\n", err))
		} else {
			output.WriteString(fmt.Sprintf("SHA256: %s\n", strings.ToLower(calculatedHash)))
		}
		output.WriteString("\n")
		
		// Try MD5 check if requested
		if md5Check {
			output.WriteString("--- Checking for Implanted MD5 ---\n")
			md5Result, err := checkImplantedMD5(config)
			if err != nil {
				output.WriteString(fmt.Sprintf("Error: %v\n", err))
			} else if md5Result == nil {
				output.WriteString("No implanted MD5 signature found.\n")
			} else {
				output.WriteString(fmt.Sprintf("Verification Method: %s\n", md5Result.VerificationMethod))
				output.WriteString(fmt.Sprintf("Stored MD5:          %s\n", md5Result.StoredMD5))
				output.WriteString(fmt.Sprintf("Calculated MD5:      %s\n", md5Result.CalculatedMD5))
				if md5Result.IsIntegrityOK {
					output.WriteString("Result: SUCCESS - Implanted MD5 is valid.\n")
				} else {
					output.WriteString("Result: FAILURE - Implanted MD5 does not match.\n")
				}
			}
			output.WriteString("\n")
		}
		
		output.WriteString("--- Summary ---\n")
		output.WriteString("ISO file verification complete.\n")
		output.WriteString("\nNote: Content verification requires the ISO to be mounted.\n")
		output.WriteString("To verify file contents, mount the ISO and select the drive from the dropdown.")
		
		resultText.Synchronize(func() {
			resultText.SetText(output.String())
		})
	}()
}

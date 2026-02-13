//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// runGUI starts the GUI mode using Fyne
func runGUI() {
	logDebug("runGUI() called (Fyne version)")

	// Create Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow(fmt.Sprintf("chkiso - ISO/Drive Verification Tool v%s", VERSION))
	
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
	
	// If no drives found, add a placeholder message
	if len(drives) == 0 {
		drives = []string{"<No CD-ROM drives found>"}
		defaultIndex = 0
		logDebug("No CD-ROM drives found, using placeholder")
	}
	
	// Check if checkisomd5.exe is available
	logDebug("Checking for checkisomd5.exe...")
	md5Available := isCheckisomd5Available()
	logDebug("checkisomd5.exe available: %v", md5Available)
	
	// Create widgets
	driveLabel := widget.NewLabel("Select Drive:")
	driveSelect := widget.NewSelect(drives, nil)
	driveSelect.SetSelectedIndex(defaultIndex)
	
	resultText := widget.NewMultiLineEntry()
	resultText.Wrapping = fyne.TextWrapWord
	resultText.Disable() // Read-only
	
	var md5Check *widget.Check
	if md5Available {
		md5Check = widget.NewCheck("Verify implanted MD5 (checkisomd5)", nil)
	}
	
	// Verify button
	verifyBtn := widget.NewButton("Verify Drive", func() {
		selectedDrive := driveSelect.Selected
		if selectedDrive == "<No CD-ROM drives found>" {
			resultText.SetText("Error: No CD-ROM drives available.\n\nPlease insert a disc or use 'Browse for ISO file...' button.")
			return
		}
		
		md5CheckEnabled := false
		if md5Check != nil {
			md5CheckEnabled = md5Check.Checked
		}
		
		verifyBtn.Disable()
		resultText.SetText("Starting verification...\n")
		
		go func() {
			// Check if drive is ready
			if !isDriveReady(selectedDrive) {
				resultText.SetText(fmt.Sprintf("Drive %s is detected but appears to be empty.\n\n"+
					"Please:\n"+
					"  • Insert a disc into the drive, or\n"+
					"  • Use 'Browse for ISO file...' button\n", selectedDrive))
				verifyBtn.Enable()
				return
			}
			
			// Perform verification
			output := captureVerificationOutput(selectedDrive, md5CheckEnabled)
			resultText.SetText(output)
			verifyBtn.Enable()
		}()
	})
	
	// Browse button
	browseBtn := widget.NewButton("Browse for ISO file...", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if reader == nil {
				return // User cancelled
			}
			
			filePath := reader.URI().Path()
			reader.Close()
			
			// Check if it's an ISO file
			ext := strings.ToLower(filepath.Ext(filePath))
			if ext != ".iso" {
				resultText.SetText(fmt.Sprintf("Error: Only ISO files are supported.\n\nYou selected: %s\n\nPlease select an ISO file (.iso extension).", filepath.Base(filePath)))
				return
			}
			
			md5CheckEnabled := false
			if md5Check != nil {
				md5CheckEnabled = md5Check.Checked
			}
			
			verifyBtn.Disable()
			browseBtn.Disable()
			resultText.SetText(fmt.Sprintf("Verifying: %s\n\nPlease wait...\n", filepath.Base(filePath)))
			
			go func() {
				output := captureVerificationOutput(filePath, md5CheckEnabled)
				resultText.SetText(output)
				verifyBtn.Enable()
				browseBtn.Enable()
			}()
		}, myWindow)
		
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".iso"}))
		fd.Show()
	})
	
	// Close button
	closeBtn := widget.NewButton("Close", func() {
		myApp.Quit()
	})
	
	// Set initial message
	if len(drives) == 1 && drives[0] == "<No CD-ROM drives found>" {
		logDebug("Setting initial message for no drives found")
		resultText.SetText("No CD-ROM drives detected on this system.\n\n" +
			"To verify an ISO file:\n" +
			"  • Click 'Browse for ISO file...' button below\n\n" +
			"To verify a CD/DVD drive:\n" +
			"  1. Insert a bootable CD/DVD into a drive\n" +
			"  2. Or mount an ISO file using Windows Explorer (right-click → Mount)\n" +
			"  3. Then relaunch this application\n\n" +
			"Command-line usage is also available:\n" +
			"  chkiso.exe path\\to\\image.iso")
		verifyBtn.Disable()
	} else {
		logDebug("Setting initial ready message")
		resultText.SetText("Ready to verify.\n\n" +
			"Select a drive from the dropdown and click 'Verify Drive',\n" +
			"or click 'Browse for ISO file...'.")
	}
	
	// Build layout
	content := container.NewVBox(
		driveLabel,
		driveSelect,
		verifyBtn,
		browseBtn,
	)
	
	if md5Check != nil {
		content.Add(md5Check)
	}
	
	content.Add(container.NewScroll(resultText))
	content.Add(closeBtn)
	
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(700, 500))
	
	logDebug("Showing Fyne window...")
	myWindow.ShowAndRun()
	logDebug("Fyne window closed")
}

// captureVerificationOutput runs verification and captures output
func captureVerificationOutput(target string, md5Check bool) string {
	var output strings.Builder
	
	// Save original stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	
	// Create a pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Sprintf("Error creating pipe: %v\n", err)
	}
	
	os.Stdout = w
	os.Stderr = w
	
	// Channel to capture the output
	done := make(chan string)
	go func() {
		var buf strings.Builder
		buffer := make([]byte, 4096)
		for {
			n, err := r.Read(buffer)
			if n > 0 {
				buf.Write(buffer[:n])
			}
			if err != nil {
				break
			}
		}
		done <- buf.String()
	}()
	
	// Run verification
	config := Config{
		IsoPath:   target,
		Md5Check:  md5Check,
		NoVerify:  false,
		GuiMode:   false,
	}
	
	performVerification(config)
	
	// Restore stdout/stderr
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	
	// Get captured output
	output.WriteString(<-done)
	
	return output.String()
}

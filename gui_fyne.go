//go:build windows && !arm64
// +build windows,!arm64

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleCP = kernel32.NewProc("GetConsoleCP")
)

// hasConsole checks if the process is attached to a console window
func hasConsole() bool {
	r1, _, _ := procGetConsoleCP.Call()
	return r1 != 0
}

// runGUI starts the GUI mode using Fyne
func runGUI() {
	logDebug("runGUI() called (Fyne version)")

	// Create Fyne application with unique ID
	myApp := app.NewWithID("com.github.pappasjfed.chkiso")
	myWindow := myApp.NewWindow(fmt.Sprintf("chkiso - ISO/Drive Verification Tool v%s", VERSION))
	
	// Set application icon
	myWindow.SetIcon(GetAppIcon())
	
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
	
	// Create result text with monospace font for better readability
	resultText := widget.NewMultiLineEntry()
	resultText.Wrapping = fyne.TextWrapWord
	resultText.Disable() // Read-only
	// Use monospace font for better output formatting
	resultText.TextStyle = fyne.TextStyle{Monospace: true}
	
	var md5Check *widget.Check
	if md5Available {
		md5Check = widget.NewCheck("Verify implanted MD5 (checkisomd5)", nil)
	}
	
	// Declare buttons before use in closures
	var verifyBtn *widget.Button
	var browseBtn *widget.Button
	
	// Verify button
	verifyBtn = widget.NewButton("Verify Drive", func() {
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
		browseBtn.Disable()
		resultText.SetText("Starting verification...\n")
		
		go func() {
			// Check if drive is ready
			if !isDriveReady(selectedDrive) {
				fyne.Do(func() {
					resultText.SetText(fmt.Sprintf("Drive %s is detected but appears to be empty.\n\n"+
						"Please:\n"+
						"  • Insert a disc into the drive, or\n"+
						"  • Use 'Browse for ISO file...' button\n", selectedDrive))
					verifyBtn.Enable()
					browseBtn.Enable()
				})
				return
			}
			
			// Show progress
			fyne.Do(func() {
				resultText.SetText(fmt.Sprintf("Verifying drive %s...\n\nStep 1/3: Reading ISO structure...\n", selectedDrive))
			})
			
			// Perform verification
			output := captureVerificationOutput(selectedDrive, md5CheckEnabled)
			
			fyne.Do(func() {
				resultText.SetText(output)
				verifyBtn.Enable()
				browseBtn.Enable()
			})
		}()
	})
	
	// Browse button
	browseBtn = widget.NewButton("Browse for ISO file...", func() {
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
			resultText.SetText(fmt.Sprintf("Verifying: %s\n\nStep 1/3: Reading ISO structure...\n", filepath.Base(filePath)))
			
			go func() {
				output := captureVerificationOutput(filePath, md5CheckEnabled)
				fyne.Do(func() {
					resultText.SetText(output)
					verifyBtn.Enable()
					browseBtn.Enable()
				})
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
	
	// Build layout with better sizing
	content := container.NewBorder(
		// Top: controls
		container.NewVBox(
			driveLabel,
			driveSelect,
			container.NewGridWithColumns(2, verifyBtn, browseBtn),
			func() fyne.CanvasObject {
				if md5Check != nil {
					return md5Check
				}
				return widget.NewLabel("") // Empty placeholder
			}(),
		),
		// Bottom: close button
		container.NewVBox(
			widget.NewSeparator(),
			closeBtn,
		),
		// Left, Right: nil
		nil, nil,
		// Center: results (takes remaining space)
		container.NewScroll(resultText),
	)
	
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	
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
		Path:      target,
		MD5Check:  md5Check,
		NoVerify:  false,
		GuiMode:   true, // Enable more verbose output for GUI
	}
	
	// Perform all verification steps
	if err := validatePath(&config); err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
	} else {
		// Display SHA256 hash
		displaySha256Hash(&config)
		
		// Verify contents
		verifyContents(&config)
		
		// Verify MD5 if requested
		if md5Check {
			verifyImplantedMD5(&config)
		}
	}
	
	// Restore stdout/stderr
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	
	// Get captured output
	output.WriteString(<-done)
	
	return output.String()
}

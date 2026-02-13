# GUI Testing Guide

This document provides instructions for testing the GUI mode on Windows.

## ⚠️ Important Note

**GUI mode may not work on all Windows systems** due to Windows tooltip control limits in the walk library. If you encounter a "TTM_ADDTOOL failed" error:
- This is a known limitation of the walk GUI library
- The error dialog will provide clear instructions to use CLI mode instead
- CLI mode has all the same features as GUI mode
- See the main README for CLI usage examples

## Testing the GUI Mode

### Prerequisites
- Windows 10 or Windows 11
- A CD/DVD drive with bootable media (e.g., Linux installation disc) (optional)
- The `chkiso-windows-amd64.exe` binary

### Test Cases

#### Test 1: Launch GUI from File Explorer
1. Navigate to the directory containing `chkiso-windows-amd64.exe` in File Explorer
2. **Double-click** the executable
3. **Possible Results**:
   - **Success**: A GUI window appears (see below for layout)
   - **Failure**: Error dialog appears with "TTM_ADDTOOL failed" and CLI instructions - this is a known limitation
4. **If GUI opens successfully**, it should show:
   - Title: "chkiso - ISO/Drive Verification Tool v2.0.0"
   - A "Select Drive:" label at the top
   - A dropdown menu listing all CD-ROM/DVD drives (or "<No CD-ROM drives found>" if none present)
   - A "Verify" button below the dropdown
   - A "Browse for ISO file..." button
   - **MD5 checkbox** (only if checkisomd5.exe is available): "Verify implanted MD5 (checkisomd5)"
   - A large text area for results (takes most of the window)
   - A "Close" button at the bottom
   - Note: Controls are arranged vertically to avoid Windows tooltip limits

#### Test 1b: Launch GUI from Command Line with -gui Flag
1. Open Command Prompt or PowerShell
2. Run: `chkiso-windows-amd64.exe -gui`
3. **Expected Result**: 
   - GUI window launches (same as double-clicking)
   - All components visible as described in Test 1
   - Works even when run from console

#### Test 1c: Launch GUI with No Drives Detected
1. Launch the GUI on a system with no CD-ROM drives
2. **Expected Result**: 
   - The GUI window should open (not disappear immediately)
   - The dropdown shows "<No CD-ROM drives found>"
   - The text area displays a helpful error message with instructions
   - The "Verify" button is disabled
   - The "Browse for ISO file..." button is available and functional
   - User can click "Browse for ISO file..." to select an ISO file for verification

#### Test 1d: MD5 Checkbox Visibility
1. Check if checkisomd5.exe is in PATH or same directory as chkiso.exe
2. Launch the GUI
3. **Expected Result if checkisomd5.exe is present**:
   - MD5 checkbox appears with text "Verify implanted MD5 (checkisomd5)"
   - Checkbox is unchecked by default
   - Can be checked/unchecked
4. **Expected Result if checkisomd5.exe is NOT present**:
   - No MD5 checkbox appears
   - All other GUI elements work normally

#### Test 2: Drive Selection and Default
1. Launch the GUI as in Test 1
2. Check the dropdown menu
3. **Expected Result**: 
   - All CD-ROM/DVD drives should be listed (e.g., "D:", "E:")
   - If the executable is running from a CD-ROM drive, that drive should be pre-selected

#### Test 3: Verify a Drive
1. Launch the GUI
2. Ensure a bootable disc is in one of the drives
3. Select the drive from the dropdown
4. Click "Verify"
5. **Expected Result**:
   - The text area should immediately show "Verifying drive X:... Please wait..."
   - The "Verify" button should be disabled during verification
   - After a few moments/minutes, results should appear showing:
     - SHA256 hash of the drive
     - List of checksum files found
     - Verification results for each file
     - Summary of verification (Success or Failure)
   - The "Verify" button should be re-enabled when complete

#### Test 3b: Verify a Drive with MD5 Checkbox
1. Launch the GUI (ensure checkisomd5.exe is available)
2. Select a drive with a disc that has implanted MD5
3. **Check** the "Verify implanted MD5 (checkisomd5)" checkbox
4. Click "Verify"
5. **Expected Result**:
   - Verification proceeds as normal
   - Results include an additional "Verifying Implanted MD5" section
   - Shows stored MD5, calculated MD5, and success/failure
   - If no implanted MD5: Shows "No implanted MD5 signature found"

#### Test 4: Verify Empty Drive
1. Launch the GUI with a CD-ROM drive that has no disc
2. Select the empty drive from the dropdown
3. Click "Verify"
4. **Expected Result**:
   - Text area displays message: "Drive X: is detected but empty"
   - Provides instructions to insert disc or use browse button
   - Verify button is re-enabled
   - No crash or error dialog

#### Test 5: Browse for ISO File
1. Launch the GUI
2. Click "Browse for ISO file..." button
3. **Expected Result**: 
   - A file dialog should appear
   - Filter should show "ISO Files (*.iso)" by default
4. Select an ISO file and click Open
5. **Expected Result**:
   - The text area should show "Verifying ISO file: [filename]..."
   - Verification should start automatically
   - Results should appear showing:
     - File name
     - SHA256 hash
     - Implanted MD5 check (if present)
     - Summary message
   - The "Verify" button should be re-enabled when complete
6. Cancel the file dialog
7. **Expected Result**: Nothing happens, window stays open

#### Test 6: Drag and Drop ISO File - TEMPORARILY DISABLED
**Note**: Drag-and-drop functionality has been temporarily disabled to resolve Windows tooltip control errors (TTM_ADDTOOL). Use "Browse for ISO file..." button instead.

~~1. Launch the GUI~~
~~2. Open File Explorer and navigate to an ISO file~~
~~3. Drag the ISO file and drop it onto the GUI window~~
~~4. **Expected Result**: Verification should start~~

#### Test 7: Close the GUI
1. After verification completes (or at any time)
2. Click the "Close" button
3. **Expected Result**: The window should close immediately

#### Test 8: Launch from Command Prompt
1. Open Command Prompt (cmd.exe)
2. Navigate to the directory with the executable
3. Run: `chkiso-windows-amd64.exe` (with no arguments)
4. **Expected Result**: The program should display usage/help text in the console (NOT launch GUI)

#### Test 9: Launch from PowerShell
1. Open PowerShell
2. Navigate to the directory with the executable
3. Run: `.\chkiso-windows-amd64.exe` (with no arguments)
4. **Expected Result**: The program should display usage/help text in the console (NOT launch GUI)

#### Test 10: CLI Mode with Arguments
1. Open Command Prompt or PowerShell
2. Run: `chkiso-windows-amd64.exe --version`
3. **Expected Result**: Should display version information in the console
4. Run: `chkiso-windows-amd64.exe E:` (use actual drive letter)
5. **Expected Result**: Should perform verification in CLI mode, showing output in the console

#### Test 11: Debug Logging
1. Open Command Prompt or PowerShell
2. Run: `chkiso-windows-amd64.exe -gui`
3. **Expected Result**:
   - Console shows: "Debug log: C:\Users\...\AppData\Local\Temp\chkiso-debug-*.log"
   - GUI window appears (or error dialog with log path)
4. Navigate to the temp directory shown
5. Open the log file
6. **Expected Result**:
   - Log contains version, platform info
   - Log shows drive detection results
   - Log shows window creation steps
   - Timestamps for each entry
7. If GUI fails to create, error dialog shows log path

## Known Behavior

### Debug Logging (NEW)
- GUI mode automatically creates debug logs in temp directory
- Log path shown on stderr/console when launching
- Logs persist after program exits
- Useful for troubleshooting GUI issues

### GUI vs CLI Detection
The program uses the following logic to determine mode:
- **GUI Mode**: Launched by double-clicking from File Explorer (no console attached)
- **CLI Mode**: Launched from PowerShell/CMD, or launched with command-line arguments

### Drive Detection
- Only CD-ROM/DVD drives (drive type 5) are listed in the dropdown
- Hard drives, SSDs, and USB drives are excluded from the GUI
- This is intentional for safety and to focus on the primary use case
- Empty drives (no disc inserted) are detected and handled with helpful messages

### Multiple Verification Methods
The GUI supports two ways to verify ISO files:
1. **Drive dropdown**: Select a CD-ROM/DVD drive and click "Verify"
2. **Browse button**: Click "Browse for ISO file..." to select an ISO via file dialog
~~3. **Drag and drop**: Temporarily disabled to prevent Windows tooltip errors~~

All methods provide the same comprehensive verification (SHA256 + MD5 + file contents when applicable)

## Troubleshooting

### Debug Logging (NEW)
- **GUI mode automatically creates debug logs** in your temp directory
- Log location: `%TEMP%\chkiso-debug-YYYYMMDD-HHMMSS.log`
- The path is displayed when launching GUI mode (check console if running from CLI)
- Logs include:
  - Version and platform information
  - Drive detection results
  - Window creation steps
  - Any errors that occur
- **Include this log file when reporting issues**

### No drives appear in dropdown
- **This is now handled gracefully**: The GUI will display "<No CD-ROM drives found>"
- The window will stay open with helpful instructions
- Use the "Browse for ISO file..." button to verify ISO files from your hard drive
- Or insert a CD/DVD or mount an ISO, then relaunch the application

### Window disappears immediately (FIXED)
- **Previous issue**: Window would close instantly if no drives were found
- **Now fixed**: Window stays open with error message and browse option
- If you still experience this issue, please report with details

### GUI doesn't launch when double-clicking
- Check that you're running on Windows
- Ensure you're double-clicking from File Explorer (not running from a console)
- Check Windows Event Viewer for any error messages
- Try launching with `-gui` flag from command line to see error messages

### GUI creation errors (e.g., TTM_ADDTOOL failed) - KNOWN LIMITATION
- **Status**: **Not fully fixable** - This is a walk library limitation
- **Issue**: Some Windows systems hit tooltip control limits when creating GUI
- **What we tried**:
  - Replaced MainWindow with Dialog (simpler structure)
  - Disabled tooltips on all widgets (`ToolTipText: ""`)
  - Removed OnDropFiles and drag-drop features
  - Flattened widget hierarchy
- **Result**: Improved but not guaranteed to work on all systems
- **If you encounter this error**:
  - This is expected on some Windows configurations
  - The error dialog will provide clear CLI usage instructions
  - **Use CLI mode instead** - it has all the same features:
    ```
    chkiso.exe E:\           # Verify drive
    chkiso.exe file.iso      # Verify ISO
    chkiso.exe file.iso -md5 # With MD5 check
    ```
  - Check the debug log for technical details
  - No need to report - this is a known library limitation

### GUI freezes during verification
- This should not happen as verification runs asynchronously
- If it does freeze, this is a bug - please report with details

### Verification takes a long time
- This is normal for large drives or slow drives
- Calculating SHA256 for an entire drive can take several minutes
- Content verification depends on the number of files and their sizes

## Reporting Issues

If you encounter any issues:
1. Note the Windows version (Windows 10/11)
2. Note the drive type and size
3. Capture any error messages
4. **Include the debug log file** from `%TEMP%\chkiso-debug-*.log`
4. If possible, provide screenshots
5. Report on the GitHub issues page

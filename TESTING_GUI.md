# GUI Testing Guide

This document provides instructions for testing the new GUI mode on Windows.

## Testing the GUI Mode

### Prerequisites
- Windows 10 or Windows 11
- A CD/DVD drive with bootable media (e.g., Linux installation disc)
- The `chkiso-windows-amd64.exe` binary

### Test Cases

#### Test 1: Launch GUI from File Explorer
1. Navigate to the directory containing `chkiso-windows-amd64.exe` in File Explorer
2. **Double-click** the executable
3. **Expected Result**: A GUI window should appear with:
   - Title: "chkiso - ISO/Drive Verification Tool v2.0.0"
   - A "Select Drive:" label
   - A dropdown menu listing all CD-ROM/DVD drives
   - A "Verify" button
   - An empty text area for results
   - A "Close" button at the bottom

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

#### Test 4: Close the GUI
1. After verification completes (or at any time)
2. Click the "Close" button
3. **Expected Result**: The window should close immediately

#### Test 5: Launch from Command Prompt
1. Open Command Prompt (cmd.exe)
2. Navigate to the directory with the executable
3. Run: `chkiso-windows-amd64.exe` (with no arguments)
4. **Expected Result**: The program should display usage/help text in the console (NOT launch GUI)

#### Test 6: Launch from PowerShell
1. Open PowerShell
2. Navigate to the directory with the executable
3. Run: `.\chkiso-windows-amd64.exe` (with no arguments)
4. **Expected Result**: The program should display usage/help text in the console (NOT launch GUI)

#### Test 7: CLI Mode with Arguments
1. Open Command Prompt or PowerShell
2. Run: `chkiso-windows-amd64.exe --version`
3. **Expected Result**: Should display version information in the console
4. Run: `chkiso-windows-amd64.exe E:` (use actual drive letter)
5. **Expected Result**: Should perform verification in CLI mode, showing output in the console

## Known Behavior

### GUI vs CLI Detection
The program uses the following logic to determine mode:
- **GUI Mode**: Launched by double-clicking from File Explorer (no console attached)
- **CLI Mode**: Launched from PowerShell/CMD, or launched with command-line arguments

### Drive Detection
- Only CD-ROM/DVD drives (drive type 5) are listed in the dropdown
- Hard drives, SSDs, and USB drives are excluded from the GUI
- This is intentional for safety and to focus on the primary use case

## Troubleshooting

### No drives appear in dropdown
- Ensure you have a physical or virtual CD-ROM drive
- The drive must be recognized by Windows as a CD-ROM drive (type 5)
- USB drives and hard drives will NOT appear in the list

### GUI doesn't launch when double-clicking
- Check that you're running on Windows
- Ensure you're double-clicking from File Explorer (not running from a console)
- Check Windows Event Viewer for any error messages

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
4. If possible, provide screenshots
5. Report on the GitHub issues page

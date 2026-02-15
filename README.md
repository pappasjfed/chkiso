# chkiso - ISO/Drive Verification Tool

[![Build Go Binary](https://github.com/pappasjfed/chkiso/actions/workflows/build-go.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/build-go.yml)
[![Release](https://github.com/pappasjfed/chkiso/actions/workflows/release.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/release.yml)

A cross-platform tool for validating ISO images and optical media. Written in Go for maximum portability and reliability.

## Features

- ✅ **Cross-platform**: Works on Windows, Linux, macOS, and FreeBSD
- ✅ **No FIPS restrictions**: MD5 hashing works everywhere (no policy blocks)
- ✅ **Single executable**: Statically-linked binary with no dependencies
- ✅ **GUI Mode (Windows)**: Easy-to-use graphical interface for non-technical users
- ✅ **Multiple verification methods**:
  - SHA256 hash verification
  - MD5 implanted hash check (checkisomd5 compatible)
  - External hash file verification
  - Content verification against embedded checksums

## GUI Mode (Windows & Linux)

**New!** chkiso now includes a modern graphical user interface using Fyne. Perfect for non-technical users who want to verify optical media without using the command line.

**✅ Reliable**: The new Fyne-based GUI works on all Windows systems without the tooltip control issues of the previous version.

**⚠️ Platform Note**: 
- **Windows AMD64**: Full GUI support with Fyne
- **Windows ARM64**: CLI-only (GUI requires OpenGL which isn't available for ARM64 cross-compilation)
- **Linux**: GUI available (requires display server)

### How to Use the GUI

**Latest Improvements (v2.0.0)**:
- 🎨 Custom gold CD icon (runtime and optionally for EXE)
- 📐 Larger window (800x600) with better layout
- 🔠 Bold monospace font for clearer output
- 📊 Progress indicators during verification
- 🔒 Prevents concurrent operations
- 📝 More verbose output (especially with checkisomd5.exe)
- 🐛 Fixed double-click detection from Explorer
- 🎯 Close button moved to bottom for better UX
- 📄 Debug log path shown in app (no second window)
- ✅ checkisomd5.exe checkbox actually runs the tool with -v flag

1. **Launch the GUI:**
   - **Windows AMD64**: Double-click `chkiso-windows-amd64.exe` (automatically launches GUI if no arguments)
   - **Windows ARM64**: Use CLI mode (see Command Line Usage section)
   - **Linux**: Run `./chkiso -gui` (GUI available as bonus feature!)
   - **Command line**: `chkiso.exe -gui` (explicitly launch GUI mode)
2. A modern window will appear with:
   - **Select Drive** dropdown - Lists all CD-ROM/DVD drives on your system
   - **Verify Drive** button - Starts drive verification
   - **Browse for ISO file...** button - Opens file picker to verify ISO files
   - **MD5 checkbox** (if checkisomd5.exe is available) - "Verify implanted MD5 (checkisomd5)"
   - **Scrollable results area** - Shows verification progress and results
   - **Close** button - Exit the application
3. **To verify a CD/DVD drive:**
   - Select the drive you want to verify from the dropdown
   - If you run chkiso from a CD/DVD drive, that drive will be pre-selected
   - Optionally check the MD5 checkbox if available
   - Click "Verify Drive" to start the verification process
   - **Note**: If a drive is empty, you'll get a helpful message
4. **To verify an ISO file:**
   - Click "Browse for ISO file..." and select the ISO from the file picker
   - The MD5 checkbox (if available) applies to ISO verification too
   - Verification starts automatically after selecting a file
5. Wait for the verification to complete (this may take several minutes)
6. Review the results in the scrollable text area
7. Click "Close" when finished

**Note**: If no CD-ROM drives are detected, the GUI will still open with a helpful message. You can use the browse button to verify ISO files.

### MD5 Verification in GUI

The GUI includes an optional MD5 checkbox if the external `checkisomd5` tool is available:

**Windows**: Looks for `checkisomd5.exe`
**Linux/macOS/FreeBSD**: Looks for `checkisomd5`

The checkbox only appears if the tool is found in:
1. System PATH
2. Same directory as chkiso executable

When checked:
- Runs the external checkisomd5 tool with `-v` (verbose) flag
- Shows detailed progress and verification output
- Applies to both drive and ISO file verification
- Falls back to internal implementation if tool fails

**Installing checkisomd5**:
```bash
# Fedora/RHEL/CentOS
sudo dnf install isomd5sum

# Debian/Ubuntu
sudo apt-get install isomd5sum

# macOS (Homebrew)
brew install isomd5sum

# Or download and place in same directory as chkiso
```

### GUI vs Command Line

The program automatically detects how it's being run:
- **GUI Mode**: When double-clicked from File Explorer (no console attached) OR when run with `-gui` flag
- **Command-Line Mode**: When run from PowerShell, Command Prompt, or with arguments

This means you can use the same executable for both GUI and command-line operations!

**To explicitly launch GUI from command line:**
```
chkiso.exe -gui
```

### Troubleshooting GUI Mode

If you encounter errors when launching the GUI:

1. **Debug Logging**: GUI mode automatically creates a debug log file in your temp directory
   - Location shown at top of results area: `Debug log: C:\...\chkiso-debug-....log`
   - Check this file for detailed error information
   - No annoying second window!

2. **Common Issues**:
   - **Display issues**: Fyne requires OpenGL support
     - Ensure your graphics drivers are up to date
     - Windows amd64 builds include GUI support
     - Windows arm64 builds are CLI-only
   - **Window creation failures**: May be due to system resource constraints
   - The debug log will show the error details
   
   **Technical Notes**: 
   - GUI uses Fyne framework (modern, cross-platform)
   - No Windows tooltip control limits (previous walk library issue resolved!)
   - Requires OpenGL for rendering
   - Linux GUI support available as bonus feature

3. **If GUI Doesn't Work**:
   - All GUI features are available in CLI mode:
     - Drive verification: `chkiso.exe E:\`
     - ISO verification: `chkiso.exe path\to\file.iso`
     - MD5 checking: `chkiso.exe file.iso -md5`
   - See Command-Line Usage section below for full details

4. **For Developers/Issues**:
   - Check the debug log file for details
   - Include the log file when reporting issues
   - Try running as administrator if permissions are an issue

## Installation

### Download Pre-built Binary (Recommended)

Download the appropriate binary for your platform from the [Releases](https://github.com/pappasjfed/chkiso/releases) page:

- **Windows**: `chkiso-windows-amd64.exe` (64-bit) or `chkiso-windows-arm64.exe` (ARM64)
  - Note: Windows 11 and modern Windows 10 are 64-bit only
  - 32-bit Windows builds discontinued (Windows 10 32-bit reached end-of-life)
- **Linux**: `chkiso-linux-amd64` (64-bit), `chkiso-linux-arm64` (ARM 64-bit), `chkiso-linux-arm` (ARM 32-bit), or `chkiso-linux-386` (32-bit)
- **macOS**: `chkiso-darwin-amd64` (Intel) or `chkiso-darwin-arm64` (Apple Silicon)
- **FreeBSD**: `chkiso-freebsd-amd64`

On Linux/macOS/FreeBSD, make the binary executable:
```bash
chmod +x chkiso-*
```

### Build from Source

Requirements: [Go 1.21+](https://golang.org/dl/)

```bash
git clone https://github.com/pappasjfed/chkiso.git
cd chkiso
go build -o chkiso
```

Or use the Makefile for cross-platform builds:
```bash
make build           # Build for current platform
make build-all       # Build for all platforms
make windows         # Build for Windows
make linux           # Build for Linux
make macos           # Build for macOS
```

## Usage

### Basic Usage

By default, chkiso displays the SHA256 hash of the ISO/drive and verifies internal file integrity:

```bash
chkiso path/to/image.iso
```

### Advanced Options

#### Verify against an expected SHA256 hash:

```bash
# Positional argument
chkiso image.iso <sha256-hash>

# Or using flag
chkiso -sha256 <sha256-hash> image.iso
```

**Note**: Content verification runs by default. Use `-noverify` to skip it.

#### Verify against a hash file:

```bash
chkiso image.iso -shafile path/to/hashfile.sha
```

#### Check implanted MD5 hash:

```bash
chkiso image.iso -md5
```

**Advantage**: No FIPS restrictions! Works on all systems regardless of security policies.

**External Tool Support**: If the `checkisomd5` tool is available (on Windows: `checkisomd5.exe`), chkiso will automatically use it with the `-v` (verbose) flag for more detailed output. Otherwise, it uses the internal MD5 implementation.

**Installing checkisomd5** (optional but provides more verbose output):
```bash
# Fedora/RHEL/CentOS
sudo dnf install isomd5sum

# Debian/Ubuntu  
sudo apt-get install isomd5sum

# macOS (Homebrew)
brew install isomd5sum
```

**Note for Windows**: Implanted MD5 check requires direct ISO file access. If you have a mounted ISO (e.g., drive H:), use the original ISO file path instead:
```bash
# This works
chkiso C:\path\to\image.iso -md5

# This will skip MD5 check (mounted ISOs don't support device-level access)
chkiso H: -md5
```
Content verification will still work fine with mounted drives.

#### Skip internal file verification:

```bash
chkiso image.iso -noverify
```

### Content Verification

By default, chkiso performs **content verification** when verifying drives (e.g., `chkiso E:`). This feature:

- **Recursively searches** for ALL checksum files on the media:
  - Files ending with `.sha` (e.g., `files.sha`, `docs.sha`, `packages.sha`)
  - Files named `sha256sum.txt` or `SHA256SUMS`
- **Processes each checksum file** found in any directory or subdirectory
- **Validates all files** referenced in each checksum file
- **Reports comprehensive results** showing which checksum files were found and processed

This ensures that if your media contains multiple checksum files in different directories (common for complex distributions or multi-component media), ALL of them will be found and verified automatically.

**Example output:**
```
--- Verifying Contents ---
Searching for checksum files (*.sha, sha256sum.txt, SHA256SUMS) in E:\...

Found 3 checksum file(s):
  1. main.sha
  2. docs/docs.sha
  3. software/packages.sha

Processing checksum file: main.sha
Verifying: readme.txt -> OK

Processing checksum file: docs.sha
Verifying: manual.pdf -> OK

Processing checksum file: packages.sha
Verifying: installer.exe -> OK

--- Verification Summary ---
Checksum files processed: 3
Total files verified: 3
Success: All 3 files verified successfully.
```

Use `-noverify` to skip content verification if you only want to check the ISO hash or implanted MD5.

### Automatic ISO Mounting (Windows)

**New Feature!** On Windows, chkiso now automatically mounts ISO files for content verification and unmounts them when done.

```bash
# Automatically mounts, verifies, and unmounts
chkiso ubuntu-22.04.iso

# The tool will:
# 1. Mount the ISO to a drive letter (e.g., H:)
# 2. Verify all checksum files found on the ISO
# 3. Automatically unmount the ISO when done
```

**How it works:**
- Uses PowerShell's `Mount-DiskImage` to mount ISOs
- Finds all checksum files (*.sha, sha256sum.txt, SHA256SUMS) automatically
- Verifies all files referenced in the checksum files
- Cleans up by unmounting the ISO automatically

**Fallback:**
If automatic mounting fails (requires admin privileges or other issues), the tool will display instructions for manual mounting.

**Manual control:**
If you prefer to mount manually, you can still use the drive letter:
```bash
# Mount manually first
Mount-DiskImage -ImagePath C:\path\to\image.iso

# Then verify using drive letter
chkiso H:

# Manually dismount when done
Dismount-DiskImage -ImagePath C:\path\to\image.iso
```

#### Verify a drive (Windows):

```bash
chkiso E:
```

#### All options:

```
Options:
  -sha256 <hash>      Expected SHA256 hash for verification
  -sha256sum <hash>   Alias for -sha256
  -sha <hash>         Alias for -sha256
  -shafile <file>     Path to SHA256 hash file
  -noverify           Skip verifying internal file hashes
  -md5                Enable implanted MD5 check
  -dismount           Dismount/eject after verification
  -eject              Alias for -dismount
  -version            Display version information
  -help               Display help information
```

### Examples

```bash
# Display SHA256 hash
chkiso ubuntu-22.04-desktop-amd64.iso

# Verify against known hash
chkiso ubuntu-22.04-desktop-amd64.iso a4acfda10b18da50e2ec50ccaf860d7f20b389df8765611142305c0e911d16fd

# Verify using hash file
chkiso ubuntu-22.04-desktop-amd64.iso -shafile SHA256SUMS

# Check implanted MD5 and verify contents
chkiso rhel-9.0-x86_64-dvd.iso -md5

# Quick hash check without content verification
chkiso image.iso <hash> -noverify
```

## Building

### Go Binary

The Go binaries are automatically built for multiple platforms via GitHub Actions.

**Manual build:**
```bash
# Current platform
go build -o chkiso

# Specific platform
GOOS=windows GOARCH=amd64 go build -o chkiso-windows-amd64.exe
GOOS=linux GOARCH=amd64 go build -o chkiso-linux-amd64
GOOS=darwin GOARCH=arm64 go build -o chkiso-darwin-arm64
```

**Using Makefile:**
```bash
make build-all    # Build for all platforms
make windows      # Windows binaries
make linux        # Linux binaries
make macos        # macOS binaries
```

## Testing

Tests run automatically on pull requests and pushes to main. The test suite validates the Go implementation against `test/test.iso` using multiple verification methods.

Run tests locally:
```bash
go test -v ./...
```

## Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| Windows  | amd64, arm64 | ✅ Fully supported |
| Linux    | amd64, 386, arm, arm64 | ✅ Fully supported |
| macOS    | amd64 (Intel), arm64 (Apple Silicon) | ✅ Fully supported |
| FreeBSD  | amd64 | ✅ Fully supported |

**Note**: Windows 32-bit (386) builds are no longer provided. Windows 11 only supports 64-bit processors, and Windows 10 32-bit has reached end-of-life. All modern Windows installations are 64-bit.

## Why Go?

This tool is written in Go to address common limitations:

1. **No FIPS Restrictions**: Works on FIPS-compliant systems without MD5 blocks
2. **Universal Compatibility**: Single binary runs anywhere without runtime dependencies
3. **No Execution Policies**: Works on locked-down systems that block scripts
4. **True Cross-platform**: Native support for Windows, Linux, macOS, and BSD
5. **Native Drive Access**: Direct access to device paths on all platforms

## License

MIT License - See LICENSE file for details.

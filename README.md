# What is this?
[![Test](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml)
[![Build Go Binary](https://github.com/pappasjfed/chkiso/actions/workflows/build-go.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/build-go.yml)
[![Build and Release](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml)

This is a project that is used to validate ISO images created by standard DSO pipelines. It checks hashes for ISOs or media.

**New in v2.0**: chkiso is now available in **Go** for true cross-platform support! The Go version:
- ✅ Works on Windows, Linux, macOS, and FreeBSD
- ✅ No FIPS restrictions on MD5 hashing
- ✅ Single statically-linked executable (no runtime dependencies)
- ✅ Can run on locked-down systems that block PowerShell scripts
- ✅ Native drive access support on all platforms
- ✅ Same command-line interface as PowerShell version

The PowerShell version is still maintained for Windows users who prefer it.

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the appropriate binary for your platform from the [Releases](https://github.com/pappasjfed/chkiso/releases) page:

- **Windows**: `chkiso-windows-amd64.exe` (64-bit) or `chkiso-windows-386.exe` (32-bit)
- **Linux**: `chkiso-linux-amd64` (64-bit), `chkiso-linux-arm64` (ARM 64-bit), etc.
- **macOS**: `chkiso-darwin-amd64` (Intel) or `chkiso-darwin-arm64` (Apple Silicon)
- **FreeBSD**: `chkiso-freebsd-amd64`

On Linux/macOS, make the binary executable:
```bash
chmod +x chkiso-linux-amd64
```

### Option 2: Build from Source

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

### Option 3: PowerShell Script (Windows Only)

If you prefer PowerShell or need drive letter support on Windows with ps2exe:

```powershell
.\chkiso.ps1 path\to\image.iso
```

## Usage

The command-line interface is consistent across both Go and PowerShell versions.

### Basic Usage

By default, chkiso:
- Displays the SHA256 hash of the ISO/drive
- Verifies internal file integrity against embedded checksum files (*.sha, sha256sum.txt)

**Go version:**
```bash
chkiso path/to/image.iso
```

**PowerShell version:**
```powershell
.\chkiso.ps1 path\to\image.iso
```

### Advanced Options

#### Check against an expected SHA256 hash:

```bash
# Positional argument
chkiso image.iso <sha256-hash>

# Or using flag
chkiso -sha256 <sha256-hash> image.iso
```

**Note**: Content verification runs by default. Use `-noverify` to skip it.

#### Check against a hash file:

```bash
chkiso image.iso -shafile path/to/hashfile.sha
```

#### Enable implanted MD5 check:

```bash
chkiso image.iso -md5
```

**Go version advantage**: No FIPS restrictions! The Go version can always verify MD5 hashes, unlike PowerShell which may be blocked by FIPS security policies.

#### Skip internal file verification:

```bash
chkiso image.iso -noverify
```

#### Verify a physical drive (Windows):

```bash
# Go version (Windows)
chkiso E:

# PowerShell version (Windows)
.\chkiso.ps1 E:
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

## Why Go Instead of PowerShell?

The original PowerShell version faced several limitations:

1. **FIPS Restrictions**: MD5 hashing is blocked on FIPS-compliant systems
2. **Locked-down Environments**: Many enterprise systems block `.ps1` script execution
3. **Platform Limitations**: PowerShell is primarily Windows-centric
4. **Drive Access**: ps2exe (PowerShell to EXE compiler) cannot access drive letters
5. **Dependencies**: Requires PowerShell runtime

The **Go version solves all these issues**:
- ✅ No FIPS restrictions - MD5 works everywhere
- ✅ Single static binary - runs on locked-down systems
- ✅ True cross-platform - Windows, Linux, macOS, FreeBSD
- ✅ Native drive access - no Win32 API limitations
- ✅ Zero dependencies - just one executable

## Building

### Go Version

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

### PowerShell Version

The PowerShell executable is built using ps2exe:

1. Install ps2exe: `Install-Module -Name ps2exe -Force`
2. Compile: 
   ```powershell
   ps2exe -inputFile chkiso.ps1 -outputFile chkiso-ps.exe `
     -noConsole:$false -title "chkiso" -version "1.0.0.0" `
     -company "chkiso" -product "chkiso" -copyright "MIT License"
   ```

**Important**: The ps2exe version (`chkiso-ps.exe`) cannot access drive letters due to Win32 device path limitations. Use the PowerShell script directly or the Go version for drive letter support.

### Code Signing

The PowerShell executable can be automatically signed if code signing certificates are configured. See [CODE_SIGNING.md](CODE_SIGNING.md) for setup instructions.

## Testing

Tests run automatically on pull requests and pushes to main. The test suite validates both Go and PowerShell versions against `test/test.iso` using multiple verification methods.

Run Go tests locally:
```bash
go test -v ./...
```

## Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| Windows  | amd64, 386, arm64 | ✅ Fully supported |
| Linux    | amd64, 386, arm, arm64 | ✅ Fully supported |
| macOS    | amd64 (Intel), arm64 (Apple Silicon) | ✅ Fully supported |
| FreeBSD  | amd64 | ✅ Fully supported |

## License

MIT License - See LICENSE file for details.

# ISO/Media Verification Tools

[![Test](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml)
[![Build and Release](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml)

Windows utilities for validating ISO images and optical media. These tools are designed for verifying ISOs created by standard DSO pipelines.

## Tools

### sha256media.exe
Calculate SHA256 hash of ISO files or physical drives (CD/DVD/Blu-ray).

**Usage:**
```cmd
sha256media.exe D:\image.iso
sha256media.exe G:
sha256media.exe image.iso -Output image.iso.sha
```

**Features:**
- Hash ISO files
- Hash physical drives using Win32 device paths (PowerShell script mode only)
- Save hash in sha256sum format
- Automatically uses sha256sum.exe if available (Git for Windows, GnuWin32)

### verifymedia.exe
Verify files on mounted ISO or physical media against hash files.

**Usage:**
```cmd
verifymedia.exe D:\image.iso
verifymedia.exe G:
```

**Features:**
- Automatically mounts ISO files
- Scans for all .sha files (also sha256sum.txt, SHA256SUMS)
- Verifies each file against its hash
- Supports Linux sha256sum format: `<hash>  <filename>`
- Reports verification results

## Download

Download the pre-compiled Windows executables from the [Releases](https://github.com/pappasjfed/chkiso/releases) page:
- `sha256media.exe` - Hash calculator
- `verifymedia.exe` - Content verifier

Place them anywhere in your PATH or run from the bin/ directory.

## Requirements

- **Windows only** - These are Windows executables compiled with ps2exe
- Optional: sha256sum.exe (from Git for Windows or GnuWin32) for improved performance

## Examples

### Calculate hash of an ISO
```cmd
sha256media.exe ubuntu-22.04.iso
```

### Verify all files on a CD/DVD drive
```cmd
verifymedia.exe D:
```

### Calculate hash and save to file
```cmd
sha256media.exe image.iso -Output image.iso.sha
```

### Verify ISO contents before burning
```cmd
verifymedia.exe image.iso
```

## Building

The Windows executables are automatically built and attached to releases via GitHub Actions.

### Local Build

To build the executables locally on Windows:

1. Run the build script:
   ```powershell
   .\build.ps1
   ```
   
   This will:
   - Install ps2exe if needed
   - Compile sha256media.ps1 to bin/sha256media.exe
   - Compile verifymedia.ps1 to bin/verifymedia.exe
   - Generate SHA256 checksums for each executable

2. Optional: Clean build
   ```powershell
   .\build.ps1 -Clean
   ```

The compiled executables will be in the `bin/` directory.

### Manual Build

If you prefer to build manually:

1. Install ps2exe: `Install-Module -Name ps2exe -Force`
2. Compile sha256media:
   ```powershell
   ps2exe -inputFile sha256media.ps1 -outputFile bin/sha256media.exe -noConsole:$false -title "sha256media" -version "2.0.0.0"
   ```
3. Compile verifymedia:
   ```powershell
   ps2exe -inputFile verifymedia.ps1 -outputFile bin/verifymedia.exe -noConsole:$false -title "verifymedia" -version "2.0.0.0"
   ```

### Code Signing

The executables can be automatically signed if code signing certificates are configured. See [CODE_SIGNING.md](CODE_SIGNING.md) for setup instructions. If certificates are not configured, the workflow will build unsigned executables.

## Testing

Tests run automatically on pull requests and pushes. The test suite validates the executables against `test/test.iso` using multiple verification methods.

### Test Files

The `test/test.iso` file contains:
- `hashes.sha` - SHA256 hashes in Linux sha256sum format
- Various test files that can be verified

## Development

### Source Files

- `sha256media.ps1` - Source code for sha256media.exe
- `verifymedia.ps1` - Source code for verifymedia.exe
- `build.ps1` - Build script to compile both executables

**Note**: The .ps1 files are source code only and are compiled to .exe files. End users only need the Windows executables.

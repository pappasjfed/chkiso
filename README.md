# What is this?
[![Test](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml)
[![Build and Release](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml)

This is a project that is used to validate ISO images created by standard DSO pipelines.  It checks hashes for ISOs or media.

## Usage

### Basic Usage
By default, the script:
- Displays the SHA256 hash of the ISO/drive
- Verifies internal file integrity against embedded checksum files (*.sha, sha256sum.txt)

```powershell
.\chkiso.ps1 path\to\image.iso
```

### Advanced Options

#### Check against an expected SHA256 hash:
```powershell
.\chkiso.ps1 path\to\image.iso <sha256-hash>
```
**Note**: Content verification runs by default. Use `-NoVerify` to skip it.

#### Check against a hash file:
```powershell
.\chkiso.ps1 path\to\image.iso -ShaFile path\to\hashfile.sha
```
**Note**: Content verification runs by default. Use `-NoVerify` to skip it.

#### Enable implanted MD5 check:
```powershell
.\chkiso.ps1 path\to\image.iso -MD5
```

**Note**: If `checkisomd5.exe` is available in the current directory or PATH, it will be used automatically to avoid FIPS restrictions.

**Note**: SHA256 calculations will automatically use `sha256sum.exe` if available in the current directory or PATH for improved performance. The script will fall back to built-in calculation if the utility is not found.

#### Skip internal file verification:
```powershell
.\chkiso.ps1 path\to\image.iso -NoVerify
```

### Windows Executable
Download the compiled `chkiso.exe` from the [Releases](https://github.com/pappasjfed/chkiso/releases) page and run it:
```cmd
chkiso.exe path\to\image.iso
```

**Important Limitation**: Due to technical limitations in compiled executables (ps2exe), `chkiso.exe` cannot access drive letters - this includes both mounted ISOs and physical CD/DVD drives. If you need to verify media via a drive letter:
- Use the ISO file path directly: `chkiso.exe C:\path\to\image.iso`
- Or use the PowerShell script: `powershell -File chkiso.ps1 E:`

The PowerShell script (`chkiso.ps1`) supports both ISO file paths and drive letters for mounted ISOs or physical media.

## Building

The Windows executable is automatically built and attached to releases via GitHub Actions. 

### Code Signing

The executable can be automatically signed if code signing certificates are configured. See [CODE_SIGNING.md](CODE_SIGNING.md) for setup instructions. If certificates are not configured, the workflow will build an unsigned executable.

### Manual Build

To build manually:

1. Install ps2exe: `Install-Module -Name ps2exe -Force`
2. Compile: `ps2exe -inputFile chkiso.ps1 -outputFile chkiso.exe -noConsole:$false -title "chkiso" -version "1.0.0.0" -company "chkiso" -product "chkiso" -copyright "MIT License"`

## Testing

Tests run automatically on pull requests and pushes to main. The test suite validates the executable against `test/test.iso` using multiple verification methods.

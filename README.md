# What is this?
[![Test](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml)
[![Build and Release](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml)

This is a project that is used to validate ISO images created by standard DSO pipelines.  It checks hashes for ISOs or media.

## Usage

### PowerShell Script
Run the PowerShell script directly:
```powershell
.\chkiso.ps1 path\to\image.iso
```

### Windows Executable
Download the compiled `chkiso.exe` from the [Releases](https://github.com/pappasjfed/chkiso/releases) page and run it:
```cmd
chkiso.exe path\to\image.iso
```

## Building

The Windows executable is automatically built and attached to releases via GitHub Actions. To build manually:

1. Install ps2exe: `Install-Module -Name ps2exe -Force`
2. Compile: `ps2exe -inputFile chkiso.ps1 -outputFile chkiso.exe -noConsole:$false -title "chkiso" -version "1.0.0.0" -company "chkiso" -product "chkiso" -copyright "MIT License"`

## Testing

Tests run automatically on pull requests and pushes to main. The test suite validates the executable against `test/test.iso` using multiple verification methods.

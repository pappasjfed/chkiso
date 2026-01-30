# Copilot Instructions for chkiso

## Repository Overview

This repository contains `chkiso`, a PowerShell-based tool for verifying ISO image integrity using multiple validation methods. The tool checks SHA256 hashes, MD5 implants, and internal file checksums. It's compiled into a Windows executable using ps2exe for distribution.

## Technology Stack

- **Language**: PowerShell (`.ps1`)
- **Compilation**: ps2exe module (PowerShell to executable)
- **Runtime**: Windows (PowerShell 5.1+ or PowerShell Core)
- **CI/CD**: GitHub Actions
- **Testing**: PowerShell scripts on Windows runners

## Project Structure

- `chkiso.ps1` - Main PowerShell script with verification logic
- `test/` - Test files including test ISO image and hash files
- `.github/workflows/` - CI/CD workflows for testing and releases
- `CODE_SIGNING.md` - Documentation for code signing setup
- `README.md` - User-facing documentation

## Building and Testing

### Testing
Tests run automatically on pull requests and pushes to main. To test locally:

```powershell
# Run the script directly with test ISO
.\chkiso.ps1 test\test.iso

# Test with hash verification
.\chkiso.ps1 test\test.iso -ShaFile test\test.iso.sha

# Test with MD5 check
.\chkiso.ps1 test\test.iso -MD5
```

### Building
The executable is built automatically via GitHub Actions. To build manually:

```powershell
# Install ps2exe module
Install-Module -Name ps2exe -Force

# Compile to executable
ps2exe -inputFile chkiso.ps1 -outputFile chkiso.exe `
  -noConsole:$false -title "chkiso" -version "1.0.0.0" `
  -company "chkiso" -product "chkiso" -copyright "MIT License"
```

## Code Style and Conventions

### PowerShell Best Practices
- Use approved PowerShell verbs (Get-, Set-, Test-, etc.)
- Follow PascalCase for function names
- Use proper parameter declarations with `[CmdletBinding()]`
- Include parameter validation and help documentation
- Use `Write-Host` for user-facing messages
- Use `Write-Error` for errors
- Properly dispose of streams and cryptographic objects

### Error Handling
- Track errors using `$script:hasErrors` flag
- Exit with proper exit codes (0 for success, 1 for failure)
- Provide clear, actionable error messages to users

### ps2exe Compatibility
- **Important**: The script must work when compiled to `.exe`
- Avoid cmdlets that don't work in ps2exe (e.g., `Get-FileHash`)
- Use .NET classes directly when needed (e.g., `FileStream`, `SHA256`)
- Drive letter access is not supported in compiled executables due to ps2exe limitations

### Parameter Design
- Maintain backward compatibility with positional parameters
- Use descriptive aliases for common use cases
- Document limitations (e.g., drive letter access in compiled exe)

## Testing Requirements

When making changes:
1. Test with the script directly (`.\chkiso.ps1`)
2. Test after compilation to `.exe`
3. Verify all verification modes work:
   - Default (SHA256 display + content verification)
   - SHA256 hash string verification
   - SHA256 hash file verification
   - MD5 implant check
   - Content-only verification
4. Ensure proper exit codes (0 = success, 1 = failure)

## GitHub Actions Workflows

### test.yml
Runs on every PR and push to main:
- Compiles the script to executable
- Optionally signs the executable (if certificates configured)
- Runs comprehensive tests with test.iso
- Tests mounted ISO scenarios (script only)

### build-release.yml
Runs on releases:
- Compiles and optionally signs executable
- Attaches artifacts to GitHub releases

## Important Notes

### Known Limitations
- Compiled `.exe` cannot access drive letters (mounted ISOs or physical drives) due to ps2exe Win32 device path limitations
- Users should use ISO file paths directly with the `.exe`
- The PowerShell script (`.ps1`) supports both file paths and drive letters

### Security Considerations
- Code signing is optional but recommended for distribution
- See `CODE_SIGNING.md` for setup instructions
- SHA256 checksums are generated for all releases

## Making Changes

1. Modify `chkiso.ps1` with your changes
2. Test locally with PowerShell script
3. Verify compilation succeeds: `ps2exe -inputFile chkiso.ps1 -outputFile chkiso.exe ...`
4. Test the compiled executable
5. Update README.md if user-facing behavior changes
6. Submit PR - automated tests will run
7. Ensure all tests pass before merging

## Documentation

- Keep `README.md` updated with user-facing changes
- Document parameter changes in script help comments
- Update `CODE_SIGNING.md` if signing process changes
- Use clear examples in documentation

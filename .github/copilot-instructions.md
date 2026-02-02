# Copilot Instructions for chkiso

## Repository Overview

This repository contains `chkiso`, a cross-platform Go tool for verifying ISO image integrity using multiple validation methods. The tool checks SHA256 hashes, MD5 implants, and internal file checksums. It compiles to native binaries for Windows, Linux, macOS, and FreeBSD.

## Technology Stack

- **Language**: Go 1.21+
- **Build System**: Go toolchain + Makefile
- **Runtime**: Cross-platform (Windows, Linux, macOS, FreeBSD)
- **CI/CD**: GitHub Actions with multi-platform builds
- **Testing**: Go testing framework + integration tests

## Project Structure

- `main.go` - Main Go source code with verification logic
- `go.mod` / `go.sum` - Go module dependencies
- `Makefile` - Build automation for multiple platforms
- `test/` - Test files including test ISO image and hash files
- `.github/workflows/` - CI/CD workflows for build, test, release, security
- `WORKFLOWS.md` - Comprehensive workflow documentation
- `README.md` - User-facing documentation

## Building and Testing

### Building
Build for your current platform:

```bash
# Using Go directly
go build -o chkiso

# Or using Makefile
make build

# Build for all platforms
make build-all

# Build for specific platforms
make windows    # All Windows targets
make linux      # All Linux targets
make macos      # All macOS targets
```

### Testing
Tests run automatically on pull requests and pushes to main. To test locally:

```bash
# Build the binary
go build -o chkiso

# Run with test ISO
./chkiso test/test.iso

# Test with hash verification
./chkiso test/test.iso -shafile test/test.iso.sha

# Test with MD5 check
./chkiso test/test.iso -md5

# Test with hash string (using env var from test.iso.sha)
./chkiso test/test.iso $(grep -o '^[a-f0-9]\{64\}' test/test.iso.sha) -noverify

# Run Go tests (if any exist)
go test -v ./...
```

## Code Style and Conventions

### Go Best Practices
- Follow standard Go formatting (`gofmt`, `go fmt`)
- Use meaningful package, function, and variable names
- Keep functions focused and single-purpose
- Handle errors explicitly - don't ignore them
- Use idiomatic Go patterns and conventions
- Add comments for exported functions and complex logic
- Use Go standard library when possible

### Error Handling
- Track errors using `hasErrors` global flag
- Exit with proper exit codes (0 for success, 1 for failure)
- Provide clear, actionable error messages to users
- Use `fmt.Fprintf(os.Stderr, ...)` for error output
- Log important operations and their results

### Cross-Platform Considerations
- **Important**: Code must work on Windows, Linux, macOS, and FreeBSD
- Use `filepath` package for path operations (handles OS differences)
- Test platform-specific features (e.g., drive letters on Windows)
- Use `runtime.GOOS` and `runtime.GOARCH` for platform detection
- Be mindful of file permissions and path separators

### Command-Line Interface
- Support both flag-based and positional arguments where appropriate
- Maintain backward compatibility with existing command patterns
- Use descriptive flag names and aliases
- Document all flags and options in help text

## Testing Requirements

When making changes:
1. Build the binary: `go build -o chkiso`
2. Test with the test ISO: `./chkiso test/test.iso`
3. Verify all verification modes work:
   - Default (SHA256 display + content verification)
   - SHA256 hash string verification (positional and flag)
   - SHA256 hash file verification
   - MD5 implant check
   - Content-only verification (skip other checks)
4. Test cross-platform if possible (at minimum, ensure code is portable)
5. Ensure proper exit codes (0 = success, 1 = failure)
6. Run `go test` if unit tests exist

## GitHub Actions Workflows

### build-go.yml
Runs on every PR and push to main:
- Sets up Go 1.21+ environment
- Builds the binary
- Tests basic functionality with test.iso
- Tests all verification modes (hash string, hash file, MD5, content)
- Validates exit codes

### release.yml
Runs on releases:
- Builds binaries for all supported platforms:
  - Windows: amd64, arm64
  - Linux: amd64, 386, arm64, arm
  - macOS: amd64, arm64
  - FreeBSD: amd64
- Generates SHA256 checksums for each binary
- Attaches all binaries and checksums to GitHub releases

### security.yml
Runs on push, PR, and scheduled (daily):
- CodeQL analysis for Go code
- Dependency review (PR only)
- Secret scanning with TruffleHog
- Go security analysis with govulncheck and gosec

### publish.yml
Publishes documentation and artifacts when needed

### documentation.yml
Validates and checks documentation including:
- Markdown link checking
- Documentation formatting

## Important Notes

### Cross-Platform Support
- Binary works on Windows, Linux, macOS, and FreeBSD
- No dependencies required - statically linked binaries
- No FIPS restrictions - MD5 works on all platforms
- Automatic ISO mounting on Windows using PowerShell

### Platform-Specific Features
- **Windows**: Supports drive letters (e.g., `E:`) and automatic ISO mounting
- **Linux/Unix**: Requires mount points or direct ISO file paths
- **All platforms**: Direct ISO file access always works

### Security Considerations
- CodeQL scanning runs automatically on all PRs
- Dependency scanning with Dependabot
- Secret scanning with TruffleHog
- Go security analysis with gosec and govulncheck
- SHA256 checksums generated for all releases

## Making Changes

1. Modify `main.go` with your changes
2. Format code: `go fmt`
3. Build: `go build -o chkiso`
4. Test locally with test ISO and various modes
5. Update README.md if user-facing behavior changes
6. Update WORKFLOWS.md if workflow behavior changes
7. Run security checks if adding dependencies
8. Submit PR - automated tests, builds, and security scans will run
9. Ensure all checks pass before merging

## Key Files

- `main.go` - All application logic (single file architecture)
- `go.mod` - Go module definition and dependencies
- `Makefile` - Build automation for multiple platforms
- `README.md` - User documentation with usage examples
- `WORKFLOWS.md` - Detailed workflow documentation for developers
- `.github/workflows/` - Complete CI/CD pipeline definitions

## Development Workflow

1. **Before starting**: Pull latest from main
2. **During development**: 
   - Test locally with `go build && ./chkiso test/test.iso`
   - Use `go fmt` to format code
   - Check for common issues: `go vet`
3. **Before committing**:
   - Ensure code builds: `go build`
   - Test all verification modes
   - Update documentation if needed
4. **PR submission**: 
   - CI will run tests, builds, and security scans
   - Fix any issues identified by automated checks

## Documentation

- Keep `README.md` updated with user-facing changes
- Update `WORKFLOWS.md` for workflow changes
- Document any new flags or options in help text
- Use clear examples in documentation
- Maintain cross-platform compatibility in examples

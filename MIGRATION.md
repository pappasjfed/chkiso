# Migration Guide: PowerShell to Go

## Overview

As of version 2.0, chkiso is now available as a **Go binary** in addition to the original PowerShell script. This document explains why we made this change and how to migrate.

## Why Go?

The PowerShell version faced several limitations that the Go version addresses:

### 1. FIPS Restrictions
**Problem**: On FIPS-compliant systems, PowerShell blocks MD5 hashing operations, preventing the `-MD5` flag from working.

**Solution**: The Go version has no FIPS restrictions and can always perform MD5 hashing.

### 2. Script Execution Policies
**Problem**: Many enterprise and government systems have strict execution policies that block `.ps1` scripts, even when digitally signed.

**Solution**: The Go version compiles to a native executable that doesn't require PowerShell and isn't blocked by script execution policies.

### 3. Platform Limitations
**Problem**: PowerShell is primarily Windows-focused. While PowerShell Core runs on Linux/macOS, it's not commonly installed on those platforms.

**Solution**: The Go version is truly cross-platform with native binaries for Windows, Linux, macOS, and FreeBSD.

### 4. Drive Access in Compiled Executables
**Problem**: When using ps2exe to compile PowerShell scripts, the resulting `.exe` cannot access Win32 device paths like `\\.\E:`, which are needed for drive letter access.

**Solution**: The Go version has native Win32 API support and can access drive letters directly.

### 5. Runtime Dependencies
**Problem**: PowerShell requires the PowerShell runtime to be installed.

**Solution**: The Go version is a single statically-linked executable with zero dependencies.

## Command-Line Compatibility

The Go version maintains **100% command-line compatibility** with the PowerShell version. All existing commands work the same way:

```bash
# These commands work identically in both versions:
chkiso image.iso
chkiso image.iso -sha256 <hash>
chkiso image.iso -shafile hashes.sha
chkiso image.iso -md5
chkiso E: -noverify
```

## Feature Comparison

| Feature | PowerShell | Go |
|---------|-----------|-----|
| SHA256 verification | ✅ | ✅ |
| MD5 implanted check | ⚠️ (FIPS restricted) | ✅ |
| Hash file verification | ✅ | ✅ |
| Content verification | ✅ | ✅ |
| Drive letter access | ⚠️ (script only) | ✅ |
| Windows support | ✅ | ✅ |
| Linux support | ⚠️ (PS Core) | ✅ |
| macOS support | ⚠️ (PS Core) | ✅ |
| FreeBSD support | ❌ | ✅ |
| Single executable | ⚠️ (ps2exe issues) | ✅ |
| No runtime needed | ❌ | ✅ |
| Works in locked-down env | ❌ | ✅ |

## Migration Steps

### For Windows Users

**Option 1: Use the Go binary (Recommended)**
1. Download `chkiso-windows-amd64.exe` from [Releases](https://github.com/pappasjfed/chkiso/releases)
2. Rename to `chkiso.exe` (optional)
3. Use exactly as before: `chkiso.exe image.iso`

**Option 2: Continue using PowerShell**
- The PowerShell script (`chkiso.ps1`) is still maintained
- Use when you need drive letter access with the script
- Available as `chkiso-ps.exe` (ps2exe version) in releases

### For Linux/macOS Users

1. Download the appropriate binary:
   - Linux: `chkiso-linux-amd64` (or arm64, arm, 386)
   - macOS Intel: `chkiso-darwin-amd64`
   - macOS Apple Silicon: `chkiso-darwin-arm64`
2. Make executable: `chmod +x chkiso-*`
3. Optionally move to PATH: `sudo mv chkiso-* /usr/local/bin/chkiso`

### For CI/CD Pipelines

Update your scripts to use the Go binary:

**Before (PowerShell):**
```yaml
- name: Verify ISO
  run: powershell -File chkiso.ps1 image.iso -sha256 ${{ env.EXPECTED_HASH }}
```

**After (Go):**
```yaml
- name: Verify ISO
  run: ./chkiso image.iso -sha256 ${{ env.EXPECTED_HASH }}
```

## Building from Source

### PowerShell Version
```powershell
Install-Module -Name ps2exe -Force
ps2exe -inputFile chkiso.ps1 -outputFile chkiso.exe
```

### Go Version
```bash
go build -o chkiso
# Or use Makefile for cross-platform builds
make build-all
```

## Frequently Asked Questions

### Q: Will the PowerShell version be deprecated?
A: No. The PowerShell version will continue to be maintained as an alternative. However, we recommend using the Go version for its superior cross-platform support and lack of restrictions.

### Q: Can I use both versions?
A: Yes! You can keep both. The PowerShell script is `chkiso.ps1`, and the Go binary can be named `chkiso` (or `chkiso.exe` on Windows).

### Q: What about backward compatibility?
A: The Go version maintains full command-line compatibility. All flags and arguments work the same way.

### Q: Will my scripts break?
A: No. If you're calling `chkiso.ps1` explicitly, it will continue to work. If you're calling `chkiso.exe`, download the Go binary and name it `chkiso.exe`.

### Q: Can I use the Go version on FIPS-compliant systems?
A: Yes! This is one of the main advantages. The Go version works on FIPS-compliant systems without any restrictions.

### Q: Does the Go version support all PowerShell features?
A: Yes. All verification features are implemented:
- SHA256 hash calculation
- MD5 implanted hash verification
- External hash file verification
- Content verification
- Drive letter access (Windows)

### Q: What if I find a bug in the Go version?
A: Please open an issue on GitHub. You can always fall back to the PowerShell version while we fix it.

## Performance

The Go version is typically **faster** than the PowerShell version:
- Native compiled code vs interpreted script
- Optimized I/O operations
- Lower memory footprint

## Support

- **Issues**: [GitHub Issues](https://github.com/pappasjfed/chkiso/issues)
- **Discussions**: [GitHub Discussions](https://github.com/pappasjfed/chkiso/discussions)
- **Documentation**: [README.md](README.md)

## License

Both versions are licensed under the MIT License.

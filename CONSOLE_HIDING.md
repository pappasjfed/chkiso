# Hiding Console Window on Windows

## Problem
When double-clicking the EXE from Windows Explorer, a second console window appears briefly.

## Solution

### Method 1: Build Flag (Recommended)
Add `-H windowsgui` flag during Windows AMD64 build to prevent console window creation.

**Update Makefile**:
```makefile
windows-amd64:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS) -H windowsgui" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
```

**Update GitHub Actions** (`.github/workflows/build-go.yml`):
```yaml
- name: Build Windows amd64 with CGO
  if: matrix.target == 'windows-amd64-cgo'
  run: go build -ldflags="-s -w -H windowsgui" -trimpath -o chkiso-windows-amd64.exe
```

### Method 2: FreeConsole API (Alternative)
Call `FreeConsole()` in the code when GUI mode is detected.

Add to `gui_fyne.go`:
```go
func hideConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	freeConsole := kernel32.NewProc("FreeConsole")
	freeConsole.Call()
}

// Call at start of runGUI():
func runGUI() {
	hideConsole() // Hide console window
	...
}
```

## Trade-offs

**-H windowsgui**:
- ✅ No console window at all
- ❌ Can't see debug output when run from CLI
- ❌ Affects ALL launches, not just GUI

**FreeConsole()**:
- ✅ Only hides when GUI mode active
- ✅ CLI mode still works normally
- ❌ Window may flash briefly before hiding

## Recommended Approach
Use `-H windowsgui` for Windows AMD64 GUI build since this version always uses GUI mode when launched without args.

For CLI-only builds (like Windows ARM64), use normal console mode.

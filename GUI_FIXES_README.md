# GUI Fixes - Implementation Status

## Overview

This document tracks the status of GUI fixes requested by the user after testing the Fyne-based GUI implementation.

## User-Reported Issues

### 1. Second Console Window Appears
**Status**: 📝 Documented, Ready to Implement

**Problem**: When double-clicking the EXE from Explorer, a second console window briefly appears.

**Solution**: Add `-H windowsgui` linker flag to Windows AMD64 build.

**Implementation**:
- See: `CONSOLE_HIDING.md` for complete guide
- Update: `Makefile` line for windows-amd64 target
- Update: `.github/workflows/build-go.yml` Windows build step
- Change: Add `-H windowsgui` to `-ldflags`

### 2. Font Color Too Light (Low Contrast)
**Status**: 📝 Documented, Ready to Implement

**Problem**: Light grey text on slightly darker grey background is hard to read. This is a color issue, not a font style issue (bold doesn't help).

**Solution**: Explicitly set text color to white for high contrast.

**Implementation**:
- File: `gui_fyne.go` line 82
- Add explicit color setting using `color.NRGBA{R: 255, G: 255, B: 255, A: 255}`
- Or use Fyne RichText with color markup
- See code snippet in `fixes_summary.md`

### 3. EXE Icon Not Showing in Explorer
**Status**: 📝 Fully Documented

**Problem**: Runtime icon shows in window, but EXE file shows generic icon in Windows Explorer.

**Solution**: Embed icon resource in executable at build time using rsrc tool.

**Implementation**:
- See: `ICON_EMBEDDING.md` for complete step-by-step guide
- Three methods documented: rsrc, fyne package, goversioninfo
- Requires build-time tool installation
- Can be integrated into CI/CD workflow

### 4. No Progress Meter During Verification
**Status**: 📝 Documented, Ready to Implement

**Problem**: Long operations (4GB DVD) show "Step 1/3: Reading..." but no visual progress indication.

**Solution**: Add ProgressBar widget that updates during operations.

**Implementation**:
- File: `gui_fyne.go` after line 82
- Add `progressBar := widget.NewProgressBar()`
- Update progress in verification goroutines
- Show percentage for each step (0%, 33%, 66%, 100%)
- See code snippet in `fixes_summary.md`

### 5. checkisomd5 Using Wrong Device Path
**Status**: 📝 Documented, Ready to Implement

**Problem**: checkisomd5 is called with "G:" but should use device path "\\.\G:" for raw device access.

**Solution**: Convert drive letter to Windows device path format before calling external tool.

**Implementation**:
- File: `windows_helpers.go` line 143
- Change from: `config.Path`
- Change to: `fmt.Sprintf("\\\\.\\%s:", config.driveLetter)` when isDrive==true
- See exact code in `fixes_summary.md`

### 6. High Memory Usage (3.5GB)
**Status**: ℹ️ Analyzed

**Problem**: Application uses 3.5GB RAM during verification.

**Analysis**:
- Code already uses streaming with `io.Copy()` 
- File is not loaded into memory entirely
- Likely causes:
  - Fyne UI buffering large text output
  - Multiple verification operations accumulating data
  - Windows file system caching
  - Debug logging accumulation

**Recommendations**:
- Limit text buffer size in resultText widget
- Clear previous output before new operations
- Consider chunked/paginated output for large results
- Profile with pprof if issue persists

## Quick Implementation Checklist

- [ ] Add `-H windowsgui` to Windows AMD64 build flags
- [ ] Set explicit white color for resultText
- [ ] Add ProgressBar widget to GUI
- [ ] Convert drive letters to device paths in runCheckisomd5
- [ ] (Optional) Embed icon using rsrc tool
- [ ] Test on actual Windows system
- [ ] Verify memory usage with large ISO

## Files to Modify

1. **windows_helpers.go** (line 143)
   - Device path conversion

2. **gui_fyne.go** (line 82+)
   - Text color fix
   - Progress bar addition

3. **Makefile** (windows-amd64 target)
   - Add `-H windowsgui` flag

4. **.github/workflows/build-go.yml** (Windows build step)
   - Add `-H windowsgui` flag

## Estimated Effort

- **Code changes**: 30-60 minutes
- **Testing**: 30 minutes
- **Icon embedding**: 15-30 minutes (optional)
- **Total**: 1-2 hours

## Testing Procedure

1. Build with all flags
2. Double-click EXE from Explorer → Should NOT show console window
3. Check text readability → Should be white/high contrast
4. Start verification → Should show progress bar updating
5. Use checkisomd5 → Should use \\.\X: device path
6. Monitor memory → Should not exceed reasonable limits

## References

- `CONSOLE_HIDING.md` - Complete guide for console window
- `ICON_EMBEDDING.md` - Complete guide for icon embedding
- `fixes_summary.md` - Quick reference with code snippets
- User feedback in issue/PR comments

## Status Summary

| Issue | Priority | Difficulty | Status |
|-------|----------|------------|---------|
| Console window | High | Easy | Documented |
| Text color | High | Easy | Documented |
| Progress bar | Medium | Easy | Documented |
| Device path | High | Easy | Documented |
| EXE icon | Low | Medium | Documented |
| Memory usage | Low | Medium | Analyzed |

**Overall**: Ready for implementation. All issues documented with clear solutions.

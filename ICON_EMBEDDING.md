# Adding Icon to Windows EXE

## Overview
To make the chkiso.exe show the gold CD icon in Windows Explorer, you need to embed the icon into the executable. This is done during the build process.

## Method 1: Using `fyne package` (Recommended)

Fyne provides a built-in packaging tool that can embed icons:

```bash
# Install fyne command if not already installed
go install fyne.io/fyne/v2/cmd/fyne@latest

# Package the application with icon
fyne package -os windows -icon Icon.png

# This creates chkiso.exe with embedded icon
```

The icon needs to be provided as a PNG file. You can export the SVG icon from `icon.go` to PNG or create a PNG version.

## Method 2: Using `goversioninfo` and rsrc

For more control, you can use `goversioninfo` or `rsrc` tools:

### Step 1: Install rsrc tool
```bash
go install github.com/akavel/rsrc@latest
```

### Step 2: Create icon.ico file
Convert the gold CD icon to .ico format (256x256 or multiple sizes). You can use online converters or tools like ImageMagick:

```bash
# If you have ImageMagick:
convert Icon.png -define icon:auto-resize=256,128,64,48,32,16 icon.ico
```

### Step 3: Create Windows resource file
Create a file named `chkiso.rc`:
```rc
IDI_ICON1 ICON "icon.ico"
```

### Step 4: Compile resource file
```bash
rsrc -ico icon.ico -o rsrc.syso
```

This creates `rsrc.syso` which Go will automatically include in the build.

### Step 5: Build normally
```bash
go build -ldflags="-s -w" -o chkiso.exe
```

The icon will now be embedded in the exe!

## Method 3: In GitHub Actions CI

Update `.github/workflows/build-go.yml` to include icon embedding:

```yaml
- name: Embed Icon (Windows amd64)
  if: matrix.goos == 'windows' && matrix.goarch == 'amd64'
  run: |
    # Install rsrc tool
    go install github.com/akavel/rsrc@latest
    
    # Create icon.ico from PNG (you'll need to add icon.png to repo)
    # Or use pre-created icon.ico
    
    # Generate rsrc.syso
    rsrc -ico icon.ico -o rsrc.syso

- name: Build
  run: |
    go build -ldflags="-s -w" -trimpath -o ${{ matrix.output }}
```

## Icon File Requirements

- **Format**: ICO (Windows icon format)
- **Sizes**: Multiple sizes recommended (16x16, 32x32, 48x48, 64x64, 128x128, 256x256)
- **Design**: Match the gold CD design from `icon.go`
- **Transparency**: Should have transparent background
- **Location**: Place `icon.ico` in repository root

## Current Status

The runtime icon (in window title bar) is already working via `icon.go`. This document describes how to add the same icon to the EXE file itself for Windows Explorer.

## Testing

After building with embedded icon:
1. View chkiso.exe in Windows Explorer
2. Icon should appear as gold CD disc
3. Right-click → Properties should show icon
4. Taskbar should show icon when running
5. Alt+Tab should show icon

## Notes

- The `rsrc.syso` file is platform-specific and should be gitignored
- Only needed for Windows builds
- Linux/macOS don't use .ico files
- The Fyne runtime icon works on all platforms

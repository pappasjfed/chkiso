//go:build !windows
// +build !windows

package main

// hasConsole always returns true on non-Windows platforms (they always have console)
func hasConsole() bool {
	return true
}

// runGUI is not supported on non-Windows platforms
func runGUI() {
	// This should never be called as GUI mode is Windows-only
	panic("GUI mode is only supported on Windows")
}

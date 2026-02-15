//go:build windows && arm64
// +build windows,arm64

package main

import (
	"fmt"
	"os"
	"syscall"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleCP = kernel32.NewProc("GetConsoleCP")
)

// hasConsole checks if the process is attached to a console window
func hasConsole() bool {
	r1, _, _ := procGetConsoleCP.Call()
	return r1 != 0
}

// runGUI shows an error message that GUI is not available on Windows ARM64
func runGUI() {
	fmt.Fprintln(os.Stderr, "\n⚠️  GUI Mode Not Available on Windows ARM64")
	fmt.Fprintln(os.Stderr, "\nThe GUI requires OpenGL which is not available for Windows ARM64 builds.")
	fmt.Fprintln(os.Stderr, "\nPlease use command-line mode instead:")
	fmt.Fprintln(os.Stderr, "\n  To verify a CD/DVD drive:")
	fmt.Fprintln(os.Stderr, "    chkiso.exe E:\\")
	fmt.Fprintln(os.Stderr, "\n  To verify an ISO file:")
	fmt.Fprintln(os.Stderr, "    chkiso.exe C:\\path\\to\\file.iso")
	fmt.Fprintln(os.Stderr, "\n  To check implanted MD5:")
	fmt.Fprintln(os.Stderr, "    chkiso.exe file.iso -md5")
	fmt.Fprintln(os.Stderr, "\n  For all options:")
	fmt.Fprintln(os.Stderr, "    chkiso.exe -help")
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

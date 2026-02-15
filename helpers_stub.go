//go:build !windows

package main

import (
	"fmt"
)

// isCheckisomd5Available always returns false on non-Windows platforms
func isCheckisomd5Available() bool {
	return false
}

// runCheckisomd5 is not available on non-Windows platforms
func runCheckisomd5(config *Config) error {
	return fmt.Errorf("checkisomd5.exe is only available on Windows")
}

//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// isCheckisomd5Available checks if checkisomd5 is available in PATH or exe directory
func isCheckisomd5Available() bool {
	// First check if it's in PATH
	_, err := exec.LookPath("checkisomd5")
	if err == nil {
		return true
	}

	// Check in the same directory as the executable
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	exeDir := filepath.Dir(exePath)
	checkisomd5Path := filepath.Join(exeDir, "checkisomd5")

	_, err = os.Stat(checkisomd5Path)
	return err == nil
}

// runCheckisomd5 runs the external checkisomd5 tool (Linux/macOS/FreeBSD)
func runCheckisomd5(config *Config) error {
	// Try to find checkisomd5 in PATH first
	checkisoPath, err := exec.LookPath("checkisomd5")
	if err != nil {
		// Check in the same directory as the executable
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %v", err)
		}

		exeDir := filepath.Dir(exePath)
		checkisoPath = filepath.Join(exeDir, "checkisomd5")

		// Verify it exists
		if _, err := os.Stat(checkisoPath); err != nil {
			return fmt.Errorf("checkisomd5 not found in PATH or executable directory")
		}
	}

	if config.GuiMode {
		fmt.Println("\n=== Running checkisomd5 ===")
		fmt.Printf("Tool path: %s\n", checkisoPath)
		fmt.Printf("ISO file: %s\n\n", config.Path)
	}

	// Run checkisomd5 with -v (verbose) flag
	cmd := exec.Command(checkisoPath, "-v", config.Path)
	output, err := cmd.CombinedOutput()

	// Print the output regardless of error (checkisomd5 may return non-zero on failure)
	fmt.Print(string(output))

	if err != nil {
		// Check if it's just a non-zero exit code (which might be expected for failed checks)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if config.GuiMode {
				fmt.Printf("\ncheckisomd5 exited with code: %d\n", exitErr.ExitCode())
			}
			// Don't return error - the output already shows what went wrong
			return nil
		}
		return fmt.Errorf("failed to run checkisomd5: %v", err)
	}

	if config.GuiMode {
		fmt.Println("\n=== checkisomd5 completed ===")
	}

	return nil
}

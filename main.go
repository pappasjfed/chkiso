package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	PVD_OFFSET          = 32768
	PVD_SIZE            = 2048
	APP_USE_OFFSET      = 883
	APP_USE_SIZE        = 512
	SECTOR_SIZE         = 2048
	SPACE_CHAR          = 0x20  // Space character used for neutralizing PVD
	VERSION             = "2.0.0"
)

var (
	hasErrors = false
)

type Config struct {
	Path               string
	Sha256Hash         string
	ShaFile            string
	NoVerify           bool
	MD5Check           bool
	Dismount           bool
	GuiMode            bool   // Explicitly request GUI mode
	isDrive            bool
	driveLetter        string
	mountedISO         bool   // Track if we mounted the ISO (vs user-mounted)
	mountedDriveLetter string // Drive letter where we mounted the ISO
}

func main() {
	// Check for explicit GUI flag first (before any other processing)
	// This allows users to force GUI mode from command line
	for _, arg := range os.Args[1:] {
		if arg == "-gui" || arg == "--gui" {
			if runtime.GOOS == "windows" {
				runGUI()
				return
			} else {
				fmt.Fprintf(os.Stderr, "Error: GUI mode is only supported on Windows\n")
				os.Exit(1)
			}
		}
	}
	
	// Check if we should run in GUI mode (Windows only)
	// GUI mode is triggered when:
	// 1. Running on Windows
	// 2. No command-line arguments provided (or only the executable name)
	// 3. No console is attached (double-clicked from Explorer)
	if runtime.GOOS == "windows" && len(os.Args) == 1 && !hasConsole() {
		runGUI()
		return
	}
	
	// If we have a console on Windows but no arguments, try to attach to parent console
	// This makes it work better when run from cmd.exe or PowerShell
	if runtime.GOOS == "windows" && len(os.Args) == 1 && hasConsole() {
		// Running from console but no args - show usage
		printUsage()
		os.Exit(1)
	}
	
	// CLI mode
	config := parseFlags()
	
	// Validate and resolve the path
	if err := validatePath(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	// Execute checks based on provided parameters
	if config.ShaFile != "" {
		verifyPathAgainstHashFile(config)
	}
	if config.Sha256Hash != "" {
		verifyPathAgainstHashString(config)
	}
	// If neither Sha256Hash nor ShaFile is provided, display SHA256 for informational purposes
	if config.Sha256Hash == "" && config.ShaFile == "" {
		displaySha256Hash(config)
	}
	if config.MD5Check {
		verifyImplantedMD5(config)
	}
	// Run VerifyContents by default unless -NoVerify is specified
	if !config.NoVerify {
		verifyContents(config)
	}
	
	if config.Dismount {
		handleDismount(config)
	}
	
	// Exit with proper code based on whether errors occurred
	if hasErrors {
		os.Exit(1)
	}
	os.Exit(0)
}

func parseFlags() *Config {
	config := &Config{}
	
	// Manual argument parsing for better flexibility
	var args []string
	i := 1
	for i < len(os.Args) {
		arg := os.Args[i]
		
		switch {
		case arg == "-version" || arg == "--version":
			fmt.Printf("chkiso version %s\n", VERSION)
			fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			os.Exit(0)
		case arg == "-help" || arg == "--help" || arg == "-h":
			printUsage()
			os.Exit(0)
		case arg == "-sha256" || arg == "--sha256" || arg == "-sha256sum" || arg == "--sha256sum" || arg == "-sha" || arg == "--sha":
			if i+1 < len(os.Args) {
				config.Sha256Hash = os.Args[i+1]
				i += 2
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires an argument\n", arg)
				os.Exit(1)
			}
		case arg == "-shafile" || arg == "--shafile":
			if i+1 < len(os.Args) {
				config.ShaFile = os.Args[i+1]
				i += 2
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires an argument\n", arg)
				os.Exit(1)
			}
		case arg == "-noverify" || arg == "--noverify":
			config.NoVerify = true
			i++
		case arg == "-md5" || arg == "--md5":
			config.MD5Check = true
			i++
		case arg == "-dismount" || arg == "--dismount" || arg == "-eject" || arg == "--eject":
			config.Dismount = true
			i++
		case arg == "-gui" || arg == "--gui":
			config.GuiMode = true
			i++
		default:
			// Positional argument
			args = append(args, arg)
			i++
		}
	}
	
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: path argument is required\n\n")
		printUsage()
		os.Exit(1)
	}
	
	config.Path = args[0]
	
	// Support positional sha256 hash (second argument)
	if len(args) >= 2 && config.Sha256Hash == "" {
		config.Sha256Hash = args[1]
	}
	
	return config
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "chkiso - ISO/Drive Verification Tool v%s\n\n", VERSION)
	fmt.Fprintf(os.Stderr, "Usage: chkiso [options] <path> [sha256-hash]\n\n")
	fmt.Fprintf(os.Stderr, "Arguments:\n")
	fmt.Fprintf(os.Stderr, "  path          Path to ISO file or drive letter (e.g., /path/to/image.iso or E:)\n")
	fmt.Fprintf(os.Stderr, "  sha256-hash   Optional SHA256 hash for verification (positional)\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  -sha256 <hash>      Expected SHA256 hash for verification\n")
	fmt.Fprintf(os.Stderr, "  -sha256sum <hash>   Alias for -sha256\n")
	fmt.Fprintf(os.Stderr, "  -sha <hash>         Alias for -sha256\n")
	fmt.Fprintf(os.Stderr, "  -shafile <file>     Path to SHA256 hash file\n")
	fmt.Fprintf(os.Stderr, "  -noverify           Skip verifying internal file hashes\n")
	fmt.Fprintf(os.Stderr, "  -md5                Enable implanted MD5 check\n")
	fmt.Fprintf(os.Stderr, "  -dismount           Dismount/eject after verification\n")
	fmt.Fprintf(os.Stderr, "  -eject              Alias for -dismount\n")
	fmt.Fprintf(os.Stderr, "  -gui                Launch GUI mode (Windows only)\n")
	fmt.Fprintf(os.Stderr, "  -version            Display version information\n")
	fmt.Fprintf(os.Stderr, "  -help               Display this help information\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  chkiso image.iso\n")
	fmt.Fprintf(os.Stderr, "  chkiso image.iso <hash>\n")
	fmt.Fprintf(os.Stderr, "  chkiso -sha256 <hash> image.iso\n")
	fmt.Fprintf(os.Stderr, "  chkiso -shafile hashes.sha image.iso\n")
	fmt.Fprintf(os.Stderr, "  chkiso -md5 image.iso\n")
	fmt.Fprintf(os.Stderr, "  chkiso -noverify E:\n")
	fmt.Fprintf(os.Stderr, "  chkiso -gui         (Windows: Launch GUI mode)\n")
}

func validatePath(config *Config) error {
	// Check if it's a drive letter (Windows style: E: or E:\)
	if runtime.GOOS == "windows" {
		drivePattern := regexp.MustCompile(`^([A-Za-z]):\\?$`)
		if matches := drivePattern.FindStringSubmatch(config.Path); matches != nil {
			config.isDrive = true
			config.driveLetter = strings.ToUpper(matches[1])
			// On Windows, we'll use device path for drive access
			return nil
		}
	}
	
	// Otherwise, treat as file path
	info, err := os.Stat(config.Path)
	if err != nil {
		return fmt.Errorf("file not found: %s", config.Path)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", config.Path)
	}
	
	// Resolve to absolute path
	absPath, err := filepath.Abs(config.Path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %v", err)
	}
	config.Path = absPath
	
	return nil
}

func getSha256Hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getSha256FromPath(config *Config) (string, error) {
	var reader io.Reader
	var file *os.File
	var err error
	
	if config.isDrive {
		fmt.Printf("Calculating SHA256 hash for drive '%s:' (this can be slow)...\n", config.driveLetter)
		// On Windows, use device path
		if runtime.GOOS == "windows" {
			devicePath := fmt.Sprintf("\\\\.\\%s:", config.driveLetter)
			file, err = os.Open(devicePath)
		} else {
			return "", fmt.Errorf("drive letters are only supported on Windows")
		}
	} else {
		fmt.Printf("Calculating SHA256 hash for file '%s'...\n", filepath.Base(config.Path))
		file, err = os.Open(config.Path)
	}
	
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	reader = file
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func verifyPathAgainstHashString(config *Config) {
	fmt.Println("\n--- Verifying Path Against Provided SHA256 Hash ---")
	expectedHash := strings.ToLower(strings.TrimSpace(config.Sha256Hash))
	
	// Validate hash format (must be 64 hex characters)
	if !regexp.MustCompile(`^[a-fA-F0-9]{64}$`).MatchString(expectedHash) {
		fmt.Fprintf(os.Stderr, "Error: Invalid SHA256 hash format. Expected 64 hexadecimal characters.\n")
		hasErrors = true
		return
	}
	
	calculatedHash, err := getSha256FromPath(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calculating hash: %v\n", err)
		hasErrors = true
		return
	}
	calculatedHash = strings.ToLower(calculatedHash)
	
	fmt.Printf("  - Expected:   %s\n", expectedHash)
	fmt.Printf("  - Calculated: %s\n", calculatedHash)
	
	if calculatedHash == expectedHash {
		fmt.Println("\033[32mResult: SUCCESS - Hashes match.\033[0m")
	} else {
		fmt.Println("\033[31mResult: FAILURE - Hashes DO NOT match.\033[0m")
		hasErrors = true
	}
}

func verifyPathAgainstHashFile(config *Config) {
	fmt.Println("\n--- Verifying Path Against SHA256 Hash File ---")
	
	content, err := os.ReadFile(config.ShaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading hash file: %v\n", err)
		hasErrors = true
		return
	}
	
	// Determine the filename pattern to search for
	var isoFileNamePattern string
	if config.isDrive {
		isoFileNamePattern = ".*\\.iso"
	} else {
		isoFileNamePattern = regexp.QuoteMeta(filepath.Base(config.Path))
	}
	
	// Try to find a hash entry matching the filename
	pattern := fmt.Sprintf(`^([a-fA-F0-9]{64})\s+\*?\s*%s`, isoFileNamePattern)
	re := regexp.MustCompile(pattern)
	genericPattern := regexp.MustCompile(`^([a-fA-F0-9]{64})\s+\*?\s*.*`)
	
	lines := strings.Split(string(content), "\n")
	var expectedHash string
	
	for _, line := range lines {
		if matches := re.FindStringSubmatch(line); matches != nil {
			expectedHash = strings.ToLower(matches[1])
			break
		}
	}
	
	// If no specific match, try generic pattern (first hash in file)
	if expectedHash == "" {
		for _, line := range lines {
			if matches := genericPattern.FindStringSubmatch(line); matches != nil {
				expectedHash = strings.ToLower(matches[1])
				break
			}
		}
	}
	
	if expectedHash == "" {
		fmt.Fprintf(os.Stderr, "Error: Could not find a valid SHA256 hash entry in the hash file '%s'\n", config.ShaFile)
		hasErrors = true
		return
	}
	
	config.Sha256Hash = expectedHash
	verifyPathAgainstHashString(config)
}

func displaySha256Hash(config *Config) {
	fmt.Println("\n--- SHA256 Hash (Informational) ---")
	calculatedHash, err := getSha256FromPath(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calculating hash: %v\n", err)
		hasErrors = true
		return
	}
	fmt.Printf("\033[33mSHA256: %s\033[0m\n", strings.ToLower(calculatedHash))
}

func verifyContents(config *Config) {
	fmt.Println("\n--- Verifying Contents ---")
	
	var mountPath string
	var needsCleanup bool
	
	if config.isDrive {
		if runtime.GOOS == "windows" {
			mountPath = fmt.Sprintf("%s:\\", config.driveLetter)
			fmt.Printf("Verifying contents of physical drive at: %s\n", mountPath)
		} else {
			fmt.Fprintf(os.Stderr, "Error: Drive verification is only supported on Windows\n")
			hasErrors = true
			return
		}
	} else {
		// For ISO files, try to mount them automatically on Windows
		if runtime.GOOS == "windows" {
			fmt.Printf("Mounting ISO: %s\n", config.Path)
			driveLetter, err := mountISO(config.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to mount ISO automatically: %v\n", err)
				fmt.Println("\nNote: For ISO files, please mount the ISO manually and verify using the mount point.")
				fmt.Println("Example (Windows): Mount-DiskImage image.iso, then run: chkiso E:")
				return
			}
			
			config.mountedISO = true
			config.mountedDriveLetter = driveLetter
			needsCleanup = true
			mountPath = fmt.Sprintf("%s:\\", driveLetter)
			fmt.Printf("Mounted to drive: %s:\n", driveLetter)
			
			// Ensure cleanup happens even if verification fails
			defer func() {
				if needsCleanup && config.mountedISO {
					fmt.Println("\nUnmounting ISO...")
					if err := dismountISO(config.Path); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to unmount ISO: %v\n", err)
						fmt.Printf("Please dismount manually using: Dismount-DiskImage -ImagePath '%s'\n", config.Path)
					} else {
						fmt.Println("ISO unmounted successfully.")
						config.mountedISO = false
					}
				}
			}()
		} else {
			// Non-Windows platforms
			fmt.Println("Note: For ISO files, please mount the ISO manually and verify using the mount point.")
			fmt.Println("Example (Linux): sudo mount -o loop image.iso /mnt, then run: chkiso /mnt")
			return
		}
	}
	
	fmt.Printf("Searching for checksum files (*.sha, sha256sum.txt, SHA256SUMS) in %s...\n", mountPath)
	
	// Find checksum files
	checksumFiles, err := findChecksumFiles(mountPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error finding checksum files: %v\n", err)
		return
	}
	
	if len(checksumFiles) == 0 {
		fmt.Println("Warning: Could not find any checksum files (*.sha, sha256sum.txt, SHA256SUMS) on the media.")
		return
	}
	
	// Report all found checksum files
	fmt.Printf("\nFound %d checksum file(s):\n", len(checksumFiles))
	for i, cf := range checksumFiles {
		relPath, err := filepath.Rel(mountPath, cf)
		if err != nil {
			relPath = cf
		}
		fmt.Printf("  %d. %s\n", i+1, relPath)
	}
	fmt.Println()
	
	totalFiles := 0
	failedFiles := 0
	
	for _, checksumFile := range checksumFiles {
		fmt.Printf("Processing checksum file: %s\n", filepath.Base(checksumFile))
		baseDir := filepath.Dir(checksumFile)
		
		file, err := os.Open(checksumFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not open checksum file: %v\n", err)
			continue
		}
		defer file.Close()  // Ensure file is closed even if we continue early
		
		scanner := bufio.NewScanner(file)
		pattern := regexp.MustCompile(`^([a-fA-F0-9]{64})\s+[\*\.\/\\]*(.*)`)
		
		for scanner.Scan() {
			line := scanner.Text()
			matches := pattern.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			
			totalFiles++
			expectedHash := strings.ToLower(matches[1])
			fileName := strings.TrimSpace(matches[2])
			
			// Validate that the file path doesn't escape the base directory
			filePathOnMedia := filepath.Join(baseDir, fileName)
			cleanPath := filepath.Clean(filePathOnMedia)
			if !strings.HasPrefix(cleanPath, filepath.Clean(baseDir)) {
				fmt.Printf("Warning: Skipping potentially unsafe path: %s (referenced in %s)\n", fileName, filepath.Base(checksumFile))
				failedFiles++
				continue
			}
			
			if _, err := os.Stat(filePathOnMedia); os.IsNotExist(err) {
				fmt.Printf("Warning: File not found on media: %s (referenced in %s)\n", fileName, filepath.Base(checksumFile))
				failedFiles++
				continue
			}
			
			fmt.Printf("Verifying: %s", fileName)
			calculatedHash, err := getSha256Hash(filePathOnMedia)
			if err != nil {
				fmt.Printf(" -> \033[31mERROR: %v\033[0m\n", err)
				failedFiles++
				continue
			}
			
			calculatedHash = strings.ToLower(calculatedHash)
			if calculatedHash == expectedHash {
				fmt.Printf(" -> \033[32mOK\033[0m\n")
			} else {
				fmt.Printf(" -> \033[31mFAILED\033[0m\n")
				failedFiles++
			}
		}
		fmt.Println()  // Add blank line between checksum files
	}
	
	fmt.Println("--- Verification Summary ---")
	fmt.Printf("Checksum files processed: %d\n", len(checksumFiles))
	fmt.Printf("Total files verified: %d\n", totalFiles)
	if failedFiles == 0 && totalFiles > 0 {
		fmt.Printf("\033[32mSuccess: All %d files verified successfully.\033[0m\n", totalFiles)
	} else if totalFiles == 0 {
		fmt.Println("No files were verified.")
	} else {
		fmt.Printf("\033[31mFailure: %d out of %d files failed verification.\033[0m\n", failedFiles, totalFiles)
		hasErrors = true
	}
}

// findChecksumFiles recursively searches for ALL checksum files in the given directory tree.
// It finds files matching: *.sha, sha256sum.txt, or SHA256SUMS (case-insensitive).
// This ensures all checksum files on the media are discovered and processed.
func findChecksumFiles(rootPath string) ([]string, error) {
	var checksumFiles []string
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log permission errors but continue walking
			fmt.Fprintf(os.Stderr, "Warning: Could not access %s: %v\n", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		
		name := strings.ToLower(info.Name())
		if strings.HasSuffix(name, ".sha") || 
		   name == "sha256sum.txt" || 
		   name == "sha256sums" {
			checksumFiles = append(checksumFiles, path)
		}
		
		return nil
	})
	
	return checksumFiles, err
}

func verifyImplantedMD5(config *Config) {
	fmt.Println("\n--- Verifying Implanted ISO MD5 (checkisomd5 compatible) ---")
	
	result, err := checkImplantedMD5(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during MD5 check: %v\n", err)
		hasErrors = true
		return
	}
	
	if result == nil {
		fmt.Println("Warning: No 'ISO MD5SUM' signature found.")
		return
	}
	
	fmt.Printf("Verification Method: %s\n", result.VerificationMethod)
	fmt.Printf("Stored MD5:          %s\n", result.StoredMD5)
	fmt.Printf("Calculated MD5:      %s\n", result.CalculatedMD5)
	
	if result.IsIntegrityOK {
		fmt.Println("\n\033[32mSUCCESS: Implanted MD5 is valid.\033[0m")
	} else {
		fmt.Println("\n\033[31mFAILURE: Implanted MD5 does not match calculated hash.\033[0m")
		hasErrors = true
	}
}

type MD5Result struct {
	VerificationMethod string
	StoredMD5          string
	CalculatedMD5      string
	IsIntegrityOK      bool
}

func checkImplantedMD5(config *Config) (*MD5Result, error) {
	var file *os.File
	var err error
	var fileLength int64
	
	if config.isDrive {
		if runtime.GOOS == "windows" {
			devicePath := fmt.Sprintf("\\\\.\\%s:", config.driveLetter)
			file, err = os.Open(devicePath)
			if err != nil {
				return nil, err
			}
			
			// For device paths, we can't use file.Stat() reliably on 32-bit Windows
			// Instead, seek to end to get the size
			fileLength, err = file.Seek(0, io.SeekEnd)
			if err != nil {
				file.Close()
				// This typically happens with virtual/mounted drives (like mounted ISOs)
				// which don't support device-level operations
				return nil, fmt.Errorf("drive %s: does not support device-level access (likely a virtual/mounted drive).\n\n"+
					"Implanted MD5 check requires direct access to the ISO file.\n"+
					"To verify the implanted MD5, use the ISO file directly:\n"+
					"  Example: chkiso path\\to\\image.iso -md5\n\n"+
					"(Content verification will still work with the mounted drive)", config.driveLetter)
			}
			// Seek back to start
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				file.Close()
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("drive letters are only supported on Windows")
		}
	} else {
		file, err = os.Open(config.Path)
		if err != nil {
			return nil, err
		}
		
		// For regular files, we can use Stat safely
		fileInfo, err := file.Stat()
		if err != nil {
			file.Close()
			return nil, err
		}
		fileLength = fileInfo.Size()
	}
	
	defer file.Close()
	
	// Read PVD block
	pvdBlock := make([]byte, PVD_SIZE)
	if _, err := file.Seek(PVD_OFFSET, 0); err != nil {
		return nil, err
	}
	if n, err := file.Read(pvdBlock); err != nil || n != PVD_SIZE {
		return nil, fmt.Errorf("could not read PVD")
	}
	
	// Extract Application Use field
	appUseData := pvdBlock[APP_USE_OFFSET : APP_USE_OFFSET+APP_USE_SIZE]
	appUseString := string(appUseData)
	
	// Look for MD5 signature
	md5Pattern := regexp.MustCompile(`ISO MD5SUM = ([0-9a-fA-F]{32})`)
	matches := md5Pattern.FindStringSubmatch(appUseString)
	if matches == nil {
		return nil, nil
	}
	
	storedHash := strings.ToLower(matches[1])
	
	// Look for SKIPSECTORS
	skipSectors := 0
	skipPattern := regexp.MustCompile(`SKIPSECTORS\s*=\s*(\d+)`)
	if skipMatches := skipPattern.FindStringSubmatch(appUseString); skipMatches != nil {
		fmt.Sscanf(skipMatches[1], "%d", &skipSectors)
	}
	
	hashEndOffset := fileLength - int64(skipSectors*SECTOR_SIZE)
	
	// Create neutralized PVD (fill Application Use field with spaces)
	neutralizedPvd := make([]byte, len(pvdBlock))
	copy(neutralizedPvd, pvdBlock)
	for i := 0; i < APP_USE_SIZE; i++ {
		neutralizedPvd[APP_USE_OFFSET+i] = SPACE_CHAR
	}
	
	// Calculate MD5 hash
	hash := md5.New()
	
	// Part A: Read from start to PVD_OFFSET
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}
	if _, err := io.CopyN(hash, file, PVD_OFFSET); err != nil {
		return nil, err
	}
	
	// Part B: Add neutralized PVD
	hash.Write(neutralizedPvd)
	
	// Part C: Read from after PVD to hashEndOffset
	if _, err := file.Seek(PVD_OFFSET+PVD_SIZE, 0); err != nil {
		return nil, err
	}
	remaining := hashEndOffset - (PVD_OFFSET + PVD_SIZE)
	if _, err := io.CopyN(hash, file, remaining); err != nil {
		return nil, err
	}
	
	calculatedMD5 := hex.EncodeToString(hash.Sum(nil))
	
	return &MD5Result{
		VerificationMethod: "ASCII String (checkisomd5 compatible)",
		StoredMD5:          storedHash,
		CalculatedMD5:      strings.ToLower(calculatedMD5),
		IsIntegrityOK:      storedHash == strings.ToLower(calculatedMD5),
	}, nil
}

// mountISO mounts an ISO file on Windows using PowerShell's Mount-DiskImage
// Returns the drive letter (e.g., "H") and an error if mounting fails
func mountISO(isoPath string) (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("automatic ISO mounting is only supported on Windows")
	}
	
	// Get absolute path
	absPath, err := filepath.Abs(isoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %v", err)
	}
	
	// Mount the ISO and get the drive letter
	// Using PassThru to get the disk object, then Get-Volume to find the drive letter
	psCommand := fmt.Sprintf(`
		$disk = Mount-DiskImage -ImagePath '%s' -PassThru
		if ($disk) {
			$volume = Get-Volume -DiskImage $disk
			if ($volume) {
				$volume.DriveLetter
			}
		}
	`, absPath)
	
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCommand)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to mount ISO: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to mount ISO: %v", err)
	}
	
	driveLetter := strings.TrimSpace(string(output))
	if driveLetter == "" {
		return "", fmt.Errorf("failed to get drive letter after mounting")
	}
	
	return driveLetter, nil
}

// dismountISO dismounts an ISO file on Windows using PowerShell's Dismount-DiskImage
func dismountISO(isoPath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("automatic ISO dismounting is only supported on Windows")
	}
	
	// Get absolute path
	absPath, err := filepath.Abs(isoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	
	psCommand := fmt.Sprintf("Dismount-DiskImage -ImagePath '%s'", absPath)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to dismount ISO: %s", string(output))
	}
	
	return nil
}

func handleDismount(config *Config) {
	if config.isDrive {
		fmt.Printf("\nNote: Ejecting drives is not yet implemented in this version.\n")
		fmt.Printf("Please eject drive %s: manually.\n", config.driveLetter)
	} else if config.mountedISO {
		// Only dismount if we mounted it
		fmt.Printf("\nDismounting ISO...\n")
		if err := dismountISO(config.Path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to dismount ISO: %v\n", err)
			fmt.Printf("Please dismount %s manually.\n", config.Path)
		} else {
			fmt.Println("ISO dismounted successfully.")
		}
	} else {
		// ISO file but we didn't mount it
		fmt.Printf("\nNote: ISO was not mounted automatically.\n")
		if config.Path != "" {
			fmt.Printf("If you mounted %s manually, please dismount it manually.\n", config.Path)
		}
	}
}

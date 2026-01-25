package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var (
	uploadExternalFlag  bool
	uploadDirectoryFlag bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload <file> [<file>...]",
	Short: "Upload a file to PMP300",
	Long: `Upload an MP3 file from your computer to the PMP300 device.

By default, files are uploaded to the internal flash. Use --external to upload to a SmartMedia card.
The filename on the device will match the local filename.
Large files may take several minutes to upload.

Examples:
  pmp300 upload song.mp3
  pmp300 upload --external song.mp3
  pmp300 upload ~/Music/album/*.mp3
  pmp300 upload --directory`,
	Aliases: []string{"put", "push"},
	Args:    validateUploadArgs, // <-- Use custom validation
	RunE:    runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().BoolVar(&uploadExternalFlag, "external", false, "Upload to external SmartMedia card instead of internal flash")
	uploadCmd.Flags().BoolVar(&uploadDirectoryFlag, "directory", false, "Upload all files in the current directory")
}

func validateUploadArgs(cmd *cobra.Command, args []string) error {
	if uploadDirectoryFlag {
		if len(args) > 0 {
			return fmt.Errorf("cannot specify files when using --directory flag")
		}
	} else {
		if len(args) == 0 {
			return fmt.Errorf("requires at least 1 file argument")
		}
	}
	return nil
}

func runUpload(cmd *cobra.Command, args []string) error {
	device, err := getDevice()
	if err != nil {
		return err
	}

	var filesToUpload []string

	if uploadDirectoryFlag {
		if len(args) > 0 {
			return fmt.Errorf("cannot specify files when using --directory flag")
		}
		// Get current working directory
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		// Read all files in the directory
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", currentDir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() { // Only upload files, not subdirectories
				// Filter for MP3 files (case-insensitive)
				if strings.HasSuffix(strings.ToLower(entry.Name()), ".mp3") {
					filesToUpload = append(filesToUpload, filepath.Join(currentDir, entry.Name()))
				}
			}
		}
		if len(filesToUpload) == 0 {
			return fmt.Errorf("no files found in current directory to upload")
		}
	} else {
		if len(args) == 0 {
			return fmt.Errorf("no files or directory specified for upload")
		}
		// Existing glob pattern expansion logic
		for _, pattern := range args {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern %s: %w", pattern, err)
			}
			if len(matches) == 0 {
				return fmt.Errorf("no files match pattern: %s", pattern)
			}
			filesToUpload = append(filesToUpload, matches...)
		}
	}

	fmt.Printf("Connecting to %s...\n", device)

	port, err := arduino.Open(device)
	if err != nil {
		return fmt.Errorf("failed to open Arduino: %w", err)
	}
	defer port.Close()

	pmp := pmp300.New(port)

	if uploadExternalFlag {
		fmt.Println("Switching to external SmartMedia card...")
		if err := pmp.SwitchStorage(pmp300.StorageExternal); err != nil {
			return fmt.Errorf("failed to switch to external storage: %w", err)
		}
		// Check if external storage is present and formatted
		if present, err := pmp.DetectExternalStorage(); err != nil || !present {
			fmt.Println("No external SmartMedia card detected or card is unreadable.")
			fmt.Println("Please ensure a SmartMedia card is inserted and properly seated.")
			return fmt.Errorf("external SmartMedia card not found or unreadable")
		}
	} else {
		// Default to internal storage
		fmt.Println("Switching to internal flash...")
		if err := pmp.SwitchStorage(pmp300.StorageInternal); err != nil {
			return fmt.Errorf("failed to switch to internal storage: %w", err)
		}
	}

	fmt.Println("Initializing PMP300...")
	if err := pmp.Initialize(); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	fmt.Printf("Uploading files to %s...\n", pmp.GetCurrentStorage().String())

	// Upload each file
	for i, filePath := range filesToUpload { // <-- Updated loop variable
		fmt.Printf("\n[%d/%d] Uploading %s...\n", i+1, len(filesToUpload), filepath.Base(filePath))

		// Read file
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("  ✗ Failed to read file: %v\n", err)
			continue
		}

		sizeMB := float64(len(data)) / 1024.0 / 1024.0
		fmt.Printf("  Size: %.2f MB\n", sizeMB)

		// Upload with progress
		var lastProgress int
		err = pmp.UploadFile(filepath.Base(filePath), data, func(current, total int) {
			percent := (current * 100) / total
			if percent != lastProgress {
				fmt.Printf("\r  Progress: %d%%", percent)
				lastProgress = percent
			}
		})

		if err != nil {
			fmt.Printf("\n  ✗ Upload failed: %v\n", err)
			continue
		}

		fmt.Printf("\n  ✓ Upload complete\n")
	}

	fmt.Printf("\nUploaded %d file(s) successfully to %s.\n", len(filesToUpload), pmp.GetCurrentStorage().String())

	return nil
}

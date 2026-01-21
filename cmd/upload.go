package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a file to PMP300",
	Long: `Upload an MP3 file from your computer to the PMP300 device.

The filename on the device will match the local filename.
Large files may take several minutes to upload.

Examples:
  pmp300 upload song.mp3
  pmp300 upload ~/Music/album/*.mp3`,
	Aliases: []string{"put", "push"},
	Args:    cobra.MinimumNArgs(1),
	RunE:    runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, args []string) error {
	device, err := getDevice()
	if err != nil {
		return err
	}

	// Expand glob patterns
	var files []string
	for _, pattern := range args {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("no files match pattern: %s", pattern)
		}
		files = append(files, matches...)
	}

	fmt.Printf("Connecting to %s...\n", device)

	port, err := arduino.Open(device)
	if err != nil {
		return fmt.Errorf("failed to open Arduino: %w", err)
	}
	defer port.Close()

	pmp := pmp300.New(port)

	fmt.Println("Initializing PMP300...")
	if err := pmp.Initialize(); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Upload each file
	for i, filePath := range files {
		fmt.Printf("\n[%d/%d] Uploading %s...\n", i+1, len(files), filepath.Base(filePath))

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

	fmt.Printf("\nUploaded %d file(s) successfully.\n", len(files))

	return nil
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var (
	outputFlag string
)

var downloadCmd = &cobra.Command{
	Use:   "download <filename>",
	Short: "Download a file from PMP300",
	Long: `Download a file from the PMP300 device to your computer.

The file will be saved to the current directory unless --output is specified.

Examples:
  pmp300 download song.mp3
  pmp300 download song.mp3 --output ~/Music/song.mp3`,
	Aliases: []string{"get", "pull"},
	Args:    cobra.ExactArgs(1),
	RunE:    runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path (default: current directory)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	filename := args[0]

	device, err := getDevice()
	if err != nil {
		return err
	}

	// Determine output path
	outputPath := outputFlag
	if outputPath == "" {
		outputPath = filepath.Join(".", filename)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("output file already exists: %s (remove it first or use --output to specify different path)", outputPath)
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

	fmt.Printf("Downloading %s...\n", filename)

	var lastProgress int
	data, err := pmp.DownloadFile(filename, func(current, total int) {
		percent := (current * 100) / total
		if percent != lastProgress {
			fmt.Printf("\rProgress: %d%% (%d / %d bytes)", percent, current, total)
			lastProgress = percent
		}
	})
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	fmt.Println()

	// Write to file
	fmt.Printf("Writing to %s...\n", outputPath)
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ Downloaded %d bytes to %s\n", len(data), outputPath)

	return nil
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	outputFlag string

	allFlag bool
)

var downloadCmd = &cobra.Command{
	Use:   "download [<filename>]",
	Short: "Download a file or all files from PMP300",
	Long: `Download a file or all files from the PMP300 device to your computer.

When downloading a single file, it will be saved to the current directory unless --output is specified.
When downloading all files, a directory can be specified with --output, otherwise a 'pmp300-download' directory will be created.

Examples:
  pmp300 download song.mp3
  pmp300 download song.mp3 --output ~/Music/song.mp3
  pmp300 download --all
  pmp300 download --all --output ~/Music/pmp300-backup`,
	Aliases: []string{"get", "pull"},
	Args:    cobra.MaximumNArgs(1),
	RunE:    runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path (for single file) or directory (for --all)")
	downloadCmd.Flags().BoolVar(&allFlag, "all", false, "Download all files")
}

func runDownload(cmd *cobra.Command, args []string) error {

	if allFlag {
		if len(args) > 0 {
			return fmt.Errorf("cannot specify a filename when using --all")
		}
		return runDownloadAll(cmd, args)
	}

	if len(args) == 0 {
		return fmt.Errorf("no filename specified")
	}

	filename := args[0]

	// Determine output path
	outputPath := outputFlag
	if outputPath == "" {
		outputPath = filepath.Join(".", filename)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("output file already exists: %s (remove it first or use --output to specify different path)", outputPath)
	}

	pmp, port, err := getInitializedPMPDevice()
	if err != nil {
		return err
	}
	defer port.Close()

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

func runDownloadAll(cmd *cobra.Command, args []string) error {
	// Determine output directory
	outputDir := outputFlag
	if outputDir == "" {
		outputDir = "pmp300-download"
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	pmp, port, err := getInitializedPMPDevice()
	if err != nil {
		return err
	}
	defer port.Close()

	fmt.Printf("Reading file list from %s...\n", pmp.GetCurrentStorage())
	files, err := pmp.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("No files on %s to download.\n", pmp.GetCurrentStorage())
		return nil
	}

	fmt.Printf("Found %d file(s). Downloading all to %s...\n\n", len(files), outputDir)

	for i, file := range files {
		filename := file.Name
		outputPath := filepath.Join(outputDir, filename)

		// Check if output file exists
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("[%d/%d] Skipping %s, file already exists.\n", i+1, len(files), filename)
			continue
		}

		fmt.Printf("[%d/%d] Downloading %s...\n", i+1, len(files), filename)

		var lastProgress int
		data, err := pmp.DownloadFile(filename, func(current, total int) {
			percent := (current * 100) / total
			if percent != lastProgress {
				fmt.Printf("\rProgress: %d%% (%d / %d bytes)", percent, current, total)
				lastProgress = percent
			}
		})
		if err != nil {
			fmt.Printf("\nDownload failed for %s: %v\n", filename, err)
			continue // Continue to the next file
		}
		fmt.Println()

		// Write to file
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			fmt.Printf("Failed to write file %s: %v\n", outputPath, err)
			continue // Continue to the next file
		}

		fmt.Printf("✓ Downloaded %d bytes to %s\n\n", len(data), outputPath)
	}

	fmt.Println("All downloads complete.")
	return nil
}

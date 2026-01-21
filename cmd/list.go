package cmd

import (
	"fmt"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var (
	verboseFlag bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List files on PMP300",
	Long: `List all files currently stored on the PMP300 device.

Shows filename, size, and timestamp for each file.
Use --verbose for additional details including block positions.`,
	Aliases: []string{"ls"},
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show detailed information")
}

func runList(cmd *cobra.Command, args []string) error {
	device, err := getDevice()
	if err != nil {
		return err
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

	fmt.Printf("Reading file list from %s...\n", pmp.GetCurrentStorage())
	files, err := pmp.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("No files on %s.\n", pmp.GetCurrentStorage())
		fmt.Println("\nTip: Use 'pmp300 storage list' to check external SmartMedia card")
		return nil
	}

	fmt.Printf("\nFound %d file(s) on %s:\n\n", len(files), pmp.GetCurrentStorage())

	if verboseFlag {
		// Verbose output
		fmt.Println("  # | Name                          | Size      | Blocks  | Position | Timestamp")
		fmt.Println("----+-------------------------------+-----------+---------+----------+-------------------")
		for i, file := range files {
			timestamp := file.Timestamp.Format("2006-01-02 15:04:05")
			if file.Timestamp.IsZero() {
				timestamp = "Unknown"
			}
			fmt.Printf("%3d | %-29s | %9d | %7d | %8d | %s\n",
				i+1, truncate(file.Name, 29), file.Size, file.BlockCount, file.BlockPosition, timestamp)
		}
	} else {
		// Simple output
		fmt.Println("  # | Name                          | Size      | Timestamp")
		fmt.Println("----+-------------------------------+-----------+-------------------")
		for i, file := range files {
			timestamp := file.Timestamp.Format("2006-01-02 15:04:05")
			if file.Timestamp.IsZero() {
				timestamp = "Unknown"
			}
			sizeMB := float64(file.Size) / 1024.0 / 1024.0
			fmt.Printf("%3d | %-29s | %7.2f MB | %s\n",
				i+1, truncate(file.Name, 29), sizeMB, timestamp)
		}
	}

	// Show device info
	fmt.Println()
	info, err := pmp.GetDeviceInfo()
	if err == nil {
		usedBlocks := info.TotalBlocks - info.FreeBlocks - info.BadBlocks
		usedMB := float64(usedBlocks*32) / 1024.0
		freeMB := float64(info.FreeBlocks*32) / 1024.0
		totalMB := float64(info.TotalBlocks*32) / 1024.0

		fmt.Printf("Storage: %.1f MB used / %.1f MB free (%.1f MB total)\n", usedMB, freeMB, totalMB)
		if info.BadBlocks > 0 {
			fmt.Printf("Bad blocks: %d\n", info.BadBlocks)
		}
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

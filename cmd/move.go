package cmd

import (
	"fmt"
	"strconv"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <from> <to>",
	Short: "Move a file to a new position in playback order",
	Long: `Change the playback order by moving a file from one position to another.

Positions are 1-based (first file is position 1).
All files between the old and new positions will shift accordingly.

Examples:
  pmp300 move 3 1      # Move 3rd file to 1st position
  pmp300 move 1 5      # Move 1st file to 5th position`,
	Args: cobra.ExactArgs(2),
	RunE: runMove,
}

func init() {
	rootCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
	// Parse positions (convert from 1-based to 0-based)
	from, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid from position: %s", args[0])
	}
	from-- // Convert to 0-based

	to, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid to position: %s", args[1])
	}
	to-- // Convert to 0-based

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

	// Get current file list to show what we're moving
	files, err := pmp.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if from >= len(files) {
		return fmt.Errorf("from position %d out of range (have %d files)", from+1, len(files))
	}
	if to >= len(files) {
		return fmt.Errorf("to position %d out of range (have %d files)", to+1, len(files))
	}

	fromFile := files[from].Name
	fmt.Printf("Moving '%s' from position %d to position %d...\n", fromFile, from+1, to+1)

	// Perform move
	if err := pmp.MoveFile(from, to); err != nil {
		return fmt.Errorf("move failed: %w", err)
	}

	fmt.Println("✓ File order updated")

	// Show new order
	fmt.Println("\nNew playback order:")
	files, _ = pmp.ListFiles()
	for i, file := range files {
		marker := ""
		if i == to {
			marker = " ◄"
		}
		fmt.Printf("  %2d. %s%s\n", i+1, file.Name, marker)
	}

	return nil
}

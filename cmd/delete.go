package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var (
	deleteAllFlag bool
	forceFlag     bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <filename>",
	Short: "Delete a file from PMP300",
	Long: `Delete one or more files from the PMP300 device.

Use --all to delete all files from the device.
Use --force to skip confirmation prompts.

Examples:
  pmp300 delete song.mp3
  pmp300 delete --all
  pmp300 delete --all --force`,
	Aliases: []string{"rm", "remove"},
	RunE:    runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVar(&deleteAllFlag, "all", false, "Delete all files")
	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation prompts")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if !deleteAllFlag && len(args) == 0 {
		return fmt.Errorf("specify filename to delete or use --all")
	}

	if deleteAllFlag && len(args) > 0 {
		return fmt.Errorf("cannot specify filename with --all")
	}

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

	if deleteAllFlag {
		return deleteAll(pmp)
	}

	// Delete individual files
	for _, filename := range args {
		if err := deleteOne(pmp, filename); err != nil {
			fmt.Printf("✗ Failed to delete %s: %v\n", filename, err)
		}
	}

	return nil
}

func deleteOne(pmp *pmp300.Device, filename string) error {
	// Confirm deletion unless force flag is set
	if !forceFlag {
		fmt.Printf("Delete %s? (y/N): ", filename)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Skipped.")
			return nil
		}
	}

	fmt.Printf("Deleting %s...\n", filename)
	if err := pmp.DeleteFile(filename); err != nil {
		return err
	}

	fmt.Printf("✓ Deleted %s\n", filename)
	return nil
}

func deleteAll(pmp *pmp300.Device) error {
	// Read file count first
	info, err := pmp.GetDeviceInfo()
	if err != nil {
		return fmt.Errorf("failed to get device info: %w", err)
	}

	if info.EntryCount == 0 {
		fmt.Println("No files to delete.")
		return nil
	}

	// Confirm deletion unless force flag is set
	if !forceFlag {
		fmt.Printf("Delete ALL %d files? This cannot be undone! (y/N): ", info.EntryCount)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	fmt.Println("Deleting all files...")
	if err := pmp.DeleteAllFiles(); err != nil {
		return fmt.Errorf("failed to delete all files: %w", err)
	}

	fmt.Printf("✓ Deleted all %d files\n", info.EntryCount)
	return nil
}

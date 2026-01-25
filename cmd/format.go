package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	checkBadBlocksFlag bool
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format/initialize the PMP300 device",
	Long: `Format the PMP300 device, erasing all files and creating a fresh directory.

WARNING: This will DELETE ALL FILES on the device!

By default, this formats the internal flash. Use --external to format an external SmartMedia card.

Use --check-bad-blocks to perform bad block detection during format.
This is recommended for new or problematic devices, but takes a very long time.

Examples:
  pmp300 format
  pmp300 format --external
  pmp300 format --check-bad-blocks`,
	RunE: runFormat,
}

func init() {
	rootCmd.AddCommand(formatCmd)
	formatCmd.Flags().BoolVar(&checkBadBlocksFlag, "check-bad-blocks", false, "Check for bad blocks (slow)")
}

func runFormat(cmd *cobra.Command, args []string) error {
	// Confirm format unless force flag is set
	if !forceFlag {
		targetStorage := "internal flash"
		if externalFlag { // externalFlag is now global
			targetStorage = "external SmartMedia card"
		}
		fmt.Printf("WARNING: This will ERASE ALL FILES on the PMP300's %s!\n", targetStorage)
		fmt.Print("Are you sure you want to format the device? (y/N): ")
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

	pmp, port, err := getInitializedPMPDevice()
	if err != nil {
		return err
	}
	defer port.Close()

	if externalFlag { // externalFlag is now global
		// Check if external storage is present and formatted
		present, err := pmp.DetectExternalStorage()

		if err != nil { // This means "present but corrupted"
			fmt.Printf("Warning: %v. Attempting to format anyway.\n", err)
		} else if !present { // This means "not present/unreadable at low level"
			return fmt.Errorf("external SmartMedia card not found or unreadable")
		}
	}

	fmt.Printf("Formatting %s...\n", pmp.GetCurrentStorage().String())
	if checkBadBlocksFlag {
		fmt.Println("Bad block checking enabled - this will take a VERY long time!")
	}

	if err := pmp.FormatDevice(checkBadBlocksFlag); err != nil {
		return fmt.Errorf("format failed: %w", err)
	}

	fmt.Println("\n✓ Format complete!")

	// Show device info
	info, err := pmp.GetDeviceInfo()
	if err != nil {
		fmt.Printf("Warning: Could not get device info after format (checksum or parsing error): %v\n", err)
		// Try to proceed with what info we have, if any was returned
		if info != nil {
			totalMB := float64(info.BlocksAvailable) * 32.0 / 1024.0
			fmt.Printf("Device ready: %.1f MB, %d blocks\n", totalMB, info.BlocksAvailable)
			if info.BlocksBad > 0 {
				fmt.Printf("Bad blocks found: %d\n", info.BlocksBad)
			}
		}
	} else {
		totalMB := float64(info.BlocksAvailable) * 32.0 / 1024.0
		fmt.Printf("Device ready: %.1f MB, %d blocks\n", totalMB, info.BlocksAvailable)
		if info.BlocksBad > 0 {
			fmt.Printf("Bad blocks found: %d\n", info.BlocksBad)
		}
	}

	return nil
}

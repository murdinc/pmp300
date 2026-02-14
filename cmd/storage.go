package cmd

import (
	"fmt"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage PMP300 storage (internal flash / external SmartMedia)",
	Long: `View and switch between internal flash and external SmartMedia card storage.

The PMP300 can use either:
  - Internal flash (32MB or 64MB on SE models)
  - External SmartMedia card (8MB to 128MB)

Each storage device has its own directory and files. They don't mix.
Think of it like swapping CDs - you see one at a time.`,
}

var storageListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show available storage devices",
	Long:  `Display which storage devices are available and which is currently active.`,
	RunE:  runStorageList,
}


func init() {
	rootCmd.AddCommand(storageCmd)
	storageCmd.AddCommand(storageListCmd)
}

func runStorageList(cmd *cobra.Command, args []string) error {
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

	// Check current storage
	current := pmp.GetCurrentStorage()
	fmt.Printf("\nCurrent storage: %s\n\n", current)

	// Try to get info from internal storage
	fmt.Println("Checking internal flash...")
	if current != pmp300.StorageInternal {
		pmp.SwitchStorage(pmp300.StorageInternal)
	}

	internalInfo, internalErr := pmp.GetDeviceInfo()
	if internalErr == nil {
		totalMB := float64(internalInfo.BlocksAvailable) * 32.0 / 1024.0
		usedMB := float64(internalInfo.BlocksUsed) * 32.0 / 1024.0
		fmt.Printf("  ✓ Internal Flash: %.1f MB (%.1f MB used, %d files)\n", totalMB, usedMB, internalInfo.EntryCount)
	} else {
		fmt.Printf("  ✗ Internal Flash: Not accessible\n")
	}

	// Try to detect external SmartMedia
	fmt.Println("\nChecking external SmartMedia...")
	hasExternal, _ := pmp.DetectExternalStorage()

	if hasExternal {
		pmp.SwitchStorage(pmp300.StorageExternal)
		externalInfo, externalErr := pmp.GetDeviceInfo()
		if externalErr == nil {
			totalMB := float64(externalInfo.BlocksAvailable) * 32.0 / 1024.0
			usedMB := float64(externalInfo.BlocksUsed) * 32.0 / 1024.0
			fmt.Printf("  ✓ External SmartMedia: %.1f MB (%.1f MB used, %d files)\n", totalMB, usedMB, externalInfo.EntryCount)
		} else {
			fmt.Printf("  ✓ External SmartMedia: Detected but unformatted\n")
		}
	} else {
		fmt.Printf("  ✗ External SmartMedia: Not detected\n")
	}

	// Switch back to original
	pmp.SwitchStorage(current)

	fmt.Println("\nUse --external flag with commands to access SmartMedia card.")

	return nil
}


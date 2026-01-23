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

var storageSwitchCmd = &cobra.Command{
	Use:   "switch <internal|external>",
	Short: "Switch to internal or external storage",
	Long: `Switch between internal flash and external SmartMedia card.

Examples:
  pmp300 storage switch internal     # Use internal flash
  pmp300 storage switch external     # Use external SmartMedia card`,
	Args: cobra.ExactArgs(1),
	RunE: runStorageSwitch,
}

func init() {
	rootCmd.AddCommand(storageCmd)
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageSwitchCmd)
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
		totalMB := float64(internalInfo.TotalBlocks*32) / 1024.0
		usedBlocks := internalInfo.TotalBlocks - internalInfo.FreeBlocks - internalInfo.BlocksBad
		usedMB := float64(usedBlocks*32) / 1024.0
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
			totalMB := float64(externalInfo.TotalBlocks*32) / 1024.0
			usedBlocks := externalInfo.TotalBlocks - externalInfo.FreeBlocks - externalInfo.BlocksBad
			usedMB := float64(usedBlocks*32) / 1024.0
			fmt.Printf("  ✓ External SmartMedia: %.1f MB (%.1f MB used, %d files)\n", totalMB, usedMB, externalInfo.EntryCount)
		} else {
			fmt.Printf("  ✓ External SmartMedia: Detected but unformatted\n")
		}
	} else {
		fmt.Printf("  ✗ External SmartMedia: Not detected\n")
	}

	// Switch back to original
	pmp.SwitchStorage(current)

	fmt.Println("\nUse 'pmp300 storage switch <internal|external>' to change active storage.")

	return nil
}

func runStorageSwitch(cmd *cobra.Command, args []string) error {
	storageType := args[0]

	var storage pmp300.StorageType
	switch storageType {
	case "internal":
		storage = pmp300.StorageInternal
	case "external":
		storage = pmp300.StorageExternal
	default:
		return fmt.Errorf("invalid storage type: %s (must be 'internal' or 'external')", storageType)
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

	fmt.Printf("Switching to %s...\n", storage)
	if err := pmp.SwitchStorage(storage); err != nil {
		return fmt.Errorf("switch failed: %w", err)
	}

	// Verify by trying to read directory
	info, err := pmp.GetDeviceInfo()
	if err != nil {
		return fmt.Errorf("switched, but cannot read storage: %w\nMake sure storage is formatted and accessible", err)
	}

	fmt.Printf("✓ Switched to %s\n\n", storage)

	totalMB := float64(info.TotalBlocks*32) / 1024.0
	usedBlocks := info.TotalBlocks - info.FreeBlocks - info.BlocksBad
	usedMB := float64(usedBlocks*32) / 1024.0

	fmt.Printf("Storage info:\n")
	fmt.Printf("  Capacity: %.1f MB\n", totalMB)
	fmt.Printf("  Used:     %.1f MB\n", usedMB)
	fmt.Printf("  Files:    %d\n", info.EntryCount)

	fmt.Println("\nAll subsequent commands will use this storage until you switch again.")

	return nil
}

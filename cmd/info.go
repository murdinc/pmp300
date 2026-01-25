package cmd

import (
	"fmt"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display PMP300 device information",
	Long: `Display detailed information about the PMP300 device including:
  - Total storage capacity
  - Free/used space
  - Number of files
  - Bad block count
  - Protocol version`,
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
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

	// Get Arduino version
	version, err := port.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get Arduino version: %w", err)
	}

	pmp := pmp300.New(port)

	fmt.Println("Initializing PMP300...")
	if err := pmp.Initialize(); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	fmt.Println("Reading device information...")
	info, err := pmp.GetDeviceInfo()
	if err != nil {
		return fmt.Errorf("failed to get device info: %w", err)
	}

	fmt.Println("\n=== PMP300 Device Information ===")

	// Storage information - C++ fields: BlocksAvailable=total, BlocksRemaining=free, BlocksUsed=used
	usedMB := float64(info.BlocksUsed) * 32.0 / 1024.0
	freeMB := float64(info.BlocksRemaining) * 32.0 / 1024.0
	totalMB := float64(info.BlocksAvailable) * 32.0 / 1024.0
	usedPercent := (float64(info.BlocksUsed) / float64(info.BlocksAvailable)) * 100

	fmt.Printf("Model:          Diamond Rio PMP300\n")
	if totalMB > 40 {
		fmt.Printf("Variant:        PMP300 SE (64MB)\n")
	} else {
		fmt.Printf("Variant:        PMP300 (32MB)\n")
	}

	fmt.Printf("\nStorage:\n")
	fmt.Printf("  Total:        %.1f MB (%d blocks)\n", totalMB, info.BlocksAvailable)
	fmt.Printf("  Used:         %.1f MB (%d blocks, %.1f%%)\n", usedMB, info.BlocksUsed, usedPercent)
	fmt.Printf("  Free:         %.1f MB (%d blocks)\n", freeMB, info.BlocksRemaining)
	if info.BlocksBad > 0 {
		badMB := float64(info.BlocksBad*32) / 1024.0
		fmt.Printf("  Bad blocks:   %.1f MB (%d blocks)\n", badMB, info.BlocksBad)
	}

	fmt.Printf("\nFiles:\n")
	fmt.Printf("  Count:        %d / %d\n", info.EntryCount, pmp300.MAX_ENTRIES)

	fmt.Printf("\nProtocol:\n")
	fmt.Printf("  Version:      %d\n", info.Version)
	fmt.Printf("  Block size:   32 KB\n")

	fmt.Printf("\nBridge:\n")
	fmt.Printf("  Device:       %s\n", port.Device())
	fmt.Printf("  Firmware:     v%s\n", version)

	fmt.Printf("\nStorage:\n")
	fmt.Printf("  Active:       %s\n", pmp.GetCurrentStorage())
	fmt.Println("  Tip: Use 'pmp300 storage list' to see all storage options")

	return nil
}

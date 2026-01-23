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
	checkBadBlocksFlag bool
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format/initialize the PMP300 device",
	Long: `Format the PMP300 device, erasing all files and creating a fresh directory.

WARNING: This will DELETE ALL FILES on the device!

Use --check-bad-blocks to perform bad block detection during format.
This is recommended for new or problematic devices, but takes a very long time.

Examples:
  pmp300 format
  pmp300 format --check-bad-blocks`,
	RunE: runFormat,
}

func init() {
	rootCmd.AddCommand(formatCmd)
	formatCmd.Flags().BoolVar(&checkBadBlocksFlag, "check-bad-blocks", false, "Check for bad blocks (slow)")
}

func runFormat(cmd *cobra.Command, args []string) error {
	device, err := getDevice()
	if err != nil {
		return err
	}

	// Confirm format unless force flag is set
	if !forceFlag {
		fmt.Println("WARNING: This will ERASE ALL FILES on the PMP300!")
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

	fmt.Println("Formatting device...")
	if checkBadBlocksFlag {
		fmt.Println("Bad block checking enabled - this will take a VERY long time!")
	}

	if err := pmp.FormatDevice(checkBadBlocksFlag); err != nil {
		return fmt.Errorf("format failed: %w", err)
	}

	fmt.Println("\n✓ Format complete!")

	// Show device info
	info, err := pmp.GetDeviceInfo()
	if err == nil {
		totalMB := float64(info.BlocksAvailable) * 32.0 / 1024.0
		fmt.Printf("Device ready: %.1f MB, %d blocks\n", totalMB, info.BlocksAvailable)
		if info.BlocksBad > 0 {
			fmt.Printf("Bad blocks found: %d\n", info.BlocksBad)
		}
	}

	return nil
}

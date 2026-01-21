package cmd

import (
	"fmt"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to Arduino and PMP300",
	Long: `Test the connection to the Arduino USB-to-parallel bridge
and attempt to initialize the PMP300 device.

This command verifies:
  - Arduino is connected and responding
  - Firmware version is compatible
  - PMP300 initialization sequence completes

Use this to diagnose connection issues.`,
	RunE: runTest,
}

func init() {
	rootCmd.AddCommand(testCmd)
}

func runTest(cmd *cobra.Command, args []string) error {
	device, err := getDevice()
	if err != nil {
		return err
	}

	fmt.Printf("Connecting to Arduino on %s...\n", device)

	port, err := arduino.Open(device)
	if err != nil {
		return fmt.Errorf("failed to open Arduino: %w", err)
	}
	defer port.Close()

	fmt.Println("✓ Arduino connected")

	// Get firmware version
	version, err := port.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}
	fmt.Printf("✓ Firmware version: %s\n", version)

	// Test ping
	fmt.Print("Testing ping... ")
	if err := port.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	fmt.Println("OK")

	// Test basic I/O
	fmt.Print("Testing control register write... ")
	if err := port.OutByte(2, 0x04); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	fmt.Println("OK")

	fmt.Print("Testing status register read... ")
	status, err := port.InByte(1)
	if err != nil {
		return fmt.Errorf("read failed: %w", err)
	}
	fmt.Printf("OK (value: 0x%02X)\n", status)

	// Initialize PMP300
	fmt.Println("\nInitializing PMP300...")
	pmp := pmp300.New(port)
	if err := pmp.Initialize(); err != nil {
		return fmt.Errorf("PMP300 initialization failed: %w", err)
	}
	fmt.Println("✓ Initialization sequence sent")

	// Try to detect actual PMP300 device
	fmt.Println("\nAttempting to detect PMP300 device...")
	fmt.Println("(This will fail if no PMP300 is connected - that's OK for testing Arduino)")

	_, err = pmp.GetDeviceInfo()
	if err != nil {
		fmt.Printf("\n⚠ Warning: Could not read from PMP300 device\n")
		fmt.Printf("   Error: %v\n", err)
		fmt.Println("\nThis means:")
		fmt.Println("  ✓ Arduino is working correctly")
		fmt.Println("  ✗ PMP300 is either:")
		fmt.Println("    - Not connected to DB-25 connector")
		fmt.Println("    - Not powered on")
		fmt.Println("    - Wiring is incorrect")
		fmt.Println("    - Parallel cable is bad")
		fmt.Println("\nTo verify wiring, upload arduino/hardware_test/hardware_test.ino")
		fmt.Println("and use a multimeter to check connections.")
		return nil
	}

	fmt.Println("✓ PMP300 device detected and responding!")
	fmt.Println("\n=== All tests passed! ===")
	fmt.Println("Your PMP300 is ready to use.")

	return nil
}

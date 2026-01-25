package cmd

import (
	"fmt"
	"os"

	"github.com/murdinc/pmp300/pkg/arduino"
	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var (
	deviceFlag   string
	externalFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "pmp300",
	Short: "CLI tool for Diamond Rio PMP300 MP3 player",
	Long: `pmp300 is a command-line interface for interacting with the
Diamond Rio PMP300 MP3 player via an Arduino USB-to-parallel bridge.

The PMP300 was one of the first portable MP3 players, released in 1998.
This tool allows you to manage files, view device information, and more
on modern computers without native parallel ports.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flag for serial device
	rootCmd.PersistentFlags().StringVarP(&deviceFlag, "device", "d", "", "Serial device (e.g., /dev/cu.usbmodem14201)")
	rootCmd.PersistentFlags().BoolVar(&externalFlag, "external", false, "Use external storage for operations")
}

// getInitializedPMPDevice returns an initialized PMP300 device with storage set and the opened Arduino port
func getInitializedPMPDevice() (*pmp300.Device, *arduino.Port, error) {
	devPath, err := getDevice()
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("Connecting to %s...\n", devPath)

	port, err := arduino.Open(devPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open Arduino: %w", err)
	}

	pmp := pmp300.New(port)

	if externalFlag {
		fmt.Println("Switching to external storage...")
		if err := pmp.SwitchStorage(pmp300.StorageExternal); err != nil {
			port.Close()
			return nil, nil, fmt.Errorf("failed to switch to external storage: %w", err)
		}
		// Call CheckPresent here to populate d.externalBlockCount
		_, _, err := pmp.CheckPresent()
		if err != nil {
			port.Close()
			return nil, nil, fmt.Errorf("failed to detect device after switching to external storage: %w", err)
		}
	} else {
		// Default to internal storage (even if not explicitly set, ensure consistency)
		fmt.Println("Switching to internal storage...")
		if err := pmp.SwitchStorage(pmp300.StorageInternal); err != nil {
			port.Close()
			return nil, nil, fmt.Errorf("failed to switch to internal storage: %w", err)
		}
		// For internal, also call CheckPresent to set d.specialEdition if applicable
		_, _, err := pmp.CheckPresent()
		if err != nil {
			port.Close()
			return nil, nil, fmt.Errorf("failed to detect device after switching to internal storage: %w", err)
		}
	}

	fmt.Println("Initializing PMP300...")
	if err := pmp.Initialize(); err != nil {
		port.Close() // Close port on initialization failure
		return nil, nil, fmt.Errorf("initialization failed: %w", err)
	}

	return pmp, port, nil
}

// getDevice returns the device path, checking environment variable if not set
func getDevice() (string, error) {
	if deviceFlag != "" {
		return deviceFlag, nil
	}

	// Check environment variable
	if dev := os.Getenv("PMP300_DEVICE"); dev != "" {
		return dev, nil
	}

	return "", fmt.Errorf("device not specified. Use --device flag or set PMP300_DEVICE environment variable")
}

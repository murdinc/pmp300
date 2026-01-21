package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	deviceFlag string
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

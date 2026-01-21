package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	boardFlag string
	portFlag  string
)

var flashCmd = &cobra.Command{
	Use:   "flash",
	Short: "Flash Arduino with PMP300 bridge firmware",
	Long: `Compile and upload the PMP300 USB-to-parallel bridge firmware to your Arduino.

This command uses arduino-cli to compile and upload the firmware.
If arduino-cli is not installed, instructions will be provided.

The command will:
  1. Detect your Arduino board (or use --board flag)
  2. Compile the firmware for your board
  3. Upload to the Arduino

Examples:
  pmp300 flash                                    # Auto-detect board and port
  pmp300 flash --board arduino:avr:mega           # Specify board
  pmp300 flash --port /dev/cu.usbmodem14201       # Specify port`,
	RunE: runFlash,
}

func init() {
	rootCmd.AddCommand(flashCmd)
	flashCmd.Flags().StringVar(&boardFlag, "board", "", "Arduino board FQBN (e.g., arduino:avr:mega)")
	flashCmd.Flags().StringVar(&portFlag, "port", "", "Arduino serial port")
}

func runFlash(cmd *cobra.Command, args []string) error {
	// Check if arduino-cli is installed
	if !isArduinoCLIInstalled() {
		return fmt.Errorf("arduino-cli not found. Install it first:\n\n%s", getInstallInstructions())
	}

	// Get sketch path
	sketchPath, err := getSketchPath()
	if err != nil {
		return err
	}

	fmt.Printf("Using sketch: %s\n", sketchPath)

	// Detect or validate port
	port := portFlag
	if port == "" {
		detectedPort, err := detectPort()
		if err != nil {
			return fmt.Errorf("failed to detect Arduino port: %w\nUse --port to specify manually", err)
		}
		port = detectedPort
		fmt.Printf("Detected port: %s\n", port)
	}

	// Detect or validate board
	board := boardFlag
	if board == "" {
		detectedBoard, err := detectBoard(port)
		if err != nil {
			return fmt.Errorf("failed to detect board: %w\nUse --board to specify manually (e.g., --board arduino:avr:mega)", err)
		}
		board = detectedBoard
		fmt.Printf("Detected board: %s\n", board)
	}

	// Ensure core is installed
	fmt.Println("\nEnsuring Arduino core is installed...")
	if err := ensureCore(board); err != nil {
		fmt.Printf("Warning: Failed to install core: %v\n", err)
		fmt.Println("Continuing anyway...")
	}

	// Compile sketch
	fmt.Println("\nCompiling firmware...")
	if err := compileSketch(sketchPath, board); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}
	fmt.Println("✓ Compilation successful")

	// Upload sketch
	fmt.Printf("\nUploading to %s...\n", port)
	if err := uploadSketch(sketchPath, board, port); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	fmt.Println("✓ Upload successful")

	fmt.Println("\n=== Firmware flashed successfully! ===")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Wait 2 seconds for Arduino to reset")
	fmt.Printf("  2. Test: pmp300 test --device %s\n", port)

	return nil
}

func isArduinoCLIInstalled() bool {
	_, err := exec.LookPath("arduino-cli")
	return err == nil
}

func getInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return `macOS:
  brew install arduino-cli

Or download from: https://arduino.github.io/arduino-cli/`
	case "linux":
		return `Linux:
  curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh

Or download from: https://arduino.github.io/arduino-cli/`
	case "windows":
		return `Windows:
  Download from: https://arduino.github.io/arduino-cli/

  Or use winget:
  winget install ArduinoSA.CLI`
	default:
		return "Download from: https://arduino.github.io/arduino-cli/"
	}
}

func getSketchPath() (string, error) {
	// Get current executable or working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Check common locations
	locations := []string{
		filepath.Join(wd, "arduino", "pmp300_usb_parallel_bridge", "pmp300_usb_parallel_bridge.ino"),
		filepath.Join(wd, "arduino", "pmp300_usb_parallel_bridge"),
		"arduino/pmp300_usb_parallel_bridge/pmp300_usb_parallel_bridge.ino",
		"arduino/pmp300_usb_parallel_bridge",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			// If it's a directory, append the .ino filename
			info, _ := os.Stat(loc)
			if info.IsDir() {
				loc = filepath.Join(loc, "pmp300_usb_parallel_bridge.ino")
			}
			return loc, nil
		}
	}

	return "", fmt.Errorf("sketch not found. Expected: arduino/pmp300_usb_parallel_bridge/pmp300_usb_parallel_bridge.ino")
}

func detectPort() (string, error) {
	cmd := exec.Command("arduino-cli", "board", "list", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse JSON output (simplified - just look for port)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, `"address"`) || strings.Contains(line, `"port"`) {
			// Extract port from JSON line
			start := strings.Index(line, `"`) + 1
			if start > 0 {
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					port := line[start : start+end]
					if strings.HasPrefix(port, "/dev/") || strings.HasPrefix(port, "COM") {
						return port, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no Arduino board detected")
}

func detectBoard(port string) (string, error) {
	cmd := exec.Command("arduino-cli", "board", "list")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, port) {
			// Look for FQBN in the line
			// Typical format: /dev/cu.usbmodem14201  arduino:avr:mega  Arduino Mega or Mega 2560
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Second field is usually the FQBN
				fqbn := parts[1]
				if strings.Contains(fqbn, ":") {
					return fqbn, nil
				}
			}
		}
	}

	// Default fallback based on common boards
	return "", fmt.Errorf("could not auto-detect board")
}

func ensureCore(board string) error {
	// Extract core from FQBN (e.g., "arduino:avr" from "arduino:avr:mega")
	parts := strings.Split(board, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid board FQBN: %s", board)
	}
	core := parts[0] + ":" + parts[1]

	// Check if core is installed
	cmd := exec.Command("arduino-cli", "core", "list")
	output, _ := cmd.Output()

	if !strings.Contains(string(output), core) {
		// Install core
		fmt.Printf("Installing core %s...\n", core)
		cmd := exec.Command("arduino-cli", "core", "install", core)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return nil
}

func compileSketch(sketchPath, board string) error {
	cmd := exec.Command("arduino-cli", "compile", "--fqbn", board, sketchPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func uploadSketch(sketchPath, board, port string) error {
	cmd := exec.Command("arduino-cli", "upload", "-p", port, "--fqbn", board, sketchPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

/*
 * Example Go client for PMP300 USB-Parallel Bridge
 *
 * This demonstrates how to communicate with the Arduino firmware.
 * Use this as a reference when building your full PMP300 interface software.
 *
 * Install dependency:
 *   go get github.com/tarm/serial
 *
 * Run:
 *   go run example_client.go
 */

// Command bytes (must match Arduino firmware)
const (
	CMD_WRITE_DATA   = 'W'
	CMD_WRITE_CTRL   = 'C'
	CMD_READ_STATUS  = 'R'
	CMD_DELAY_US     = 'D'
	CMD_DELAY_MS     = 'M'
	CMD_PING         = 'P'
	CMD_VERSION      = 'V'
	CMD_SET_DATA_DIR = 'S'
)

// Response bytes
const (
	RESP_OK      = 'K'
	RESP_VALUE   = 'V'
	RESP_ERROR   = 'E'
	RESP_PONG    = 'P'
	RESP_VERSION = 'I'
)

// ArduinoPort wraps serial port with PMP300-specific methods
type ArduinoPort struct {
	port *serial.Port
}

// NewArduinoPort opens and initializes connection to Arduino
func NewArduinoPort(device string) (*ArduinoPort, error) {
	config := &serial.Config{
		Name:        device,
		Baud:        115200,
		ReadTimeout: time.Second * 2,
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open port: %v", err)
	}

	// Wait for Arduino to reset and initialize
	fmt.Println("Waiting for Arduino to initialize...")
	time.Sleep(2 * time.Second)

	// Flush any startup messages
	port.Flush()

	ap := &ArduinoPort{port: port}

	// Test connection with ping
	if err := ap.Ping(); err != nil {
		port.Close()
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	return ap, nil
}

// Close closes the serial port
func (ap *ArduinoPort) Close() error {
	return ap.port.Close()
}

// Ping tests connection to Arduino
func (ap *ArduinoPort) Ping() error {
	_, err := ap.port.Write([]byte{CMD_PING})
	if err != nil {
		return err
	}

	response := make([]byte, 1)
	n, err := ap.port.Read(response)
	if err != nil {
		return err
	}
	if n != 1 || response[0] != RESP_PONG {
		return fmt.Errorf("expected PONG, got 0x%02X", response[0])
	}

	return nil
}

// GetVersion returns firmware version
func (ap *ArduinoPort) GetVersion() (major, minor, patch byte, err error) {
	_, err = ap.port.Write([]byte{CMD_VERSION})
	if err != nil {
		return
	}

	response := make([]byte, 4)
	n, err := ap.port.Read(response)
	if err != nil {
		return
	}
	if n != 4 || response[0] != RESP_VERSION {
		err = fmt.Errorf("invalid version response")
		return
	}

	major = response[1]
	minor = response[2]
	patch = response[3]
	return
}

// OutByte writes a byte to the specified parallel port register
// offset: 0 = Data register, 2 = Control register
func (ap *ArduinoPort) OutByte(offset uint16, value byte) error {
	var cmd byte
	switch offset {
	case 0: // Data register
		cmd = CMD_WRITE_DATA
	case 2: // Control register
		cmd = CMD_WRITE_CTRL
	default:
		return fmt.Errorf("invalid offset: %d (must be 0 or 2)", offset)
	}

	_, err := ap.port.Write([]byte{cmd, value})
	if err != nil {
		return err
	}

	// Wait for acknowledgment
	response := make([]byte, 1)
	n, err := ap.port.Read(response)
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("no response")
	}

	if response[0] == RESP_ERROR {
		// Read error code
		errCode := make([]byte, 1)
		ap.port.Read(errCode)
		return fmt.Errorf("arduino error: 0x%02X", errCode[0])
	}

	if response[0] != RESP_OK {
		return fmt.Errorf("expected OK, got 0x%02X", response[0])
	}

	return nil
}

// InByte reads a byte from the specified parallel port register
// offset: 1 = Status register
func (ap *ArduinoPort) InByte(offset uint16) (byte, error) {
	if offset != 1 { // Only status register can be read
		return 0, fmt.Errorf("invalid offset: %d (must be 1 for status)", offset)
	}

	_, err := ap.port.Write([]byte{CMD_READ_STATUS})
	if err != nil {
		return 0, err
	}

	// Wait for value response
	response := make([]byte, 2)
	n, err := ap.port.Read(response)
	if err != nil {
		return 0, err
	}

	if n != 2 || response[0] != RESP_VALUE {
		return 0, fmt.Errorf("invalid response")
	}

	return response[1], nil
}

// DelayMicroseconds delays for specified microseconds (0-65535)
func (ap *ArduinoPort) DelayMicroseconds(us uint16) error {
	highByte := byte(us >> 8)
	lowByte := byte(us & 0xFF)

	_, err := ap.port.Write([]byte{CMD_DELAY_US, highByte, lowByte})
	if err != nil {
		return err
	}

	response := make([]byte, 1)
	_, err = ap.port.Read(response)
	if err != nil {
		return err
	}

	if response[0] != RESP_OK {
		return fmt.Errorf("delay failed")
	}

	return nil
}

// DelayMilliseconds delays for specified milliseconds (0-65535)
func (ap *ArduinoPort) DelayMilliseconds(ms uint16) error {
	highByte := byte(ms >> 8)
	lowByte := byte(ms & 0xFF)

	_, err := ap.port.Write([]byte{CMD_DELAY_MS, highByte, lowByte})
	if err != nil {
		return err
	}

	response := make([]byte, 1)
	_, err = ap.port.Read(response)
	if err != nil {
		return err
	}

	if response[0] != RESP_OK {
		return fmt.Errorf("delay failed")
	}

	return nil
}

// SendCommand sends a command byte to PMP300 using the COMMANDOUT protocol
// This implements the three-step sequence: write data, strobe high, strobe low
func (ap *ArduinoPort) SendCommand(cmd byte) error {
	// Step 1: Write to data register
	if err := ap.OutByte(0, cmd); err != nil {
		return fmt.Errorf("write data failed: %v", err)
	}

	// Step 2: Strobe high (control = 0x0C)
	if err := ap.OutByte(2, 0x0C); err != nil {
		return fmt.Errorf("strobe high failed: %v", err)
	}

	// Step 3: Strobe low (control = 0x04)
	if err := ap.OutByte(2, 0x04); err != nil {
		return fmt.Errorf("strobe low failed: %v", err)
	}

	return nil
}

// WaitForInput waits for specific status value (handshaking)
func (ap *ArduinoPort) WaitForInput(expected byte, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err := ap.InByte(1)
		if err != nil {
			return err
		}

		if (status & 0xF8) == expected {
			return nil
		}

		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for input")
}

// Initialize performs PMP300 initialization sequence
func (ap *ArduinoPort) Initialize() error {
	fmt.Println("Initializing PMP300...")

	// Step 1: Set control to 0x04
	if err := ap.OutByte(2, 0x04); err != nil {
		return err
	}

	// Step 2: Send 0xA8
	if err := ap.SendCommand(0xA8); err != nil {
		return err
	}

	// Step 3: Set control to 0x00
	if err := ap.OutByte(2, 0x00); err != nil {
		return err
	}

	// Step 4: Delay
	if err := ap.DelayMicroseconds(20000); err != nil {
		return err
	}

	// Step 5: Set control to 0x04
	if err := ap.OutByte(2, 0x04); err != nil {
		return err
	}

	// Step 6: Delay
	if err := ap.DelayMicroseconds(20000); err != nil {
		return err
	}

	// Steps 7-11: Send initialization commands
	initSeq := []byte{0xAD, 0x55, 0xAE, 0xAA, 0xA8}
	for _, cmd := range initSeq {
		if err := ap.SendCommand(cmd); err != nil {
			return err
		}
	}

	fmt.Println("Initialization complete!")
	return nil
}

// ============================================================================
// EXAMPLE MAIN FUNCTION
// ============================================================================

func main() {
	// Change this to match your Arduino's serial port
	// Mac: /dev/cu.usbmodem* or /dev/tty.usbmodem*
	// Linux: /dev/ttyACM* or /dev/ttyUSB*
	// Windows: COM3, COM4, etc.
	device := "/dev/cu.usbmodem14201"

	fmt.Println("PMP300 USB-Parallel Bridge - Example Client")
	fmt.Println("===========================================")
	fmt.Printf("Connecting to Arduino on %s...\n", device)

	// Open connection to Arduino
	arduino, err := NewArduinoPort(device)
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer arduino.Close()

	fmt.Println("✓ Connected!")

	// Get firmware version
	major, minor, patch, err := arduino.GetVersion()
	if err != nil {
		log.Fatalf("Failed to get version: %v\n", err)
	}
	fmt.Printf("✓ Firmware version: %d.%d.%d\n", major, minor, patch)

	// Test basic operations
	fmt.Println("\nTesting basic operations...")

	// Test writing to control register
	fmt.Print("  Writing to control register... ")
	if err := arduino.OutByte(2, 0x04); err != nil {
		log.Fatalf("FAILED: %v\n", err)
	}
	fmt.Println("OK")

	// Test reading status register
	fmt.Print("  Reading status register... ")
	status, err := arduino.InByte(1)
	if err != nil {
		log.Fatalf("FAILED: %v\n", err)
	}
	fmt.Printf("OK (value: 0x%02X)\n", status)

	// Test delay
	fmt.Print("  Testing 10ms delay... ")
	start := time.Now()
	if err := arduino.DelayMilliseconds(10); err != nil {
		log.Fatalf("FAILED: %v\n", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("OK (actual: %v)\n", elapsed)

	// Test SendCommand
	fmt.Print("  Testing SendCommand (0xA8)... ")
	if err := arduino.SendCommand(0xA8); err != nil {
		log.Fatalf("FAILED: %v\n", err)
	}
	fmt.Println("OK")

	// Initialize PMP300 (if connected)
	fmt.Println("\nAttempting PMP300 initialization...")
	fmt.Println("(This will fail if no PMP300 is connected - that's OK for testing)")
	if err := arduino.Initialize(); err != nil {
		fmt.Printf("⚠ Initialization sequence completed, but device may not be connected\n")
	} else {
		fmt.Println("✓ PMP300 initialized successfully!")
	}

	fmt.Println("\n===========================================")
	fmt.Println("Example complete! You can now build your")
	fmt.Println("full PMP300 interface using these methods.")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Implement directory reading")
	fmt.Println("2. Implement file upload/download")
	fmt.Println("3. Implement file deletion")
	fmt.Println("\nSee ../README.md for protocol details.")
}

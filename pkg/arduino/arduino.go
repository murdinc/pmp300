package arduino

import (
	"fmt"
	"time"

	"github.com/tarm/serial"
)

// Command bytes - must match Arduino firmware
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

// Error codes
const (
	ERR_UNKNOWN_CMD   = 0x01
	ERR_TIMEOUT       = 0x02
	ERR_INVALID_PARAM = 0x03
)

// Port represents connection to Arduino USB-Parallel bridge
type Port struct {
	serial *serial.Port
	device string
}

// Version contains firmware version information
type Version struct {
	Major byte
	Minor byte
	Patch byte
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Open opens connection to Arduino on specified serial device
func Open(device string) (*Port, error) {
	config := &serial.Config{
		Name:        device,
		Baud:        115200,
		ReadTimeout: time.Second * 3,
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	ap := &Port{
		serial: port,
		device: device,
	}

	// Wait for Arduino to reset and initialize
	time.Sleep(2 * time.Second)

	// Flush any startup messages
	port.Flush()

	// Test connection with ping
	if err := ap.Ping(); err != nil {
		port.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return ap, nil
}

// Close closes the serial port connection
func (p *Port) Close() error {
	if p.serial != nil {
		return p.serial.Close()
	}
	return nil
}

// Ping tests connection to Arduino
func (p *Port) Ping() error {
	_, err := p.serial.Write([]byte{CMD_PING})
	if err != nil {
		return err
	}

	response := make([]byte, 1)
	n, err := p.serial.Read(response)
	if err != nil {
		return err
	}
	if n != 1 || response[0] != RESP_PONG {
		return fmt.Errorf("ping failed: expected PONG, got 0x%02X", response[0])
	}

	return nil
}

// GetVersion returns firmware version
func (p *Port) GetVersion() (*Version, error) {
	_, err := p.serial.Write([]byte{CMD_VERSION})
	if err != nil {
		return nil, err
	}

	response := make([]byte, 4)
	n, err := p.serial.Read(response)
	if err != nil {
		return nil, err
	}
	if n != 4 || response[0] != RESP_VERSION {
		return nil, fmt.Errorf("invalid version response")
	}

	return &Version{
		Major: response[1],
		Minor: response[2],
		Patch: response[3],
	}, nil
}

// OutByte writes a byte to the specified parallel port register
// offset: 0 = Data register, 2 = Control register
func (p *Port) OutByte(offset uint16, value byte) error {
	var cmd byte
	switch offset {
	case 0: // Data register
		cmd = CMD_WRITE_DATA
	case 2: // Control register
		cmd = CMD_WRITE_CTRL
	default:
		return fmt.Errorf("invalid offset: %d (must be 0 or 2)", offset)
	}

	_, err := p.serial.Write([]byte{cmd, value})
	if err != nil {
		return err
	}

	// Wait for acknowledgment
	response := make([]byte, 1)
	n, err := p.serial.Read(response)
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("no response")
	}

	if response[0] == RESP_ERROR {
		// Read error code
		errCode := make([]byte, 1)
		p.serial.Read(errCode)
		return fmt.Errorf("arduino error: 0x%02X", errCode[0])
	}

	if response[0] != RESP_OK {
		return fmt.Errorf("expected OK, got 0x%02X", response[0])
	}

	return nil
}

// InByte reads a byte from the specified parallel port register
// offset: 1 = Status register
func (p *Port) InByte(offset uint16) (byte, error) {
	if offset != 1 { // Only status register can be read
		return 0, fmt.Errorf("invalid offset: %d (must be 1 for status)", offset)
	}

	_, err := p.serial.Write([]byte{CMD_READ_STATUS})
	if err != nil {
		return 0, err
	}

	// Wait for value response
	response := make([]byte, 2)
	n, err := p.serial.Read(response)
	if err != nil {
		return 0, err
	}

	if n != 2 || response[0] != RESP_VALUE {
		return 0, fmt.Errorf("invalid response")
	}

	return response[1], nil
}

// DelayMicroseconds delays for specified microseconds (0-65535)
func (p *Port) DelayMicroseconds(us uint16) error {
	highByte := byte(us >> 8)
	lowByte := byte(us & 0xFF)

	_, err := p.serial.Write([]byte{CMD_DELAY_US, highByte, lowByte})
	if err != nil {
		return err
	}

	response := make([]byte, 1)
	_, err = p.serial.Read(response)
	if err != nil {
		return err
	}

	if response[0] != RESP_OK {
		return fmt.Errorf("delay failed")
	}

	return nil
}

// DelayMilliseconds delays for specified milliseconds (0-65535)
func (p *Port) DelayMilliseconds(ms uint16) error {
	highByte := byte(ms >> 8)
	lowByte := byte(ms & 0xFF)

	_, err := p.serial.Write([]byte{CMD_DELAY_MS, highByte, lowByte})
	if err != nil {
		return err
	}

	response := make([]byte, 1)
	_, err = p.serial.Read(response)
	if err != nil {
		return err
	}

	if response[0] != RESP_OK {
		return fmt.Errorf("delay failed")
	}

	return nil
}

// SetDataDirection sets data pin direction
// dir: 'I' for input, 'O' for output
func (p *Port) SetDataDirection(dir byte) error {
	if dir != 'I' && dir != 'O' {
		return fmt.Errorf("invalid direction: must be 'I' or 'O'")
	}

	_, err := p.serial.Write([]byte{CMD_SET_DATA_DIR, dir})
	if err != nil {
		return err
	}

	response := make([]byte, 1)
	_, err = p.serial.Read(response)
	if err != nil {
		return err
	}

	if response[0] != RESP_OK {
		return fmt.Errorf("set direction failed")
	}

	return nil
}

// Device returns the serial device path
func (p *Port) Device() string {
	return p.device
}

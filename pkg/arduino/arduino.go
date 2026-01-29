package arduino

import (
	"fmt"
	"time"

	"go.bug.st/serial"
)

// Command bytes - must match Arduino firmware
const (
	CMD_PING            = 'P'
	CMD_VERSION         = 'V'
	CMD_WRITE_DATA      = 'W'
	CMD_WRITE_CTRL      = 'C'
	CMD_READ_STATUS     = 'R'
	CMD_DELAY_MS        = 'M'
	CMD_COMMANDOUT      = 'c' // Optimized COMMANDOUT(data, ctrl1, ctrl2)
	CMD_READ_NIBBLE_BLK = 'n'
	CMD_WRITE_PMP_CHUNK = 'w' // Write 528 bytes with PMP300 control toggling
)

// Response bytes
const (
	RESP_OK      = 'K'
	RESP_VALUE   = 'V'
	RESP_ERROR   = 'E'
	RESP_PONG    = 'P'
	RESP_VERSION = 'I'
)

// Port represents connection to Arduino USB-Parallel bridge
type Port struct {
	port   serial.Port
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
	mode := &serial.Mode{
		BaudRate: 500000,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(device, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	if err = port.SetReadTimeout(10 * time.Second); err != nil {
		port.Close()
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	ap := &Port{port: port, device: device}

	// Wait for Arduino to reset
	time.Sleep(2 * time.Second)

	// Flush buffers and test connection
	port.ResetInputBuffer()
	port.ResetOutputBuffer()

	if err := ap.Ping(); err != nil {
		port.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return ap, nil
}

// Close closes the serial port
func (p *Port) Close() error {
	if p.port != nil {
		return p.port.Close()
	}
	return nil
}

// Device returns the serial device path
func (p *Port) Device() string {
	return p.device
}

// Ping tests connection
func (p *Port) Ping() error {
	if _, err := p.port.Write([]byte{CMD_PING}); err != nil {
		return err
	}
	resp := make([]byte, 1)
	if n, err := p.port.Read(resp); err != nil || n != 1 || resp[0] != RESP_PONG {
		return fmt.Errorf("ping failed")
	}
	return nil
}

// GetVersion returns firmware version
func (p *Port) GetVersion() (*Version, error) {
	if _, err := p.port.Write([]byte{CMD_VERSION}); err != nil {
		return nil, err
	}
	resp := make([]byte, 4)
	if _, err := p.readFull(resp); err != nil {
		return nil, err
	}
	if resp[0] != RESP_VERSION {
		return nil, fmt.Errorf("invalid version response")
	}
	return &Version{Major: resp[1], Minor: resp[2], Patch: resp[3]}, nil
}

// OutByte writes a byte to data (offset=0) or control (offset=2) register
func (p *Port) OutByte(offset uint16, value byte) error {
	var cmd byte
	if offset == 0 {
		cmd = CMD_WRITE_DATA
	} else if offset == 2 {
		cmd = CMD_WRITE_CTRL
	} else {
		return fmt.Errorf("invalid offset: %d", offset)
	}

	if _, err := p.port.Write([]byte{cmd, value}); err != nil {
		return err
	}
	return p.waitOK()
}

// InByte reads status register (offset must be 1)
func (p *Port) InByte(offset uint16) (byte, error) {
	if offset != 1 {
		return 0, fmt.Errorf("invalid offset: %d", offset)
	}
	if _, err := p.port.Write([]byte{CMD_READ_STATUS}); err != nil {
		return 0, err
	}
	resp := make([]byte, 2)
	if _, err := p.readFull(resp); err != nil {
		return 0, err
	}
	if resp[0] != RESP_VALUE {
		return 0, fmt.Errorf("invalid status response")
	}
	return resp[1], nil
}

// DelayMilliseconds delays for specified milliseconds
func (p *Port) DelayMilliseconds(ms uint16) error {
	if _, err := p.port.Write([]byte{CMD_DELAY_MS, byte(ms >> 8), byte(ms & 0xFF)}); err != nil {
		return err
	}
	return p.waitOK()
}

// CommandOut executes COMMANDOUT(data, ctrl1, ctrl2) in one USB round-trip
// This is the optimized version - replaces 3 round-trips with 1
func (p *Port) CommandOut(data, ctrl1, ctrl2 byte) error {
	if _, err := p.port.Write([]byte{CMD_COMMANDOUT, data, ctrl1, ctrl2}); err != nil {
		return err
	}
	return p.waitOK()
}

// ReadNibbleBlock reads multiple bytes using PMP300 nibble protocol
func (p *Port) ReadNibbleBlock(count uint16) ([]byte, error) {
	if _, err := p.port.Write([]byte{CMD_READ_NIBBLE_BLK, byte(count >> 8), byte(count & 0xFF)}); err != nil {
		return nil, err
	}

	// Wait for OK
	if err := p.waitOK(); err != nil {
		return nil, err
	}

	// Read data
	data := make([]byte, count)
	if _, err := p.readFull(data); err != nil {
		return nil, err
	}
	return data, nil
}

// WritePMPChunk writes 528 bytes (512 data + 16 end block) with PMP300 control toggling
// This is highly optimized - sends all data in one USB transfer, Arduino handles control toggling
func (p *Port) WritePMPChunk(data []byte) error {
	if len(data) != 528 {
		return fmt.Errorf("chunk must be exactly 528 bytes, got %d", len(data))
	}

	// Send command followed by all 528 bytes
	buf := make([]byte, 1+528)
	buf[0] = CMD_WRITE_PMP_CHUNK
	copy(buf[1:], data)

	if _, err := p.port.Write(buf); err != nil {
		return err
	}
	return p.waitOK()
}

// GetNibbleByte reads one byte using nibble protocol (for single bytes, uses block command)
func (p *Port) GetNibbleByte() (byte, error) {
	data, err := p.ReadNibbleBlock(1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

// Helper: wait for OK response
func (p *Port) waitOK() error {
	resp := make([]byte, 1)
	if _, err := p.readFull(resp); err != nil {
		return err
	}
	if resp[0] == RESP_ERROR {
		errCode := make([]byte, 1)
		p.port.Read(errCode)
		return fmt.Errorf("arduino error: 0x%02X", errCode[0])
	}
	if resp[0] != RESP_OK {
		return fmt.Errorf("expected OK, got 0x%02X", resp[0])
	}
	return nil
}

// Helper: read exactly len(buf) bytes
func (p *Port) readFull(buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := p.port.Read(buf[total:])
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

# PMP300 USB-Parallel Bridge Protocol Specification

This document describes the serial communication protocol between the host computer (Mac/PC) and the Arduino USB-Parallel bridge firmware.

## Connection Parameters

- **Baud Rate**: 115200
- **Data Bits**: 8
- **Parity**: None
- **Stop Bits**: 1
- **Flow Control**: None

## Communication Model

- **Binary Protocol**: All commands and responses are raw bytes (not ASCII/text)
- **Synchronous**: Each command receives a response before the next command
- **Big-Endian**: Multi-byte values sent high byte first

## Command Reference

All commands are single-byte commands, optionally followed by parameter bytes.

### 0x01 - Write Data Register

**Command**: `'W'` (0x57)

**Format**:
```
Send: 'W' <byte_value>
Recv: 'K'
```

**Description**: Writes an 8-bit value to the data register (parallel port data pins D0-D7).
Automatically sets data pins to OUTPUT mode.

**Parameters**:
- `byte_value`: Value to write (0x00-0xFF)

**Response**:
- `'K'` (0x4B): Success

**Example**:
```
Send: 'W' 0xA8
Recv: 'K'
```

---

### 0x02 - Write Control Register

**Command**: `'C'` (0x43)

**Format**:
```
Send: 'C' <byte_value>
Recv: 'K'
```

**Description**: Writes to the control register. Only bits 2 and 3 are used:
- Bit 2: nInitialize (Pin 16 on DB-25)
- Bit 3: nSelect-In (Pin 17 on DB-25)

Other bits are ignored.

**Parameters**:
- `byte_value`: Control value (typically 0x00, 0x04, 0x08, or 0x0C)

**Response**:
- `'K'` (0x4B): Success

**Example**:
```
Send: 'C' 0x04    // Set bit 2 high, bit 3 low
Recv: 'K'
```

---

### 0x03 - Read Status Register

**Command**: `'R'` (0x52)

**Format**:
```
Send: 'R'
Recv: 'V' <byte_value>
```

**Description**: Reads the status register. Returns a byte with status bits in positions 3-7:
- Bit 3: nError (Pin 15)
- Bit 4: Select (Pin 13)
- Bit 5: Paper-Out (Pin 12)
- Bit 6: nAck (Pin 10)
- Bit 7: Busy (Pin 11)

Bits 0-2 are always 0.

**Response**:
- `'V'` (0x56): Value follows
- `byte_value`: Status register value (0x00-0xFF)

**Example**:
```
Send: 'R'
Recv: 'V' 0xF8    // Status value 0xF8
```

---

### 0x04 - Delay Microseconds

**Command**: `'D'` (0x44)

**Format**:
```
Send: 'D' <high_byte> <low_byte>
Recv: 'K'
```

**Description**: Delays for the specified number of microseconds (0-65535).

**Parameters**:
- `high_byte`: High byte of 16-bit microsecond value
- `low_byte`: Low byte of 16-bit microsecond value

**Response**:
- `'K'` (0x4B): Success (sent after delay completes)

**Example**:
```
Send: 'D' 0x00 0x0A    // Delay 10 microseconds
Recv: 'K'

Send: 'D' 0x27 0x10    // Delay 10000 microseconds (10ms)
Recv: 'K'
```

---

### 0x05 - Delay Milliseconds

**Command**: `'M'` (0x4D)

**Format**:
```
Send: 'M' <high_byte> <low_byte>
Recv: 'K'
```

**Description**: Delays for the specified number of milliseconds (0-65535).

**Parameters**:
- `high_byte`: High byte of 16-bit millisecond value
- `low_byte`: Low byte of 16-bit millisecond value

**Response**:
- `'K'` (0x4B): Success (sent after delay completes)

**Example**:
```
Send: 'M' 0x00 0x64    // Delay 100 milliseconds
Recv: 'K'

Send: 'M' 0x13 0x88    // Delay 5000 milliseconds (5 seconds)
Recv: 'K'
```

---

### 0x06 - Ping

**Command**: `'P'` (0x50)

**Format**:
```
Send: 'P'
Recv: 'P'
```

**Description**: Connection test. Arduino responds immediately with 'P'.

**Response**:
- `'P'` (0x50): Pong

**Example**:
```
Send: 'P'
Recv: 'P'
```

---

### 0x07 - Get Version

**Command**: `'V'` (0x56)

**Format**:
```
Send: 'V'
Recv: 'I' <major> <minor> <patch>
```

**Description**: Get firmware version information.

**Response**:
- `'I'` (0x49): Version info follows
- `major`: Major version (0-255)
- `minor`: Minor version (0-255)
- `patch`: Patch version (0-255)

**Example**:
```
Send: 'V'
Recv: 'I' 0x01 0x00 0x00    // Version 1.0.0
```

---

### 0x08 - Set Data Direction

**Command**: `'S'` (0x53)

**Format**:
```
Send: 'S' <direction>
Recv: 'K'
```

**Description**: Set the direction of data pins (INPUT or OUTPUT).

**Parameters**:
- `direction`:
  - `'I'` (0x49): Set as INPUT (for reading data from device)
  - `'O'` (0x4F): Set as OUTPUT (for writing data to device)

**Response**:
- `'K'` (0x4B): Success
- `'E' 0x03`: Error - invalid parameter

**Example**:
```
Send: 'S' 'O'    // Set data pins as outputs
Recv: 'K'

Send: 'S' 'I'    // Set data pins as inputs
Recv: 'K'
```

---

## Error Responses

**Format**:
```
Recv: 'E' <error_code>
```

**Error Codes**:
- `0x01`: Unknown command
- `0x02`: Timeout waiting for parameter
- `0x03`: Invalid parameter value

**Example**:
```
Send: 'Z'        // Invalid command
Recv: 'E' 0x01   // Unknown command error
```

---

## Typical Command Sequences

### Initialize PMP300 Device

Based on the protocol in README.md, the initialization sequence is:

```
1. Write Control: 'C' 0x04
2. Write Data: 'W' 0xA8
   Write Control: 'C' 0x0C
   Write Control: 'C' 0x04
3. Write Control: 'C' 0x00
4. Delay: 'D' 0x4E 0x20  (20000 microseconds)
5. Write Control: 'C' 0x04
6. Delay: 'D' 0x4E 0x20  (20000 microseconds)
7. Send commands: 0xAD, 0x55, 0xAE, 0xAA, 0xA8
```

Actual sequence:
```go
// Step 1: Set control to 0x04
SendCommand('C', 0x04)

// Step 2: Send 0xA8 using SendCommand macro
SendCommand('W', 0xA8)  // Write to data register
SendCommand('C', 0x0C)  // Strobe high
SendCommand('C', 0x04)  // Strobe low

// Step 3: Set control to 0x00
SendCommand('C', 0x00)

// Step 4: Delay
SendCommand('D', 0x4E, 0x20)  // 20000 microseconds

// Step 5: Set control to 0x04
SendCommand('C', 0x04)

// Step 6: Delay
SendCommand('D', 0x4E, 0x20)  // 20000 microseconds

// Steps 7-11: Send initialization commands
for _, cmd := range []byte{0xAD, 0x55, 0xAE, 0xAA, 0xA8} {
    SendCommand('W', cmd)
    SendCommand('C', 0x0C)
    SendCommand('C', 0x04)
}
```

### Send Command Byte to PMP300

The COMMANDOUT macro from README.md translates to:

```go
func SendCommand(cmd byte) {
    // Write to data register
    serial.Write([]byte{'W', cmd})
    waitForResponse('K')

    // Strobe high (control = 0x0C)
    serial.Write([]byte{'C', 0x0C})
    waitForResponse('K')

    // Strobe low (control = 0x04)
    serial.Write([]byte{'C', 0x04})
    waitForResponse('K')
}
```

### Read Status Register

```go
func ReadStatus() byte {
    serial.Write([]byte{'R'})
    response := make([]byte, 2)
    serial.Read(response)

    if response[0] != 'V' {
        panic("Expected 'V' response")
    }

    return response[1]
}
```

### Wait for Input (Handshaking)

```go
func WaitForInput(expected byte, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        serial.Write([]byte{'R'})
        response := make([]byte, 2)
        serial.Read(response)

        status := response[1]
        if (status & 0xF8) == expected {
            return nil
        }

        time.Sleep(10 * time.Millisecond)
    }

    return errors.New("timeout waiting for input")
}
```

---

## Implementation Notes for Mac Software

### Serial Port Configuration

On macOS, Arduino devices typically appear as:
- `/dev/cu.usbmodem*` (for Uno/Mega with ATmega16U2)
- `/dev/cu.usbserial-*` (for older FTDI-based boards)

Use Go's `github.com/tarm/serial` or `go.bug.st/serial` library:

```go
import "github.com/tarm/serial"

config := &serial.Config{
    Name: "/dev/cu.usbmodem14201",
    Baud: 115200,
}

port, err := serial.OpenPort(config)
```

### Wrapper Functions

Create wrapper functions that match the parallel port API from README.md:

```go
type ArduinoPort struct {
    port *serial.Port
}

func (a *ArduinoPort) OutByte(offset uint16, value byte) error {
    var cmd byte
    switch offset {
    case 0: // Data register
        cmd = 'W'
    case 2: // Control register
        cmd = 'C'
    default:
        return fmt.Errorf("invalid offset: %d", offset)
    }

    _, err := a.port.Write([]byte{cmd, value})
    if err != nil {
        return err
    }

    // Wait for 'K' response
    response := make([]byte, 1)
    _, err = a.port.Read(response)
    if err != nil {
        return err
    }

    if response[0] != 'K' {
        return fmt.Errorf("expected 'K', got 0x%02X", response[0])
    }

    return nil
}

func (a *ArduinoPort) InByte(offset uint16) (byte, error) {
    if offset != 1 { // Status register
        return 0, fmt.Errorf("invalid offset: %d", offset)
    }

    _, err := a.port.Write([]byte{'R'})
    if err != nil {
        return 0, err
    }

    // Wait for 'V' + value response
    response := make([]byte, 2)
    _, err = a.port.Read(response)
    if err != nil {
        return 0, err
    }

    if response[0] != 'V' {
        return 0, fmt.Errorf("expected 'V', got 0x%02X", response[0])
    }

    return response[1], nil
}
```

### Error Handling

Always check for error responses:

```go
func sendCommand(port *serial.Port, cmd byte, params ...byte) error {
    msg := append([]byte{cmd}, params...)
    _, err := port.Write(msg)
    if err != nil {
        return err
    }

    response := make([]byte, 1)
    _, err = port.Read(response)
    if err != nil {
        return err
    }

    switch response[0] {
    case 'K', 'V', 'P', 'I':
        return nil
    case 'E':
        // Read error code
        errCode := make([]byte, 1)
        port.Read(errCode)
        return fmt.Errorf("arduino error: 0x%02X", errCode[0])
    default:
        return fmt.Errorf("unexpected response: 0x%02X", response[0])
    }
}
```

### Startup Sequence

When connecting to Arduino:

1. Open serial port at 115200 baud
2. Wait 2 seconds for Arduino to reset and initialize
3. Send ping command to verify connection
4. Optionally get version info

```go
func connect(device string) (*ArduinoPort, error) {
    config := &serial.Config{
        Name: device,
        Baud: 115200,
    }

    port, err := serial.OpenPort(config)
    if err != nil {
        return nil, err
    }

    // Wait for Arduino reset and startup
    time.Sleep(2 * time.Second)

    // Drain any startup messages
    port.Flush()

    // Test connection with ping
    port.Write([]byte{'P'})
    response := make([]byte, 1)
    port.Read(response)

    if response[0] != 'P' {
        port.Close()
        return nil, fmt.Errorf("ping failed")
    }

    return &ArduinoPort{port: port}, nil
}
```

---

## Performance Considerations

### USB Latency

- Each command has ~1-2ms USB round-trip latency
- For operations requiring many commands (like reading 32KB), this adds up
- Consider batching operations where possible

### Timing Accuracy

- Microsecond delays are handled by Arduino, providing accurate timing
- Don't rely on delays in Mac software - use Arduino delay commands
- Arduino's `delayMicroseconds()` is accurate to ~4μs on 16MHz boards

### Buffer Management

- Arduino serial buffer is 64 bytes by default
- Don't send multiple commands without waiting for responses
- Implement proper handshaking for reliable communication

---

## Testing the Bridge

### Test 1: Basic Connectivity

```go
// Ping test
port.Write([]byte{'P'})
response := make([]byte, 1)
port.Read(response)
// Should receive 'P'
```

### Test 2: Write and Read

```go
// Write to control register
port.Write([]byte{'C', 0x04})
response := make([]byte, 1)
port.Read(response)
// Should receive 'K'

// Read status register
port.Write([]byte{'R'})
response = make([]byte, 2)
port.Read(response)
// Should receive 'V' followed by status byte
```

### Test 3: Timing

```go
// Test microsecond delay
start := time.Now()
port.Write([]byte{'D', 0x27, 0x10}) // 10000 microseconds
response := make([]byte, 1)
port.Read(response)
elapsed := time.Since(start)
// elapsed should be ~10-12ms (10ms delay + USB latency)
```

---

## Protocol Version History

### Version 1.0.0
- Initial implementation
- Basic parallel port register access
- Microsecond and millisecond delays
- Ping and version commands
- Data direction control

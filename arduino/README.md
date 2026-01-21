# Arduino Firmware and Examples

This directory contains Arduino firmware and example code for building a USB-to-parallel-port bridge for the Diamond Rio PMP300.

## Directory Contents

### `pmp300_usb_parallel_bridge/`
**Main production firmware** - Full USB-to-parallel bridge implementation

- `pmp300_usb_parallel_bridge.ino` - Complete Arduino sketch
- `PROTOCOL.md` - Binary protocol specification
- `README.md` - Firmware documentation and usage guide

**Use this for**: Production use with your PMP300

### `hardware_test/`
**Hardware testing sketch** - Verify your wiring before using the full bridge

- `hardware_test.ino` - Interactive test sketch

**Use this for**: Testing your Arduino-to-DB25 connections with a multimeter or logic analyzer

### `example_client.go`
**Example Go client** - Demonstrates how to communicate with the bridge from your Mac

- Complete example showing all protocol commands
- Reference implementation for building your PMP300 software
- Includes initialization sequence

**Use this for**: Learning how to write your Mac software

## Quick Start Guide

### 1. Test Your Hardware

First, verify your wiring is correct:

```bash
# Upload the hardware test sketch
1. Open arduino/hardware_test/hardware_test.ino in Arduino IDE
2. Select your board (Tools → Board → Arduino Mega/Uno)
3. Upload to Arduino
4. Open Serial Monitor at 115200 baud
5. Follow on-screen menu to test pins
6. Use multimeter to verify voltages
```

### 2. Upload Production Firmware

Once wiring is verified:

```bash
# Upload the USB-parallel bridge firmware
1. Open arduino/pmp300_usb_parallel_bridge/pmp300_usb_parallel_bridge.ino
2. Select your board (Tools → Board → Arduino Mega/Uno)
3. Upload to Arduino
4. Verify startup message in Serial Monitor
```

### 3. Test the Bridge

Test communication from your computer:

```bash
# Run the example Go client
cd arduino
go get github.com/tarm/serial
go run example_client.go

# Edit example_client.go first to set your serial port device
# Mac: /dev/cu.usbmodem*
# Linux: /dev/ttyACM*
# Windows: COM3, COM4, etc.
```

### 4. Build Your Software

Use the example client as a starting point:

1. Copy the `ArduinoPort` struct and methods from `example_client.go`
2. Implement the PMP300 protocol from the main README.md
3. Reference `pmp300_usb_parallel_bridge/PROTOCOL.md` for command details

## Which Board Should I Use?

### Arduino Mega 2560 (Recommended)
- ✅ Plenty of pins (54 digital I/O)
- ✅ More RAM (8KB)
- ✅ Room for debugging/expansion
- ✅ Easier wiring layout
- 💰 Cost: $25-30

### Arduino Uno (Works Fine)
- ✅ Sufficient pins (20 I/O, 18 available)
- ⚠️ Less RAM (2KB)
- ⚠️ Tighter pin layout
- 💰 Cost: $15-20

Both work! Mega is more comfortable, Uno saves $10.

## Wiring Reference

See main project README.md for complete wiring diagrams.

**Key points:**
- Use **DB-25 female connector** (accepts your PMP300 cable)
- Connect **Arduino GND to DB-25 pins 18-25**
- All connections are direct (no level shifters needed - both 5V)
- Verify with multimeter before connecting PMP300

## Protocol Overview

Communication uses binary serial protocol at 115200 baud:

```
Write Data:    'W' <byte>         → 'K'
Write Control: 'C' <byte>         → 'K'
Read Status:   'R'                → 'V' <byte>
Delay μs:      'D' <high> <low>   → 'K'
Ping:          'P'                → 'P'
Version:       'V'                → 'I' <maj> <min> <patch>
```

See `pmp300_usb_parallel_bridge/PROTOCOL.md` for complete details.

## Example: Initialize PMP300

```go
// Connect to Arduino
arduino, _ := NewArduinoPort("/dev/cu.usbmodem14201")

// PMP300 initialization sequence
arduino.OutByte(2, 0x04)              // Control = 0x04
arduino.SendCommand(0xA8)             // Send 0xA8
arduino.OutByte(2, 0x00)              // Control = 0x00
arduino.DelayMicroseconds(20000)      // Delay 20ms
arduino.OutByte(2, 0x04)              // Control = 0x04
arduino.DelayMicroseconds(20000)      // Delay 20ms

// Send init sequence
for _, cmd := range []byte{0xAD, 0x55, 0xAE, 0xAA, 0xA8} {
    arduino.SendCommand(cmd)
}
```

## Troubleshooting

### Arduino Not Responding
- Check USB cable (some are power-only)
- Verify correct port selected in Arduino IDE
- Press reset button and try again
- Check Serial Monitor baud rate (115200)

### No Data from PMP300
- Verify all pins connected correctly
- Check ground connection (critical!)
- Test with multimeter (should see 0V or 5V)
- Use hardware_test sketch to verify pins

### Mac Software Can't Connect
- Check serial port device name (`ls /dev/cu.*`)
- Wait 2 seconds after opening port (Arduino resets)
- Verify firmware uploaded successfully
- Try ping command first

### Wrong Data/Timing
- Check pin mapping matches your board type
- Verify firmware auto-detected correct board
- Check for swapped pins (use walking 1's test)
- Monitor with logic analyzer if available

## Development Workflow

1. **Hardware assembly**: Wire Arduino to DB-25 connector
2. **Hardware testing**: Upload `hardware_test` sketch, verify with multimeter
3. **Firmware upload**: Upload `pmp300_usb_parallel_bridge` firmware
4. **Communication test**: Run `example_client.go` to verify bridge
5. **Software development**: Build your PMP300 interface using example as reference

## Performance Notes

- USB latency: ~1-2ms per command
- Reading 32KB directory: ~65 seconds worst case
- Timing accuracy: ~4μs on 16MHz Arduino (adequate for PMP300)
- Buffer size: 64 bytes (commands processed synchronously)

## Safety

⚠️ **5V only**: Do not use 3.3V Arduino boards (Due, Zero) without level shifters
⚠️ **Current limit**: Max 40mA per pin - parallel port devices should not exceed this
⚠️ **ESD protection**: Handle vintage PMP300 on anti-static surface
⚠️ **Verify wiring**: Double-check before powering on

## Resources

- Main project README: `../README.md` (PMP300 protocol documentation)
- Protocol spec: `pmp300_usb_parallel_bridge/PROTOCOL.md`
- Firmware docs: `pmp300_usb_parallel_bridge/README.md`
- Example code: `example_client.go`

## License

MIT License - See main project LICENSE file

## Credits

Based on PMP300 protocol reverse engineering by:
- Snowblind Alliance RIO Utility
- wfx_rio project
- Community reverse engineering efforts

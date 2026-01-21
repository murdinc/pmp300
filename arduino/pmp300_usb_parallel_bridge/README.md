# PMP300 USB-Parallel Bridge - Arduino Firmware

This directory contains the Arduino firmware that implements a USB-to-5V-parallel-port bridge for interfacing with the Diamond Rio PMP300 MP3 player.

## Quick Start

### 1. Hardware Setup

**Supported Boards:**
- Arduino Mega 2560 (recommended - 54 I/O pins)
- Arduino Uno (works - 20 I/O pins, tight fit)

**Required Components:**
- Arduino board
- DB-25 female connector
- Jumper wires or solid core wire
- USB cable (for Arduino-to-computer connection)

**Wiring:**
See the main project README.md for complete wiring diagrams. The firmware automatically detects whether you're using a Mega or Uno and configures pins accordingly.

### 2. Upload Firmware

1. Open `pmp300_usb_parallel_bridge.ino` in Arduino IDE
2. Select your board: Tools → Board → Arduino Mega 2560 (or Uno)
3. Select the correct port: Tools → Port → /dev/cu.usbmodem* (on Mac)
4. Click Upload button
5. Wait for upload to complete

### 3. Verify Installation

Open Serial Monitor (Tools → Serial Monitor):
- Set baud rate to **115200**
- You should see:
  ```
  PMP300 USB-Parallel Bridge v1.0.0
  Board: Arduino Mega 2560
  Ready.
  ```

### 4. Test Connection

Send a ping command to test:
- In Serial Monitor, change line ending to "No line ending"
- Type: `P` and click Send
- You should receive: `P` back

## Files in This Directory

- `pmp300_usb_parallel_bridge.ino` - Main Arduino sketch
- `PROTOCOL.md` - Complete communication protocol specification
- `README.md` - This file

## Firmware Features

✅ **Automatic board detection** - Works with Mega and Uno
✅ **Binary serial protocol** - Efficient communication at 115200 baud
✅ **Full parallel port emulation** - Data, Control, and Status registers
✅ **Bidirectional data pins** - Can read and write
✅ **Microsecond timing** - Accurate delays for protocol compliance
✅ **Error handling** - Timeouts and invalid command detection
✅ **Debugging support** - Version and ping commands

## Pin Assignments

### Arduino Mega 2560

```
Data Pins (Bidirectional):
  D22-D29 → DB-25 pins 2-9 (Data0-Data7)

Control Pins (Output):
  D30 → DB-25 pin 16 (nInitialize)
  D31 → DB-25 pin 17 (nSelect-In)

Status Pins (Input):
  D32 → DB-25 pin 15 (nError)
  D33 → DB-25 pin 13 (Select)
  D34 → DB-25 pin 12 (Paper-Out)
  D35 → DB-25 pin 10 (nAck)
  D36 → DB-25 pin 11 (Busy)

Ground:
  GND → DB-25 pins 18-25
```

### Arduino Uno

```
Data Pins (Bidirectional):
  D2-D9 → DB-25 pins 2-9 (Data0-Data7)

Control Pins (Output):
  D10 → DB-25 pin 16 (nInitialize)
  D11 → DB-25 pin 17 (nSelect-In)

Status Pins (Input):
  D12 → DB-25 pin 15 (nError)
  D13 → DB-25 pin 13 (Select)
  A0  → DB-25 pin 12 (Paper-Out)
  A1  → DB-25 pin 10 (nAck)
  A2  → DB-25 pin 11 (Busy)

Ground:
  GND → DB-25 pins 18-25
```

## Command Reference (Quick)

See `PROTOCOL.md` for complete details.

| Command | Byte | Parameters | Response | Description |
|---------|------|------------|----------|-------------|
| Write Data | `'W'` | 1 byte | `'K'` | Write to data register |
| Write Control | `'C'` | 1 byte | `'K'` | Write to control register |
| Read Status | `'R'` | none | `'V'` + byte | Read status register |
| Delay μs | `'D'` | 2 bytes | `'K'` | Delay microseconds |
| Delay ms | `'M'` | 2 bytes | `'K'` | Delay milliseconds |
| Ping | `'P'` | none | `'P'` | Connection test |
| Version | `'V'` | none | `'I'` + 3 bytes | Get firmware version |
| Set Direction | `'S'` | `'I'` or `'O'` | `'K'` | Set data pin direction |

## Troubleshooting

### Upload fails
- **Check board selection**: Make sure you selected the correct board type
- **Check port**: Verify correct USB port is selected
- **Try different cable**: Some USB cables are power-only
- **Reset board**: Press reset button before uploading

### No serial output
- **Check baud rate**: Must be 115200
- **Wait after reset**: Arduino resets when you open Serial Monitor - wait 2 seconds
- **Check USB connection**: Try unplugging and replugging

### Device not responding
- **Check wiring**: Verify all pins are connected correctly
- **Check ground**: Must have common ground between Arduino and DB-25
- **Use multimeter**: Test continuity of wires
- **Check power**: Arduino should have power LED lit

### Wrong data/timing
- **Check pin mapping**: Compare wiring to pin assignments above
- **Verify board type**: Make sure firmware matches your physical board
- **Check inverted signals**: Note that some pins are inverted (nError, nAck, etc.)

## Development

### Modifying the Firmware

If you need to modify the firmware:

1. **Pin assignments**: Edit the `#define` statements at the top
2. **Protocol changes**: Update command handlers and PROTOCOL.md
3. **Timing adjustments**: Modify delay values in command handlers
4. **Add features**: Add new command handlers in the switch statement

### Testing Without PMP300

You can test the firmware without a PMP300 connected:

1. Upload firmware to Arduino
2. Open Serial Monitor at 115200 baud
3. Send test commands (see PROTOCOL.md)
4. Use multimeter to verify pin outputs
5. Add LEDs to output pins for visual debugging

### Adding Debug Output

To add debug messages (be careful not to interfere with protocol):

```cpp
// Only send debug messages before main loop starts
void setup() {
  Serial.begin(115200);

  // Debug output here is safe
  Serial.println("Debug: Initializing...");

  // ... rest of setup ...

  Serial.println("Ready.");
  // After this, only use protocol commands
}
```

### Version Updates

When updating firmware version:

1. Update version constants at top of .ino file:
   ```cpp
   #define FW_VERSION_MAJOR  1
   #define FW_VERSION_MINOR  0
   #define FW_VERSION_PATCH  1
   ```

2. Update PROTOCOL.md with changelog

3. Test with version command: Send `'V'`, receive `'I'` + 3 version bytes

## Performance Notes

### USB Latency
- Each command has ~1-2ms round-trip time over USB
- Reading 32KB directory (32768 bytes) takes ~65 seconds at worst case
- This is acceptable for the PMP300's intended use

### Timing Accuracy
- `delayMicroseconds()` is accurate to ~4μs on 16MHz Arduino
- PMP300 protocol requires 15-100μs delays (easily achievable)
- Critical timing is handled by Arduino, not host computer

### Buffer Limitations
- Arduino serial buffer is 64 bytes
- Commands are processed one at a time (synchronous)
- No command batching - keeps protocol simple and reliable

## Safety Notes

⚠️ **Voltage**: Arduino Mega/Uno outputs are 5V - matches PMP300 parallel port requirements perfectly. Do not use 3.3V boards (Due, Zero, etc.) without level shifters.

⚠️ **Current**: Each Arduino pin can source/sink 40mA max. Parallel port devices should not exceed this.

⚠️ **Shorts**: Double-check wiring before powering on. A short circuit can damage the Arduino.

⚠️ **ESD**: Handle electronics on an anti-static surface. The PMP300 is vintage and may be sensitive.

## License

MIT License - See main project LICENSE file

## Support

For issues, questions, or improvements:
1. Check PROTOCOL.md for communication protocol details
2. See main project README.md for PMP300 protocol documentation
3. File issues on the project repository

## Credits

Based on reverse-engineered PMP300 protocol documentation from:
- Snowblind Alliance RIO Utility v1.07
- wfx_rio project
- Community reverse engineering efforts

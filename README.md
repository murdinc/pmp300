# Diamond Rio PMP300 Management Tool

Complete solution for managing the Diamond Rio PMP300 MP3 player on modern computers.

The PMP300 was one of the first portable MP3 players (1998). This project provides a CLI tool and Arduino-based USB interface to use it with modern computers that lack parallel ports.

## Quick Start

1. **Hardware Setup**: Connect Arduino to PMP300 (see [Wiring Guide](arduino/WIRING.md))
2. **Flash Firmware**: `pmp300 flash` (requires arduino-cli)
3. **Test**: `pmp300 test`
4. **Use**: `pmp300 list`, `pmp300 upload song.mp3`, etc.

See [QUICKSTART.md](QUICKSTART.md) for detailed 5-minute setup guide.

## What You Can Do

### File Operations
- Upload MP3 files to device
- Download files from device
- Delete individual files or all files
- List files with metadata (size, timestamp)

### Device Management
- View device information (capacity, free space)
- Change playback order of songs
- Format/initialize device
- Bad block detection
- Switch between internal flash and external SmartMedia card

### Not Supported
- Firmware updates (no public protocol exists)

## CLI Commands

```bash
pmp300 test                          # Test Arduino and PMP300 connection
pmp300 flash                         # Flash Arduino with firmware
pmp300 list                          # List all files
pmp300 info                          # Show device information
pmp300 upload song.mp3               # Upload file(s)
pmp300 download song.mp3             # Download file
pmp300 delete song.mp3               # Delete file(s)
pmp300 move 3 1                      # Rearrange playback order
pmp300 format                        # Format device
pmp300 storage list                  # Show available storage (internal/external)
pmp300 storage switch external       # Switch to SmartMedia card
pmp300 version                       # Show version
```

See [CLI_README.md](CLI_README.md) for complete command reference.

## Hardware

### Device Specifications

- **Model**: Diamond Rio PMP300
- **Storage**: 32MB internal flash (64MB on SE models)
- **Interface**: Parallel Port (LPT)
- **Block Size**: 32KB (32,768 bytes)
- **Total Blocks**: 1024 (32MB) or 2048 (64MB SE)

### Modern Interface Options

**Recommended: Arduino USB Bridge**
- Arduino Mega 2560 or Uno
- DB-25 female connector
- 15 signal wires + ground
- Cost: ~$30-50 total

See [arduino/](arduino/) directory for:
- Complete firmware (auto-detects Mega/Uno)
- Wiring diagrams
- Protocol documentation
- Hardware test sketch

### Parallel Port Pinout

DB-25 female connector (PC side):

**Data Pins (8 signals, bidirectional):**
- Pins 2-9: Data 0-7

**Control Pins (2 signals used, output):**
- Pin 16: nInitialize
- Pin 17: nSelect-In

**Status Pins (5 signals, input):**
- Pin 15: nError
- Pin 13: Select
- Pin 12: Paper-Out
- Pin 10: nAck
- Pin 11: Busy

**Ground:**
- Pins 18-25: Ground (connect all)

See [arduino/WIRING.md](arduino/WIRING.md) for complete wiring diagrams.

## Protocol Overview

### Communication
- 5V TTL logic levels
- Bidirectional parallel port
- Command/response protocol
- Microsecond timing control

### Data Organization
- **Block 0**: Directory and FAT
- **Blocks 1+**: File data (32KB each)
- **Max Files**: 60
- **Max Filename**: 127 characters

### Timing (typical)
- Reading directory: ~30 seconds
- Uploading 1MB file: ~5 minutes
- Uploading full 32MB: ~2.5 hours
- Bad block scan: ~3-4 hours (32MB)

### Protocol Version
Version 107 (0x6B)

## Installation

### Requirements
- Go 1.21 or later
- Arduino Mega 2560 or Uno
- arduino-cli (for flashing firmware)

### Build

```bash
git clone https://github.com/murdinc/pmp300
cd pmp300
make build
# Binary at: build/pmp300
```

Or install to system:
```bash
make install
# Installs to $GOPATH/bin/pmp300
```

### Flash Arduino Firmware

```bash
# Install arduino-cli first
brew install arduino-cli  # macOS
# or see: https://arduino.github.io/arduino-cli/

# Flash firmware (auto-detects board and port)
pmp300 flash
```

### Set Device

```bash
# Set once per session
export PMP300_DEVICE=/dev/cu.usbmodem14201

# Or add to ~/.zshrc or ~/.bashrc
echo 'export PMP300_DEVICE=/dev/cu.usbmodem14201' >> ~/.zshrc
```

## Example Usage

```bash
# Test everything works
pmp300 test

# Check device info
pmp300 info

# Upload music to internal flash (default)
pmp300 upload ~/Music/*.mp3

# List files on internal flash
pmp300 list --verbose

# Check if SmartMedia card is inserted
pmp300 storage list

# Switch to external SmartMedia card
pmp300 storage switch external

# Upload to SmartMedia card
pmp300 upload ~/Music/album/*.mp3

# List files on SmartMedia
pmp300 list

# Switch back to internal
pmp300 storage switch internal

# Rearrange songs
pmp300 move 5 1    # Move track 5 to first position

# Download a file
pmp300 download "01 Song.mp3" --output ~/backup/

# Delete files
pmp300 delete old-song.mp3
pmp300 delete --all  # Delete everything

# Format device (formats current storage)
pmp300 format
```

## Technical Details

### Port Registers

| Register | Offset | Access | Description |
|----------|--------|--------|-------------|
| Data     | +0     | R/W    | 8-bit data |
| Status   | +1     | R      | Device status (bits 3-7 used) |
| Control  | +2     | R/W    | Control signals (bits 2-3 used) |

### Directory Structure

**Header (32 bytes):**
- Entry count, free blocks, bad blocks, total blocks
- Timestamp, checksum, version

**Entries (60 × 144 bytes):**
- Block position, block count, total size
- Timestamp (YYMMDDHHMMSS)
- Filename (128 bytes, null-terminated)

**FAT (8192 bytes):**
- One byte per block
- 0x00 = Used, 0x0F = Bad, 0xFF = Free

**Total: 32KB (one block)**

### Block Operations

**Reading:**
1. Send read command (0xA0)
2. Send 3-byte block address
3. Wait for device ready
4. Read 32KB data

**Writing:**
1. Send write command (0xAB)
2. Send 3-byte block address
3. Write data in 512-byte chunks (64 chunks)
4. Send checksum per chunk
5. Wait for acknowledge per chunk

## Project Structure

```
pmp300/
├── main.go              # CLI entry point
├── cmd/                 # CLI commands
├── pkg/
│   ├── arduino/         # Arduino bridge protocol
│   └── pmp300/          # PMP300 device protocol
├── arduino/             # Firmware and wiring docs
├── CLI_README.md        # Complete CLI reference
├── QUICKSTART.md        # 5-minute setup guide
└── Makefile             # Build automation
```

## Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Get running in 5 minutes
- **[CLI_README.md](CLI_README.md)** - Complete command reference
- **[arduino/README.md](arduino/README.md)** - Firmware documentation
- **[arduino/PROTOCOL.md](arduino/PROTOCOL.md)** - Arduino bridge protocol
- **[arduino/WIRING.md](arduino/WIRING.md)** - Wiring diagrams

## Troubleshooting

### Device not found
```bash
# Find Arduino device
./scripts/find-device.sh
export PMP300_DEVICE=/dev/cu.usbmodem14201
```

### Test fails
```bash
# Flash firmware first
pmp300 flash

# Check wiring
# Upload arduino/hardware_test/hardware_test.ino
# Use multimeter to verify connections
```

### Upload/download slow
**This is normal!** Parallel port protocol is inherently slow:
- ~3.5 KB/second typical
- Limited by protocol, not implementation
- Original software had same speeds

### Connection issues
- Check all 15 signal wires are connected
- Verify ground connection (critical!)
- Ensure PMP300 is powered on
- Try different USB cable for Arduino
- Press Arduino reset button

## Performance Notes

| Operation | Time |
|-----------|------|
| Read directory | ~30 seconds |
| Upload 1MB | ~5 minutes |
| Upload 32MB | ~2.5 hours |
| Download 1MB | ~5 minutes |
| Bad block scan (32MB) | ~3-4 hours |

## Development

### Building from Source
```bash
git clone https://github.com/murdinc/pmp300
cd pmp300
make deps      # Download dependencies
make build     # Build CLI tool
make test      # Run tests
```

### Running Development Version
```bash
go run main.go test
go run main.go list
```

### Code Organization
- `cmd/` - Cobra CLI commands
- `pkg/arduino/` - Arduino bridge communication
- `pkg/pmp300/` - PMP300 protocol implementation
- `arduino/` - Firmware and documentation

## References

- [Snowblind Alliance RIO Utility v1.07](http://slackware.cs.utah.edu/pub/slackware/slackware-8.0/contrib/rio.txt)
- [wfx_rio GitHub Repository](https://github.com/creaktive/wfx_rio)
- [OSDev Wiki - Parallel Port](http://wiki.osdev.org/Parallel_port)
- [IEEE 1284 Standard](https://www.ardent-tool.com/comms/an062_Updating_the_parallel_port.pdf)

## License

This documentation and software are provided for educational and preservation purposes. The PMP300 protocol is based on reverse-engineered information from open-source implementations.

---

**Note**: Direct hardware access requires elevated privileges. Use caution when working with hardware interfaces.

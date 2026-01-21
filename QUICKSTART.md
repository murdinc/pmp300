# PMP300 Quick Start Guide

Get up and running with your PMP300 in 5 minutes!

## Prerequisites

- Arduino Mega 2560 or Uno
- DB-25 female connector wired to Arduino (see `arduino/WIRING.md`)
- PMP300 device with original parallel cable
- Go 1.21 or later installed

## Step 1: Hardware Setup

1. **Wire Arduino to DB-25** following `arduino/WIRING.md`
2. **Upload firmware** to Arduino:
   ```bash
   # Open arduino/pmp300_usb_parallel_bridge/pmp300_usb_parallel_bridge.ino
   # in Arduino IDE and upload
   ```
3. **Connect** Arduino to computer via USB
4. **Connect** PMP300 cable to DB-25 connector
5. **Power on** PMP300

## Step 2: Build CLI Tool

```bash
# Navigate to project directory
cd /path/to/pmp300

# Download dependencies
go mod download

# Build
make build
# or: go build -o build/pmp300

# The binary is now at: build/pmp300
```

## Step 3: Find Your Device

```bash
# Run device finder script
chmod +x scripts/find-device.sh
./scripts/find-device.sh

# On macOS, typically shows:
# /dev/cu.usbmodem14201

# Set environment variable
export PMP300_DEVICE=/dev/cu.usbmodem14201
```

## Step 4: Test Connection

```bash
./build/pmp300 test
```

You should see:
```
Connecting to Arduino on /dev/cu.usbmodem14201...
✓ Arduino connected
✓ Firmware version: 1.0.0
Testing ping... OK
Testing control register write... OK
Testing status register read... OK (value: 0xF8)

Initializing PMP300...
✓ PMP300 initialized successfully

=== All tests passed! ===
Your PMP300 is ready to use.
```

## Step 5: Basic Operations

### List Files
```bash
./build/pmp300 list
```

### Get Device Info
```bash
./build/pmp300 info
```

### Upload a File
```bash
./build/pmp300 upload ~/Music/song.mp3
```

### Download a File
```bash
./build/pmp300 download song.mp3
```

### Delete a File
```bash
./build/pmp300 delete song.mp3
```

## Common Issues

### "device not specified"
**Solution**: Set environment variable
```bash
export PMP300_DEVICE=/dev/cu.usbmodem14201
```

Or add to `~/.bashrc` or `~/.zshrc`:
```bash
echo 'export PMP300_DEVICE=/dev/cu.usbmodem14201' >> ~/.zshrc
source ~/.zshrc
```

### "failed to open Arduino"
**Solutions**:
- Check USB cable is connected
- Try a different USB cable (some are power-only)
- Check device path: `ls /dev/cu.usbmodem*`
- Unplug and replug Arduino

### "ping failed"
**Solutions**:
- Ensure firmware is uploaded to Arduino
- Press reset button on Arduino
- Wait 2-3 seconds and try again
- Check serial monitor shows "PMP300 USB-Parallel Bridge v1.0.0"

### "initialization failed"
**Solutions**:
- Check all 15 signal wires are connected (see wiring diagram)
- Verify ground connection (critical!)
- Use multimeter to test continuity
- Upload `hardware_test.ino` to verify wiring
- Ensure PMP300 has power and is turned on

### Upload is very slow
**This is normal!** The parallel port protocol is slow:
- ~3.5 KB/second typical
- 1 MB file = ~5 minutes
- This matches the original parallel port utility speeds

## Next Steps

1. **Read CLI_README.md** for complete command reference
2. **Check arduino/PROTOCOL.md** for technical details
3. **See main README.md** for PMP300 protocol documentation

## Daily Usage

Once set up, typical workflow:

```bash
# Set device once per session (or add to shell profile)
export PMP300_DEVICE=/dev/cu.usbmodem14201

# Upload new music
pmp300 upload ~/Music/new-album/*.mp3

# Check what's on device
pmp300 list

# Rearrange playback order
pmp300 move 5 1

# Download a file
pmp300 download "favorite.mp3"

# Delete old files to make space
pmp300 delete "old-song.mp3"
```

## Tips

- Use tab completion with `pmp300` command
- Run `pmp300 --help` for command list
- Run `pmp300 <command> --help` for command-specific help
- Add `alias pmp='pmp300'` to your shell for shorter command

## Getting Help

If you encounter issues:

1. Run diagnostics:
   ```bash
   pmp300 test
   pmp300 info
   ```

2. Check hardware:
   - Upload `arduino/hardware_test/hardware_test.ino`
   - Use multimeter to verify connections
   - Look for loose wires

3. Check firmware:
   - Open Arduino Serial Monitor at 115200 baud
   - Should see startup message
   - Send 'P' and should receive 'P' back

4. Check documentation:
   - `CLI_README.md` - Full command reference
   - `arduino/PROTOCOL.md` - Communication protocol
   - `arduino/WIRING.md` - Wiring diagrams
   - `README.md` - PMP300 protocol details

Enjoy your vintage MP3 player on modern hardware! 🎵

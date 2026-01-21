## PMP300 CLI Tool

Command-line interface for managing files on the Diamond Rio PMP300 MP3 player via Arduino USB-to-parallel bridge.

## Installation

```bash
# Clone or navigate to the project
cd /path/to/pmp300

# Build the CLI tool
go build -o pmp300

# Optional: Install to system
go install
```

## Quick Start

```bash
# Set your Arduino device (or use --device flag with every command)
export PMP300_DEVICE=/dev/cu.usbmodem14201

# Test connection
./pmp300 test

# List files
./pmp300 list

# Upload a file
./pmp300 upload song.mp3

# Download a file
./pmp300 download song.mp3

# Delete a file
./pmp300 delete song.mp3

# Get device info
./pmp300 info
```

## Commands

### `pmp300 test`
Test connection to Arduino and PMP300.

```bash
pmp300 test --device /dev/cu.usbmodem14201
```

### `pmp300 list` (aliases: `ls`)
List all files on the device.

```bash
pmp300 list                    # Simple listing
pmp300 list --verbose          # Detailed with block info
```

### `pmp300 info`
Display device information (capacity, free space, file count, etc.).

```bash
pmp300 info
```

### `pmp300 upload` (aliases: `put`, `push`)
Upload files to the PMP300.

```bash
pmp300 upload song.mp3                   # Upload single file
pmp300 upload *.mp3                      # Upload multiple files
pmp300 upload ~/Music/album/*.mp3        # Upload from path
```

### `pmp300 download` (aliases: `get`, `pull`)
Download files from the PMP300.

```bash
pmp300 download song.mp3                      # Download to current dir
pmp300 download song.mp3 --output ~/Music/    # Download to specific path
```

### `pmp300 delete` (aliases: `rm`, `remove`)
Delete files from the PMP300.

```bash
pmp300 delete song.mp3                   # Delete single file
pmp300 delete song1.mp3 song2.mp3        # Delete multiple files
pmp300 delete --all                      # Delete ALL files
pmp300 delete --all --force              # Delete all without confirmation
```

### `pmp300 move`
Change playback order by moving a file to a new position.

```bash
pmp300 move 3 1        # Move file from position 3 to position 1
pmp300 move 1 5        # Move file from position 1 to position 5
```

Positions are 1-based (first file is 1, not 0).

### `pmp300 format`
Format/initialize the device (erases all files).

```bash
pmp300 format                           # Quick format
pmp300 format --check-bad-blocks        # Format with bad block detection (very slow)
pmp300 format --force                   # Skip confirmation
```

**WARNING**: Format erases all files!

## Global Flags

### `--device` / `-d`
Specify serial device path.

```bash
pmp300 list --device /dev/cu.usbmodem14201
```

Alternatively, set the `PMP300_DEVICE` environment variable:

```bash
export PMP300_DEVICE=/dev/cu.usbmodem14201
pmp300 list  # Uses environment variable
```

### Finding Your Device

**macOS:**
```bash
ls /dev/cu.usbmodem*
# Typically: /dev/cu.usbmodem14201 or similar
```

**Linux:**
```bash
ls /dev/ttyACM*
# Typically: /dev/ttyACM0 or /dev/ttyUSB0
```

**Windows:**
```
# Check Device Manager for COM port
# Typically: COM3, COM4, etc.
```

## Examples

### Complete Workflow

```bash
# 1. Set device
export PMP300_DEVICE=/dev/cu.usbmodem14201

# 2. Test connection
pmp300 test

# 3. Check device info
pmp300 info

# 4. Upload an album
pmp300 upload ~/Music/MyAlbum/*.mp3

# 5. List files
pmp300 list

# 6. Rearrange playback order
pmp300 move 5 1    # Move track 5 to first position

# 7. Download a file
pmp300 download "01 Song.mp3"

# 8. Delete unwanted file
pmp300 delete "old_song.mp3"
```

### Batch Upload

```bash
# Upload entire music library
find ~/Music -name "*.mp3" -exec pmp300 upload {} \;

# Or use glob patterns
pmp300 upload ~/Music/**/*.mp3
```

### Backup All Files

```bash
# Download all files
mkdir pmp300_backup
cd pmp300_backup
pmp300 list | awk '{print $3}' | tail -n +4 | head -n -1 | \
  while read file; do pmp300 download "$file"; done
```

## Troubleshooting

### "device not specified"
Set the `PMP300_DEVICE` environment variable or use `--device` flag.

```bash
export PMP300_DEVICE=/dev/cu.usbmodem14201
```

### "failed to open Arduino"
- Check device path is correct
- Ensure Arduino is connected
- Check USB cable (some cables are power-only)
- Try unplugging and replugging Arduino

### "ping failed"
- Check Arduino firmware is uploaded
- Press Arduino reset button
- Wait 2-3 seconds and try again

### "initialization failed"
- Check all wiring between Arduino and DB-25
- Verify PMP300 is powered and connected
- Test with `pmp300 test` first

### Upload/Download Timeout
- Large files take time (7-9 minutes for 32MB)
- USB latency adds ~1-2ms per operation
- Ensure stable USB connection
- Don't disconnect during transfer

## Performance Notes

### Upload Speeds
- ~3.5 KB/s typical
- 1 MB file ≈ 5 minutes
- 32 MB ≈ 2.5 hours
- Limited by parallel port protocol and USB latency

### Download Speeds
- Similar to upload (~3.5 KB/s)
- Reading directory: ~30 seconds

### Bad Block Checking
- Tests each block with bit patterns
- ~10-15 seconds per block
- Full 32MB scan ≈ 3-4 hours
- Full 64MB scan ≈ 6-8 hours
- Only recommended for new devices or after errors

## Tips

1. **Use environment variable**: Set `PMP300_DEVICE` to avoid typing `--device` every time

2. **Check space first**: Use `pmp300 info` before uploading to verify free space

3. **Organize playback order**: Use `pmp300 move` to arrange songs in desired play order

4. **Backup before format**: Download all files before formatting

5. **Start with test**: Always run `pmp300 test` first when troubleshooting

6. **Filenames**: PMP300 supports up to 127-character filenames

7. **File types**: PMP300 plays MP3 files only (32-128 kbps recommended)

## Development

### Building from Source

```bash
git clone https://github.com/murdinc/pmp300
cd pmp300
go mod download
go build -o pmp300
```

### Running Tests

```bash
go test ./...
```

### Project Structure

```
pmp300/
├── main.go                  # Entry point
├── cmd/                     # Cobra commands
│   ├── root.go             # Root command
│   ├── test.go             # Test command
│   ├── list.go             # List command
│   ├── info.go             # Info command
│   ├── upload.go           # Upload command
│   ├── download.go         # Download command
│   ├── delete.go           # Delete command
│   ├── move.go             # Move command
│   └── format.go           # Format command
├── pkg/
│   ├── arduino/            # Arduino bridge protocol
│   │   └── arduino.go
│   └── pmp300/             # PMP300 protocol implementation
│       ├── pmp300.go       # Core protocol
│       ├── download.go     # Download operations
│       ├── upload.go       # Upload operations
│       ├── delete.go       # Delete operations
│       ├── reorder.go      # Reorder operations
│       └── format.go       # Format operations
└── arduino/                # Arduino firmware
    └── pmp300_usb_parallel_bridge/
```

## License

MIT License - See LICENSE file

## Credits

Based on PMP300 protocol reverse engineering by the community.

# Diamond Rio PMP300 Communication Protocol

Complete documentation for interfacing with the Diamond Rio PMP300 MP3 player via parallel port.

## Table of Contents

- [Supported Operations](#supported-operations)
- [Hardware Overview](#hardware-overview)
- [Parallel Port Pinout](#parallel-port-pinout)
- [Port Registers](#port-registers)
- [Protocol Specification](#protocol-specification)
- [Command Reference](#command-reference)
- [Data Structures](#data-structures)
- [Modern USB Interface Using Arduino](#modern-usb-interface-using-arduino)
- [Go Implementation Guide](#go-implementation-guide)
- [Example Code](#example-code)

## Supported Operations

### What You Can Do With The PMP300

The following operations are **fully documented** and supported via the parallel port protocol:

#### ✅ File Operations
- **Read/Download Files** - Download MP3 files from device to computer
- **Write/Upload Files** - Upload MP3 files from computer to device (32KB blocks with retry logic)
- **Delete File** - Remove individual files by name
- **Delete All Files** - Clear all files while preserving bad block markers

#### ✅ Directory Management
- **Read Directory** - Download directory/FAT table from block 0
- **List Files** - View all files with metadata (filename, size, timestamp, block position)
- **Reorder Files** - Change playback order of files in directory
- **Get Device Info** - View free/used/bad blocks, entry counts, timestamps

#### ✅ Device Management
- **Initialize/Format** - Prepare device for use (with or without bad block detection)
- **Mark Bad Blocks** - Test and flag defective storage blocks using bit patterns (0xAA, 0x55)
- **Check Device Present** - Verify connectivity via manufacturer/device ID codes
- **Switch Flash Memory** - Toggle between internal and external SmartMedia flash

#### ❌ Firmware Operations
- **Firmware Updates** - Not supported. No public protocol exists for updating firmware.

### Operation Details

| Operation | Complexity | Retry Logic | Checksum | Notes |
|-----------|------------|-------------|----------|-------|
| Read Directory | Medium | No | Yes (2 checksums) | Block 0, contains FAT |
| Upload File | High | Yes (3x) | Yes | 32KB blocks, ~7-9 min for 32MB |
| Download File | High | Yes (3x) | Yes | Validates each 512-byte chunk |
| Delete File | Medium | No | Yes | Updates FAT and directory |
| Initialize | High | No | Yes | Optional bad block scan |
| Bad Block Check | Very High | No | No | Tests with 0xAA/0x55 patterns |

## Hardware Overview

### Device Specifications

- **Model**: Diamond Rio PMP300
- **Interface**: PC Parallel Port (LPT)
- **Default Port Address**: 0x378 (alternative: 0x278, 0x3bc)
- **Storage**: 32MB internal flash (64MB on SE models)
- **Block Size**: 32KB (32,768 bytes)
- **Total Blocks**: 1024 (32MB) or 2048 (64MB SE)

### Connection Requirements

- DB-25 parallel port connector (female on PC, proprietary on PMP300)
- Bidirectional parallel port support
- Direct port I/O access (requires root/admin privileges)
- Voltage: 5V TTL logic levels

## Parallel Port Pinout

### DB-25 Connector (PC Side)

```
Pin | Signal Name    | Direction | Register | Bit | Description
----|----------------|-----------|----------|-----|---------------------------
1   | nStrobe        | Out       | Control  | 0   | Strobe (inverted)
2   | Data0          | Out       | Data     | 0   | Data bit 0 (LSB)
3   | Data1          | Out       | Data     | 1   | Data bit 1
4   | Data2          | Out       | Data     | 2   | Data bit 2
5   | Data3          | Out       | Data     | 3   | Data bit 3
6   | Data4          | Out       | Data     | 4   | Data bit 4
7   | Data5          | Out       | Data     | 5   | Data bit 5
8   | Data6          | Out       | Data     | 6   | Data bit 6
9   | Data7          | Out       | Data     | 7   | Data bit 7 (MSB)
10  | nAck           | In        | Status   | 6   | Acknowledge (inverted)
11  | Busy           | In        | Status   | 7   | Busy (inverted)
12  | Paper-Out      | In        | Status   | 5   | Paper Out
13  | Select         | In        | Status   | 4   | Select
14  | nAutoFeed      | Out       | Control  | 1   | Auto Feed (inverted)
15  | nError         | In        | Status   | 3   | Error (inverted)
16  | nInitialize    | Out       | Control  | 2   | Initialize (inverted)
17  | nSelect-In     | Out       | Control  | 3   | Select In (inverted)
18-25| Ground        | -         | -        | -   | Ground
```

**Note**: "n" prefix indicates inverted/active-low signals.

## Port Registers

### Register Map (Base Address + Offset)

| Register | Offset | Access | Description                    |
|----------|--------|--------|--------------------------------|
| Data     | +0     | R/W    | 8-bit data output/input        |
| Status   | +1     | R      | Device status signals          |
| Control  | +2     | R/W    | Control signals & handshaking  |

### Data Register (Base + 0)

```
Bit 7 6 5 4 3 2 1 0
    │ │ │ │ │ │ │ └─ Data0 (Pin 2)
    │ │ │ │ │ │ └─── Data1 (Pin 3)
    │ │ │ │ │ └───── Data2 (Pin 4)
    │ │ │ │ └─────── Data3 (Pin 5)
    │ │ │ └───────── Data4 (Pin 6)
    │ │ └─────────── Data5 (Pin 7)
    │ └───────────── Data6 (Pin 8)
    └─────────────── Data7 (Pin 9)
```

**Read/Write**: Writing sets output pins. Reading returns last written value.

### Status Register (Base + 1)

```
Bit 7 6 5 4 3 2 1 0
    │ │ │ │ │ │ │ └─ Reserved
    │ │ │ │ │ │ └─── Reserved
    │ │ │ │ │ └───── Reserved
    │ │ │ │ └─────── nError (Pin 15, inverted)
    │ │ │ └───────── Select (Pin 13)
    │ │ └─────────── Paper-Out (Pin 12)
    │ └───────────── nAck (Pin 10, inverted)
    └─────────────── Busy (Pin 11, inverted)
```

**Read-Only**: Returns device status. Used for handshaking and data reception.

### Control Register (Base + 2)

```
Bit 7 6 5 4 3 2 1 0
    │ │ │ │ │ │ │ └─ nStrobe (Pin 1, inverted)
    │ │ │ │ │ │ └─── nAutoFeed (Pin 14, inverted)
    │ │ │ │ │ └───── nInitialize (Pin 16, inverted)
    │ │ │ │ └─────── nSelect-In (Pin 17, inverted)
    │ │ │ └───────── IRQ Enable (not used)
    └─┴─┴─────────── Reserved
```

**Read/Write**: Controls handshaking signals.

## Protocol Specification

### Protocol Version

**Version**: 107 (0x6B)

### Timing Parameters

Platform-specific delays (in microseconds):

| Platform   | Init Delay | TX Delay | RX Delay |
|------------|------------|----------|----------|
| Windows NT | 2,000      | 2        | 10       |
| Linux/Unix | 20,000     | 15       | 100      |

**Note**: Adjust delays based on system performance. Faster systems may need increased delays.

### Command Structure

Commands are sent using a three-step sequence:

1. **Write Data Byte**: Write command byte to Data Register
2. **Strobe High**: Write 0x0C to Control Register
3. **Strobe Low**: Write 0x04 to Control Register

```go
// COMMANDOUT macro equivalent
func SendCommand(port uint16, cmd byte) {
    OutByte(port+0, cmd)        // Data register
    OutByte(port+2, 0x0C)       // Control: strobe high
    OutByte(port+2, 0x04)       // Control: strobe low
}
```

### Initialization Sequence

The device must be initialized before communication:

```
1. Set Control = 0x04
2. Send command 0xA8 (with control transitions)
3. Set Control = 0x00
4. Delay (initialization delay)
5. Set Control = 0x04
6. Delay (initialization delay)
7. Send command 0xAD
8. Send command 0x55
9. Send command 0xAE
10. Send command 0xAA
11. Send command 0xA8
```

### Data Reception

Reading a byte requires two nibble operations:

```go
func GetDataByte(port uint16) byte {
    var result byte

    // Get high nibble
    OutByte(port+2, 0x00)           // Control low
    status := InByte(port+1)        // Read status
    result = (status >> 4) & 0x0F   // Extract bits 4-7

    // Get low nibble
    OutByte(port+2, 0x04)           // Control high
    status = InByte(port+1)         // Read status
    result |= (status & 0x0F) << 4  // Extract bits 0-3

    // Reverse bit order
    result = ReverseBits(result)
    return result
}
```

### Handshaking

**Wait for Input**:
- Read Status Register
- Mask with 0xF8
- Wait for specific value or timeout (1 second)

**Wait for Acknowledge**:
- Check Status Register bit 3 (0x08)
- Timeout after 1 second

## Command Reference

### Low-Level Command Codes

All commands are sent via the parallel port using the `SendCommand()` sequence (data byte + control strobes).

#### Initialization Commands

```go
const (
    CMD_INIT_1 = 0xA8  // Initial synchronization
    CMD_INIT_2 = 0xAD  // Configuration setup
    CMD_INIT_3 = 0x55  // Alternating bit pattern (01010101)
    CMD_INIT_4 = 0xAE  // Setup variant
    CMD_INIT_5 = 0xAA  // Alternating bit pattern (10101010)
)
```

**Initialization Sequence**:
```
1. Set Control = 0x04
2. SendCommand(0xA8)
3. Set Control = 0x00, Delay, Set Control = 0x04, Delay
4. SendCommand(0xAD)
5. SendCommand(0x55)
6. SendCommand(0xAE)
7. SendCommand(0xAA)
8. SendCommand(0xA8)
```

#### Block Transfer Commands

```go
const (
    CMD_BLOCK_INIT   = 0xAB  // Initialize block transfer / Read device ID
    CMD_SET_ADDR_1   = 0xA1  // Set block address (part 1)
    CMD_SET_ADDR_2   = 0xA2  // Set block address (part 2)
    CMD_READ_BLOCK   = 0xA0  // Read 32KB block
    CMD_WRITE_MODE   = 0xD0  // Enter write mode
    CMD_BLOCK_MODE   = 0x60  // Block mode initialization
    CMD_TRANSFER_CTL = 0x70  // Data transfer control
    CMD_BLOCK_END    = 0x10  // Block completion marker
    CMD_GET_ID       = 0x90  // Read manufacturer/device ID
)
```

#### Block Addressing

For reading/writing 32KB blocks, you need to send a 3-byte address:

```go
// Calculate block address (example for block N)
func GetBlockAddress(blockNum uint16, externalFlash bool) (hi, mid, lo byte) {
    addr := uint32(blockNum) * BLOCK_SIZE

    if externalFlash {
        // External flash uses different addressing
        addr += 0x200000  // Offset for external
    }

    hi = byte((addr >> 16) & 0xFF)
    mid = byte((addr >> 8) & 0xFF)
    lo = byte(addr & 0xFF)
    return
}

// Send read command for block N
SendCommand(CMD_READ_BLOCK)
SendCommand(hi)
SendCommand(mid)
SendCommand(lo)
```

### Operation Command Sequences

#### Read Directory (Block 0)

```go
func ReadDirectory() {
    // 1. Initialize communication
    Initialize()

    // 2. Send read command for block 0
    SendCommand(0xA0)
    SendCommand(0x00)  // Hi byte
    SendCommand(0x00)  // Mid byte
    SendCommand(0x00)  // Lo byte

    // 3. Wait for device ready
    WaitForInput(0xF8, 5*time.Second)

    // 4. Read 32KB (32768 bytes)
    for i := 0; i < 32768; i++ {
        data[i] = GetDataByte()
    }

    // 5. Validate checksums
    ValidateDirectoryChecksums()
}
```

#### Upload File (Write Blocks)

```go
func UploadFile(filename string, data []byte) {
    // 1. Calculate blocks needed
    numBlocks := (len(data) + BLOCK_SIZE - 1) / BLOCK_SIZE

    // 2. Find free blocks in FAT
    freeBlocks := FindFreeBlocks(numBlocks)

    // 3. Write each 32KB block
    for i, blockNum := range freeBlocks {
        blockData := data[i*BLOCK_SIZE : min((i+1)*BLOCK_SIZE, len(data))]
        WriteBlock(blockNum, blockData)
    }

    // 4. Update directory entry
    AddDirectoryEntry(filename, freeBlocks, len(data))

    // 5. Write updated directory to block 0
    WriteDirectory()
}

func WriteBlock(blockNum uint16, data []byte) {
    hi, mid, lo := GetBlockAddress(blockNum, false)

    // Send write command
    SendCommand(0xAB)
    SendCommand(hi)
    SendCommand(mid)
    SendCommand(lo)

    // Write data in 512-byte chunks
    for chunk := 0; chunk < 64; chunk++ {
        chunkData := data[chunk*512 : (chunk+1)*512]

        for _, b := range chunkData {
            SendCommand(b)
        }

        // Send end-of-chunk marker with checksum
        WriteEndMarker(chunk, checksum)

        WaitForAck(1 * time.Second)
    }
}
```

#### Download File (Read Blocks)

```go
func DownloadFile(entry DirectoryEntry) []byte {
    fileData := make([]byte, entry.TotalSize)

    // Read each block
    for i := 0; i < int(entry.BlockCount); i++ {
        blockNum := entry.BlockPosition + uint16(i)
        blockData := ReadBlock(blockNum)

        // Copy block data to file buffer
        offset := i * BLOCK_SIZE
        bytesToCopy := min(BLOCK_SIZE, int(entry.TotalSize)-offset)
        copy(fileData[offset:], blockData[:bytesToCopy])
    }

    return fileData
}
```

#### Delete File

```go
func DeleteFile(filename string) {
    // 1. Read directory
    dir := ReadDirectory()

    // 2. Find file entry
    entry, index := FindFile(dir, filename)

    // 3. Mark blocks as free in FAT
    for i := 0; i < int(entry.BlockCount); i++ {
        blockNum := entry.BlockPosition + uint16(i)
        dir.BlockUsage[blockNum] = BLOCK_FREE
    }

    // 4. Remove directory entry (shift remaining entries)
    RemoveDirectoryEntry(dir, index)

    // 5. Update checksums
    dir.Header.Checksum = CalculateHeaderChecksum(dir.Header)

    // 6. Write directory back to block 0
    WriteDirectory(dir)
}
```

#### Initialize/Format Device

```go
func InitializeDevice(checkBadBlocks bool) {
    // 1. Initialize communication
    Initialize()

    // 2. Create empty directory
    dir := DirectoryBlock{
        Header: DirectoryHeader{
            EntryCount:  0,
            FreeBlocks:  1024,  // or 2048 for SE
            BadBlocks:   0,
            TotalBlocks: 1024,
            Version:     107,
        },
    }

    // 3. Initialize FAT (all blocks free except block 0)
    dir.BlockUsage[0] = BLOCK_USED  // Directory block
    for i := 1; i < len(dir.BlockUsage); i++ {
        dir.BlockUsage[i] = BLOCK_FREE
    }

    // 4. Optional: Check for bad blocks
    if checkBadBlocks {
        for i := 1; i < len(dir.BlockUsage); i++ {
            if IsBadBlock(i) {
                dir.BlockUsage[i] = BLOCK_BAD
                dir.Header.BadBlocks++
                dir.Header.FreeBlocks--
            }
        }
    }

    // 5. Calculate checksums
    dir.Header.Checksum = CalculateHeaderChecksum(dir.Header)

    // 6. Write directory to block 0
    WriteBlock(0, SerializeDirectory(dir))
}

func IsBadBlock(blockNum int) bool {
    // Test with alternating patterns
    pattern1 := bytes.Repeat([]byte{0xAA}, BLOCK_SIZE)
    pattern2 := bytes.Repeat([]byte{0x55}, BLOCK_SIZE)

    // Write and read back pattern 1
    WriteBlock(uint16(blockNum), pattern1)
    result1 := ReadBlock(uint16(blockNum))

    // Write and read back pattern 2
    WriteBlock(uint16(blockNum), pattern2)
    result2 := ReadBlock(uint16(blockNum))

    // Check if patterns match
    return !bytes.Equal(pattern1, result1) || !bytes.Equal(pattern2, result2)
}
```

#### Check Device Present

```go
func CheckDevicePresent() bool {
    // Send device ID command
    SendCommand(0x90)

    // Read manufacturer and device codes
    manufacturer := GetDataByte()
    device := GetDataByte()

    // Expected values for PMP300 flash chip
    // (actual values depend on flash manufacturer)
    return manufacturer != 0xFF && device != 0xFF
}
```

### Retry Logic

All block transfers implement retry logic:

```go
const MAX_RETRIES = 3

func WriteBlockWithRetry(blockNum uint16, data []byte) error {
    for attempt := 0; attempt < MAX_RETRIES; attempt++ {
        err := WriteBlock(blockNum, data)
        if err == nil {
            return nil
        }

        // Wait 1 second before retry
        time.Sleep(1 * time.Second)
    }

    return fmt.Errorf("failed after %d attempts", MAX_RETRIES)
}
```

## Data Structures

### Block Types

```go
const (
    BLOCK_USED = 0x00  // Block contains data
    BLOCK_BAD  = 0x0F  // Bad block (skip)
    BLOCK_FREE = 0xFF  // Available block
)

const BLOCK_SIZE = 32768  // 32KB blocks
```

### Directory Header

```go
type DirectoryHeader struct {
    EntryCount      uint16    // Number of files
    FreeBlocks      uint16    // Available 32KB blocks
    BadBlocks       uint16    // Defective blocks
    TotalBlocks     uint16    // Total blocks on device
    Timestamp       [6]byte   // Device timestamp
    Checksum        uint16    // Header checksum
    Version         uint8     // Protocol version (107)
    Reserved        [13]byte  // Padding
}
```

**Size**: 32 bytes

### Directory Entry

```go
type DirectoryEntry struct {
    BlockPosition   uint16    // Starting block number
    BlockCount      uint16    // Number of blocks used
    BlockSizeMod    uint16    // Bytes used in last block
    TotalSize       uint32    // Total file size in bytes
    Timestamp       [6]byte   // Upload timestamp (YYMMDDHHMMSS)
    Filename        [128]byte // Null-terminated filename
}
```

**Size**: 144 bytes
**Max Entries per Directory**: 60

### Directory Block

```go
type DirectoryBlock struct {
    Header          DirectoryHeader      // 32 bytes
    Entries         [60]DirectoryEntry   // 8640 bytes
    BlockUsage      [8192]byte          // FAT: one byte per block
    // Padding to 32KB
}
```

**Size**: 32,768 bytes (one block)

### End Block Marker (512-byte chunks)

```go
type EndBlockMarker struct {
    PrevBlock       uint16    // Previous block number (0xFFFF if first)
    NextBlock       uint16    // Next block number (0xFFFF if last)
    ChunkIndex      uint16    // Chunk number (0-63 for 32KB)
    Reserved        [506]byte // Padding
    Checksum        uint16    // Chunk checksum
}
```

**Size**: 512 bytes
**Chunks per 32KB block**: 64

## Go Implementation Guide

### Prerequisites

1. **Direct Port I/O Access**

You'll need a library for direct port I/O. Options:

- **Linux**: Use `/dev/port` or `ioperm()`/`iopl()` syscalls
- **Windows**: Use a kernel driver (e.g., WinIO, InpOut32)
- **Cross-platform**: Consider using cgo with platform-specific libraries

2. **Permissions**

- Linux: Run as root or use `setcap` for port access
- Windows: Requires admin privileges and driver installation

### Port I/O Functions

```go
package main

import (
    "fmt"
    "os"
    "syscall"
    "time"
    "unsafe"
)

// Linux-specific port I/O using /dev/port
type ParallelPort struct {
    baseAddr uint16
    portFile *os.File
}

func NewParallelPort(baseAddr uint16) (*ParallelPort, error) {
    // Open /dev/port for direct hardware access (requires root)
    file, err := os.OpenFile("/dev/port", os.O_RDWR, 0)
    if err != nil {
        return nil, fmt.Errorf("failed to open /dev/port: %v (requires root)", err)
    }

    return &ParallelPort{
        baseAddr: baseAddr,
        portFile: file,
    }, nil
}

func (p *ParallelPort) Close() error {
    return p.portFile.Close()
}

func (p *ParallelPort) OutByte(offset uint16, value byte) error {
    _, err := p.portFile.WriteAt([]byte{value}, int64(p.baseAddr+offset))
    return err
}

func (p *ParallelPort) InByte(offset uint16) (byte, error) {
    buf := make([]byte, 1)
    _, err := p.portFile.ReadAt(buf, int64(p.baseAddr+offset))
    return buf[0], err
}
```

### Alternative: Using ioperm syscall (Linux)

```go
// #include <sys/io.h>
// #cgo LDFLAGS: -static
import "C"

func EnablePortAccess(baseAddr uint16) error {
    // Request permission for 3 ports (data, status, control)
    ret := C.ioperm(C.ulong(baseAddr), 3, 1)
    if ret != 0 {
        return fmt.Errorf("ioperm failed (requires root)")
    }
    return nil
}

func OutByte(port uint16, value byte) {
    C.outb(C.uchar(value), C.ushort(port))
}

func InByte(port uint16) byte {
    return byte(C.inb(C.ushort(port)))
}
```

## Example Code

### Basic Initialization

```go
package main

import (
    "fmt"
    "time"
)

const (
    DEFAULT_BASE_ADDR = 0x378

    // Register offsets
    DATA_REG    = 0
    STATUS_REG  = 1
    CONTROL_REG = 2

    // Timing delays (microseconds)
    INIT_DELAY = 20000
    TX_DELAY   = 15
    RX_DELAY   = 100
)

type RioPMP300 struct {
    port *ParallelPort
}

func NewRioPMP300(baseAddr uint16) (*RioPMP300, error) {
    port, err := NewParallelPort(baseAddr)
    if err != nil {
        return nil, err
    }

    return &RioPMP300{port: port}, nil
}

func (r *RioPMP300) SendCommand(cmd byte) error {
    // Write command to data register
    if err := r.port.OutByte(DATA_REG, cmd); err != nil {
        return err
    }

    // Strobe high
    if err := r.port.OutByte(CONTROL_REG, 0x0C); err != nil {
        return err
    }

    // Strobe low
    if err := r.port.OutByte(CONTROL_REG, 0x04); err != nil {
        return err
    }

    time.Sleep(time.Duration(TX_DELAY) * time.Microsecond)
    return nil
}

func (r *RioPMP300) Initialize() error {
    fmt.Println("Initializing Rio PMP300...")

    // Step 1: Set control to 0x04
    if err := r.port.OutByte(CONTROL_REG, 0x04); err != nil {
        return err
    }

    // Step 2: Send 0xA8
    if err := r.SendCommand(0xA8); err != nil {
        return err
    }

    // Step 3: Set control to 0x00
    if err := r.port.OutByte(CONTROL_REG, 0x00); err != nil {
        return err
    }

    // Step 4: Delay
    time.Sleep(time.Duration(INIT_DELAY) * time.Microsecond)

    // Step 5: Set control to 0x04
    if err := r.port.OutByte(CONTROL_REG, 0x04); err != nil {
        return err
    }

    // Step 6: Delay
    time.Sleep(time.Duration(INIT_DELAY) * time.Microsecond)

    // Step 7-11: Send initialization sequence
    initSeq := []byte{0xAD, 0x55, 0xAE, 0xAA, 0xA8}
    for _, cmd := range initSeq {
        if err := r.SendCommand(cmd); err != nil {
            return err
        }
    }

    fmt.Println("Initialization complete!")
    return nil
}

func (r *RioPMP300) GetDataByte() (byte, error) {
    var result byte

    // Get high nibble
    if err := r.port.OutByte(CONTROL_REG, 0x00); err != nil {
        return 0, err
    }

    status, err := r.port.InByte(STATUS_REG)
    if err != nil {
        return 0, err
    }
    result = (status >> 4) & 0x0F

    // Get low nibble
    if err := r.port.OutByte(CONTROL_REG, 0x04); err != nil {
        return 0, err
    }

    status, err = r.port.InByte(STATUS_REG)
    if err != nil {
        return 0, err
    }
    result |= (status & 0x0F) << 4

    // Reverse bit order
    result = reverseBits(result)

    time.Sleep(time.Duration(RX_DELAY) * time.Microsecond)
    return result, nil
}

func reverseBits(b byte) byte {
    b = (b&0xF0)>>4 | (b&0x0F)<<4
    b = (b&0xCC)>>2 | (b&0x33)<<2
    b = (b&0xAA)>>1 | (b&0x55)<<1
    return b
}

func (r *RioPMP300) WaitForInput(expected byte, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        status, err := r.port.InByte(STATUS_REG)
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

func (r *RioPMP300) WaitForAck(timeout time.Duration) error {
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        status, err := r.port.InByte(STATUS_REG)
        if err != nil {
            return err
        }

        if (status & 0x08) != 0 {
            return nil
        }

        time.Sleep(10 * time.Millisecond)
    }

    return fmt.Errorf("timeout waiting for acknowledge")
}

func (r *RioPMP300) Close() error {
    return r.port.Close()
}
```

### Reading Directory

```go
func (r *RioPMP300) ReadDirectory() (*DirectoryBlock, error) {
    // Send read directory command (block 0)
    if err := r.SendCommand(0xA0); err != nil {
        return nil, err
    }

    // Send block address (0x00, 0x00, 0x00)
    if err := r.SendCommand(0x00); err != nil {
        return nil, err
    }
    if err := r.SendCommand(0x00); err != nil {
        return nil, err
    }
    if err := r.SendCommand(0x00); err != nil {
        return nil, err
    }

    // Wait for device ready
    if err := r.WaitForInput(0xF8, 5*time.Second); err != nil {
        return nil, err
    }

    // Read 32KB directory block
    dirData := make([]byte, BLOCK_SIZE)
    for i := 0; i < BLOCK_SIZE; i++ {
        b, err := r.GetDataByte()
        if err != nil {
            return nil, err
        }
        dirData[i] = b
    }

    // Parse directory structure
    dir := &DirectoryBlock{}
    // ... parse dirData into DirectoryBlock struct ...

    return dir, nil
}
```

### Complete Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // Initialize Rio PMP300 on default parallel port
    rio, err := NewRioPMP300(DEFAULT_BASE_ADDR)
    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }
    defer rio.Close()

    // Initialize device
    if err := rio.Initialize(); err != nil {
        log.Fatalf("Initialization failed: %v", err)
    }

    // Read directory
    dir, err := rio.ReadDirectory()
    if err != nil {
        log.Fatalf("Failed to read directory: %v", err)
    }

    fmt.Printf("Device has %d files\n", dir.Header.EntryCount)
    fmt.Printf("Free blocks: %d\n", dir.Header.FreeBlocks)
    fmt.Printf("Bad blocks: %d\n", dir.Header.BadBlocks)

    // List files
    for i := 0; i < int(dir.Header.EntryCount); i++ {
        entry := dir.Entries[i]
        filename := string(entry.Filename[:])
        fmt.Printf("%d. %s (%d bytes)\n", i+1, filename, entry.TotalSize)
    }
}
```

## Modern USB Interface Using Arduino

For modern computers without parallel ports, you can use an Arduino as a USB-to-parallel bridge. This approach provides native 5V I/O without the need for vintage hardware or complex USB adapters.

### Hardware Requirements

**Recommended Boards:**
- **Arduino Mega 2560**: 54 digital I/O pins (plenty of headroom)
- **Arduino Uno**: 20 total pins (18 available after USB serial)

**Additional Components:**
- DB-25 female connector (to accept the original PMP300 cable's male connector)
- Jumper wires or solder for connections
- Optional: DB-25 breakout board for easier wiring

**Connector Type:**
- Use a **DB-25 female connector** - this is what a PC parallel port looks like
- Your existing PMP300 cable has a DB-25 male connector that plugs into the PC
- The Arduino acts as a replacement for the PC's parallel port

### Pin Requirements

Based on the protocol analysis (see Protocol Specification section), you need:

| Signal Type | Count | Direction | Notes |
|-------------|-------|-----------|-------|
| Data | 8 | Bidirectional | All 8 bits used |
| Control | 2 | Output | Only bits 2,3 used (nInitialize, nSelect-In) |
| Status | 5 | Input | Bits 3-7 (nError, Select, Paper-Out, nAck, Busy) |
| Ground | 1+ | - | Connect to DB-25 pins 18-25 |
| **Total** | **15** | - | Plus ground |

### Arduino Mega 2560 Wiring

The Mega provides plenty of pins with excellent organization:

```
Signal          Arduino Pin   DB-25 Pin   Direction   Description
-------------   -----------   ---------   ---------   ---------------------------
Data0           22            2           I/O         Data bit 0 (LSB)
Data1           23            3           I/O         Data bit 1
Data2           24            4           I/O         Data bit 2
Data3           25            5           I/O         Data bit 3
Data4           26            6           I/O         Data bit 4
Data5           27            7           I/O         Data bit 5
Data6           28            8           I/O         Data bit 6
Data7           29            9           I/O         Data bit 7 (MSB)

nInitialize     30            16          OUT         Control bit 2 (inverted)
nSelect-In      31            17          OUT         Control bit 3 (inverted)

nError          32            15          IN          Status bit 3 (inverted)
Select          33            13          IN          Status bit 4
Paper-Out       34            12          IN          Status bit 5
nAck            35            10          IN          Status bit 6 (inverted)
Busy            36            11          IN          Status bit 7 (inverted)

Ground          GND           18-25       -           Connect all ground pins
```

**Unused Control Signals:**
- nStrobe (Pin 1) - Not used by protocol
- nAutoFeed (Pin 14) - Not used by protocol

### Arduino Uno Wiring

The Uno has fewer pins but still sufficient (18 available after USB serial):

```
Signal          Arduino Pin   DB-25 Pin   Direction   Description
-------------   -----------   ---------   ---------   ---------------------------
Data0           2             2           I/O         Data bit 0 (LSB)
Data1           3             3           I/O         Data bit 1
Data2           4             4           I/O         Data bit 2
Data3           5             5           I/O         Data bit 3
Data4           6             6           I/O         Data bit 4
Data5           7             7           I/O         Data bit 5
Data6           8             8           I/O         Data bit 6
Data7           9             9           I/O         Data bit 7 (MSB)

nInitialize     10            16          OUT         Control bit 2 (inverted)
nSelect-In      11            17          OUT         Control bit 3 (inverted)

nError          12            15          IN          Status bit 3 (inverted)
Select          13            13          IN          Status bit 4
Paper-Out       A0            12          IN          Status bit 5
nAck            A1            10          IN          Status bit 6 (inverted)
Busy            A2            11          IN          Status bit 7 (inverted)

Ground          GND           18-25       -           Connect all ground pins
```

**Note:** Pins 0 (RX) and 1 (TX) are reserved for USB serial communication with your computer.

### Voltage Compatibility

✅ **Perfect Match - No Level Shifters Needed:**
- Arduino Mega/Uno: 5V TTL logic
- PMP300 Parallel Port: 5V TTL logic (see Hardware Overview)
- Direct connection is safe and recommended

### Arduino Firmware Architecture

Your Arduino firmware needs to implement:

1. **USB Serial Command Interface**
   - Receive commands from Mac/PC over USB serial
   - Send responses and status back

2. **Low-Level Port I/O**
   - Write to data pins (8 bits)
   - Write to control pins (2 bits)
   - Read from status pins (5 bits)
   - Implement microsecond-precision delays (see Timing Parameters)

3. **Command Protocol**

Example command protocol:
```
Commands (Mac → Arduino):
'W' <port> <byte>      - Write byte to port (0=Data, 2=Control)
'R' <port>             - Read byte from port (1=Status)
'D' <microseconds>     - Delay for N microseconds
'I'                    - Initialize PMP300

Responses (Arduino → Mac):
'K'                    - Command acknowledged
'V' <byte>             - Value read from port
'E' <code>             - Error code
```

### Example Arduino Firmware Skeleton

```cpp
// Pin definitions for Arduino Mega
#define DATA0  22
#define DATA1  23
// ... define all pins ...

void setup() {
  Serial.begin(115200);  // USB serial to computer

  // Configure data pins as outputs initially
  pinMode(DATA0, OUTPUT);
  pinMode(DATA1, OUTPUT);
  // ... configure all pins ...

  // Configure control pins as outputs
  pinMode(NINIT, OUTPUT);
  pinMode(NSELECT, OUTPUT);

  // Configure status pins as inputs
  pinMode(NERROR, INPUT);
  pinMode(SELECT, INPUT);
  // ... configure all status pins ...
}

void loop() {
  if (Serial.available()) {
    char cmd = Serial.read();

    switch(cmd) {
      case 'W':  // Write to port
        handleWrite();
        break;
      case 'R':  // Read from port
        handleRead();
        break;
      case 'D':  // Delay
        handleDelay();
        break;
      case 'I':  // Initialize
        handleInit();
        break;
    }
  }
}

void writeDataByte(uint8_t value) {
  // Set data pins as outputs
  pinMode(DATA0, OUTPUT);
  // ... set all data pins as OUTPUT ...

  // Write each bit
  digitalWrite(DATA0, value & 0x01);
  digitalWrite(DATA1, value & 0x02);
  // ... write all 8 bits ...
}

uint8_t readStatusByte() {
  uint8_t result = 0;

  // Read status pins and construct byte
  if (digitalRead(NERROR)) result |= 0x08;
  if (digitalRead(SELECT))  result |= 0x10;
  if (digitalRead(PAPEROUT)) result |= 0x20;
  if (digitalRead(NACK))    result |= 0x40;
  if (digitalRead(BUSY))    result |= 0x80;

  return result;
}

void writeControlBits(uint8_t value) {
  // Only bits 2 and 3 are used
  digitalWrite(NINIT, value & 0x04);
  digitalWrite(NSELECT, value & 0x08);
}
```

### Mac/PC Software Integration

Your host software communicates with Arduino over USB serial:

```go
package main

import (
    "github.com/tarm/serial"
    "time"
)

type ArduinoPort struct {
    port *serial.Port
}

func NewArduinoPort(device string) (*ArduinoPort, error) {
    config := &serial.Config{
        Name: device,        // e.g., "/dev/cu.usbmodem14201" on Mac
        Baud: 115200,
    }

    port, err := serial.OpenPort(config)
    if err != nil {
        return nil, err
    }

    return &ArduinoPort{port: port}, nil
}

func (a *ArduinoPort) OutByte(portOffset uint8, value byte) error {
    // Send write command to Arduino
    cmd := []byte{'W', portOffset, value}
    _, err := a.port.Write(cmd)
    if err != nil {
        return err
    }

    // Wait for acknowledgment
    response := make([]byte, 1)
    _, err = a.port.Read(response)
    return err
}

func (a *ArduinoPort) InByte(portOffset uint8) (byte, error) {
    // Send read command to Arduino
    cmd := []byte{'R', portOffset}
    _, err := a.port.Write(cmd)
    if err != nil {
        return 0, err
    }

    // Wait for value response
    response := make([]byte, 2)  // 'V' + byte value
    _, err = a.port.Read(response)
    if err != nil {
        return 0, err
    }

    return response[1], nil
}

func (a *ArduinoPort) Delay(microseconds int) {
    time.Sleep(time.Duration(microseconds) * time.Microsecond)
}

func (a *ArduinoPort) Close() error {
    return a.port.Close()
}

// Use with PMP300 protocol implementation
func main() {
    arduino, err := NewArduinoPort("/dev/cu.usbmodem14201")
    if err != nil {
        panic(err)
    }
    defer arduino.Close()

    // Now use arduino.OutByte() and arduino.InByte()
    // exactly like the parallel port I/O in the protocol examples
}
```

### Bill of Materials

| Item | Quantity | Estimated Cost |
|------|----------|----------------|
| Arduino Mega 2560 (or Uno) | 1 | $20-30 ($15-20 for Uno) |
| DB-25 Female Connector | 1 | $3-5 |
| Jumper Wires or Wire | ~20 | $5-10 |
| Optional: DB-25 Breakout Board | 1 | $5-10 |
| **Total** | | **$28-55** |

### Assembly Tips

1. **Use a breadboard first** to test connections before soldering
2. **Label all wires** - 15+ connections are easy to mix up
3. **Test continuity** with a multimeter before powering on
4. **Start simple** - Test data lines first, then add control/status
5. **Add LEDs** to Arduino outputs for visual debugging
6. **Keep wires short** to minimize noise and signal degradation

### Troubleshooting

**Device not responding:**
- Check ground connection between Arduino and DB-25
- Verify 5V power to Arduino
- Test with multimeter: data/control pins should read 0V or 5V

**Timing issues:**
- Increase delay values in firmware
- Check USB serial baud rate (115200 recommended)
- Monitor serial traffic for command/response sync

**Incorrect data:**
- Verify pin mapping matches wiring
- Check for inverted signals (nError, nAck, nInitialize, etc.)
- Use Arduino Serial.print() to debug values

## References

- [Snowblind Alliance RIO Utility v1.07](http://slackware.cs.utah.edu/pub/slackware/slackware-8.0/contrib/rio.txt)
- [wfx_rio GitHub Repository](https://github.com/creaktive/wfx_rio)
- [OSDev Wiki - Parallel Port](http://wiki.osdev.org/Parallel_port)
- [IEEE 1284 Standard](https://www.ardent-tool.com/comms/an062_Updating_the_parallel_port.pdf)
- [Beyond Logic - Interfacing the Parallel Port](http://wearcam.org/seatsale/programs/www.beyondlogic.org/spp/parallel.htm)

## License

This documentation is provided for educational and preservation purposes. The Rio PMP300 protocol is based on reverse-engineered information from open-source implementations.

---

**Note**: Direct hardware access requires elevated privileges. Use caution when working with hardware interfaces.

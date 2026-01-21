#!/bin/bash
# find-device.sh - Helps find Arduino serial device

echo "Searching for Arduino serial devices..."
echo

# macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    devices=$(ls /dev/cu.usbmodem* 2>/dev/null || ls /dev/cu.usbserial* 2>/dev/null)
    if [ -n "$devices" ]; then
        echo "Found device(s) on macOS:"
        echo "$devices"
        echo
        first_device=$(echo "$devices" | head -n1)
        echo "To use with pmp300:"
        echo "  export PMP300_DEVICE=$first_device"
        echo "  pmp300 test"
    else
        echo "No Arduino devices found on macOS."
        echo "Expected paths:"
        echo "  /dev/cu.usbmodem*"
        echo "  /dev/cu.usbserial*"
    fi

# Linux
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    devices=$(ls /dev/ttyACM* 2>/dev/null || ls /dev/ttyUSB* 2>/dev/null)
    if [ -n "$devices" ]; then
        echo "Found device(s) on Linux:"
        echo "$devices"
        echo
        first_device=$(echo "$devices" | head -n1)
        echo "To use with pmp300:"
        echo "  export PMP300_DEVICE=$first_device"
        echo "  pmp300 test"
    else
        echo "No Arduino devices found on Linux."
        echo "Expected paths:"
        echo "  /dev/ttyACM*"
        echo "  /dev/ttyUSB*"
    fi

# Windows (Git Bash/WSL)
else
    echo "Windows users:"
    echo "  1. Open Device Manager"
    echo "  2. Look under 'Ports (COM & LPT)'"
    echo "  3. Find 'Arduino' or 'USB Serial Device'"
    echo "  4. Note the COM port (e.g., COM3)"
    echo
    echo "To use with pmp300:"
    echo "  export PMP300_DEVICE=COM3"
    echo "  pmp300 test"
fi

echo
echo "Troubleshooting:"
echo "  - Ensure Arduino is connected via USB"
echo "  - Try unplugging and replugging the Arduino"
echo "  - Check that firmware is uploaded to Arduino"
echo "  - Some USB cables are power-only (try a different cable)"

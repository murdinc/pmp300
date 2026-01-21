/*
 * PMP300 USB-to-Parallel Bridge
 *
 * This Arduino sketch implements a USB-to-5V-parallel-port bridge for interfacing
 * with the Diamond Rio PMP300 MP3 player on modern computers without parallel ports.
 *
 * Hardware: Arduino Mega 2560 or Arduino Uno
 * Interface: USB Serial at 115200 baud
 *
 * Protocol Documentation:
 * See PROTOCOL.md in this directory for complete command reference.
 *
 * Author: Based on PMP300 protocol reverse engineering
 * License: MIT
 */

// ============================================================================
// BOARD CONFIGURATION - Automatically detects Arduino Mega vs Uno
// ============================================================================

#if defined(__AVR_ATmega2560__)
  // Arduino Mega 2560 Pin Definitions
  #define BOARD_TYPE "Arduino Mega 2560"

  // Data pins (8 bits, bidirectional)
  #define DATA0_PIN  22
  #define DATA1_PIN  23
  #define DATA2_PIN  24
  #define DATA3_PIN  25
  #define DATA4_PIN  26
  #define DATA5_PIN  27
  #define DATA6_PIN  28
  #define DATA7_PIN  29

  // Control pins (2 bits, output only)
  #define CTRL_NINIT_PIN      30  // Control bit 2 (nInitialize)
  #define CTRL_NSELECT_PIN    31  // Control bit 3 (nSelect-In)

  // Status pins (5 bits, input only)
  #define STATUS_NERROR_PIN   32  // Status bit 3 (nError)
  #define STATUS_SELECT_PIN   33  // Status bit 4 (Select)
  #define STATUS_PAPEROUT_PIN 34  // Status bit 5 (Paper-Out)
  #define STATUS_NACK_PIN     35  // Status bit 6 (nAck)
  #define STATUS_BUSY_PIN     36  // Status bit 7 (Busy)

#else
  // Arduino Uno Pin Definitions
  #define BOARD_TYPE "Arduino Uno"

  // Data pins (8 bits, bidirectional)
  #define DATA0_PIN  2
  #define DATA1_PIN  3
  #define DATA2_PIN  4
  #define DATA3_PIN  5
  #define DATA4_PIN  6
  #define DATA5_PIN  7
  #define DATA6_PIN  8
  #define DATA7_PIN  9

  // Control pins (2 bits, output only)
  #define CTRL_NINIT_PIN      10  // Control bit 2 (nInitialize)
  #define CTRL_NSELECT_PIN    11  // Control bit 3 (nSelect-In)

  // Status pins (5 bits, input only)
  #define STATUS_NERROR_PIN   12  // Status bit 3 (nError)
  #define STATUS_SELECT_PIN   13  // Status bit 4 (Select)
  #define STATUS_PAPEROUT_PIN A0  // Status bit 5 (Paper-Out)
  #define STATUS_NACK_PIN     A1  // Status bit 6 (nAck)
  #define STATUS_BUSY_PIN     A2  // Status bit 7 (Busy)
#endif

// ============================================================================
// PROTOCOL CONSTANTS
// ============================================================================

#define SERIAL_BAUD_RATE 115200

// Command bytes (Host -> Arduino)
#define CMD_WRITE_DATA    'W'  // Write byte to data register
#define CMD_WRITE_CTRL    'C'  // Write byte to control register
#define CMD_READ_STATUS   'R'  // Read byte from status register
#define CMD_DELAY_US      'D'  // Delay microseconds
#define CMD_DELAY_MS      'M'  // Delay milliseconds
#define CMD_PING          'P'  // Ping (connection test)
#define CMD_VERSION       'V'  // Get version info
#define CMD_SET_DATA_DIR  'S'  // Set data pin direction (I=input, O=output)

// Response bytes (Arduino -> Host)
#define RESP_OK           'K'  // Command acknowledged/success
#define RESP_VALUE        'V'  // Value follows (1 byte)
#define RESP_ERROR        'E'  // Error code follows (1 byte)
#define RESP_PONG         'P'  // Pong response
#define RESP_VERSION      'I'  // Version info follows

// Error codes
#define ERR_UNKNOWN_CMD   0x01  // Unknown command received
#define ERR_TIMEOUT       0x02  // Timeout waiting for data
#define ERR_INVALID_PARAM 0x03  // Invalid parameter

// Firmware version
#define FW_VERSION_MAJOR  1
#define FW_VERSION_MINOR  0
#define FW_VERSION_PATCH  0

// ============================================================================
// PIN ARRAYS FOR EFFICIENT PROCESSING
// ============================================================================

const uint8_t dataPins[8] = {
  DATA0_PIN, DATA1_PIN, DATA2_PIN, DATA3_PIN,
  DATA4_PIN, DATA5_PIN, DATA6_PIN, DATA7_PIN
};

const uint8_t statusPins[5] = {
  STATUS_NERROR_PIN,   // Bit 3
  STATUS_SELECT_PIN,   // Bit 4
  STATUS_PAPEROUT_PIN, // Bit 5
  STATUS_NACK_PIN,     // Bit 6
  STATUS_BUSY_PIN      // Bit 7
};

// ============================================================================
// GLOBAL STATE
// ============================================================================

bool dataDirection = OUTPUT; // Current direction of data pins (INPUT or OUTPUT)

// ============================================================================
// SETUP - Initialize hardware and serial communication
// ============================================================================

void setup() {
  // Initialize USB serial
  Serial.begin(SERIAL_BAUD_RATE);

  // Wait for serial port to connect (important for Leonardo/Micro)
  while (!Serial && millis() < 3000) {
    ; // Wait up to 3 seconds
  }

  // Initialize data pins as outputs initially
  setDataDirection(OUTPUT);

  // Initialize control pins as outputs, set to idle state
  pinMode(CTRL_NINIT_PIN, OUTPUT);
  pinMode(CTRL_NSELECT_PIN, OUTPUT);
  digitalWrite(CTRL_NINIT_PIN, LOW);
  digitalWrite(CTRL_NSELECT_PIN, LOW);

  // Initialize status pins as inputs
  pinMode(STATUS_NERROR_PIN, INPUT);
  pinMode(STATUS_SELECT_PIN, INPUT);
  pinMode(STATUS_PAPEROUT_PIN, INPUT);
  pinMode(STATUS_NACK_PIN, INPUT);
  pinMode(STATUS_BUSY_PIN, INPUT);

  // Send startup message
  delay(100); // Let serial stabilize
  Serial.print("PMP300 USB-Parallel Bridge v");
  Serial.print(FW_VERSION_MAJOR);
  Serial.print(".");
  Serial.print(FW_VERSION_MINOR);
  Serial.print(".");
  Serial.println(FW_VERSION_PATCH);
  Serial.print("Board: ");
  Serial.println(BOARD_TYPE);
  Serial.println("Ready.");
}

// ============================================================================
// MAIN LOOP - Process incoming commands
// ============================================================================

void loop() {
  if (Serial.available() > 0) {
    uint8_t cmd = Serial.read();

    switch(cmd) {
      case CMD_WRITE_DATA:
        handleWriteData();
        break;

      case CMD_WRITE_CTRL:
        handleWriteControl();
        break;

      case CMD_READ_STATUS:
        handleReadStatus();
        break;

      case CMD_DELAY_US:
        handleDelayMicroseconds();
        break;

      case CMD_DELAY_MS:
        handleDelayMilliseconds();
        break;

      case CMD_PING:
        handlePing();
        break;

      case CMD_VERSION:
        handleVersion();
        break;

      case CMD_SET_DATA_DIR:
        handleSetDataDirection();
        break;

      default:
        sendError(ERR_UNKNOWN_CMD);
        break;
    }
  }
}

// ============================================================================
// COMMAND HANDLERS
// ============================================================================

/*
 * CMD_WRITE_DATA ('W')
 * Write a byte to the data register (8 data pins)
 *
 * Protocol: 'W' <byte>
 * Response: 'K'
 *
 * Automatically sets data pins to OUTPUT mode if needed.
 */
void handleWriteData() {
  uint8_t value = waitForByte();

  // Ensure data pins are outputs
  if (dataDirection != OUTPUT) {
    setDataDirection(OUTPUT);
  }

  // Write each bit to corresponding pin
  for (uint8_t i = 0; i < 8; i++) {
    digitalWrite(dataPins[i], (value >> i) & 0x01);
  }

  sendOK();
}

/*
 * CMD_WRITE_CTRL ('C')
 * Write to control register (bits 2 and 3 only)
 *
 * Protocol: 'C' <byte>
 * Response: 'K'
 *
 * Note: Only bits 2 (nInitialize) and 3 (nSelect-In) are used by PMP300 protocol
 */
void handleWriteControl() {
  uint8_t value = waitForByte();

  // Extract and write control bits
  // Bit 2: nInitialize
  digitalWrite(CTRL_NINIT_PIN, (value >> 2) & 0x01);

  // Bit 3: nSelect-In
  digitalWrite(CTRL_NSELECT_PIN, (value >> 3) & 0x01);

  sendOK();
}

/*
 * CMD_READ_STATUS ('R')
 * Read byte from status register (bits 3-7)
 *
 * Protocol: 'R'
 * Response: 'V' <byte>
 *
 * Returns a byte with status bits in positions 3-7:
 *   Bit 3: nError
 *   Bit 4: Select
 *   Bit 5: Paper-Out
 *   Bit 6: nAck
 *   Bit 7: Busy
 * Bits 0-2 are always 0.
 */
void handleReadStatus() {
  uint8_t result = 0;

  // Read status pins and construct byte
  // Note: Status bits start at bit position 3
  if (digitalRead(STATUS_NERROR_PIN))   result |= 0x08; // Bit 3
  if (digitalRead(STATUS_SELECT_PIN))   result |= 0x10; // Bit 4
  if (digitalRead(STATUS_PAPEROUT_PIN)) result |= 0x20; // Bit 5
  if (digitalRead(STATUS_NACK_PIN))     result |= 0x40; // Bit 6
  if (digitalRead(STATUS_BUSY_PIN))     result |= 0x80; // Bit 7

  sendValue(result);
}

/*
 * CMD_DELAY_US ('D')
 * Delay for specified microseconds (16-bit value)
 *
 * Protocol: 'D' <high_byte> <low_byte>
 * Response: 'K'
 *
 * Max delay: 65535 microseconds (~65ms)
 */
void handleDelayMicroseconds() {
  uint8_t highByte = waitForByte();
  uint8_t lowByte = waitForByte();
  uint16_t microseconds = (highByte << 8) | lowByte;

  delayMicroseconds(microseconds);

  sendOK();
}

/*
 * CMD_DELAY_MS ('M')
 * Delay for specified milliseconds (16-bit value)
 *
 * Protocol: 'M' <high_byte> <low_byte>
 * Response: 'K'
 *
 * Max delay: 65535 milliseconds (~65 seconds)
 */
void handleDelayMilliseconds() {
  uint8_t highByte = waitForByte();
  uint8_t lowByte = waitForByte();
  uint16_t milliseconds = (highByte << 8) | lowByte;

  delay(milliseconds);

  sendOK();
}

/*
 * CMD_PING ('P')
 * Connection test - responds immediately
 *
 * Protocol: 'P'
 * Response: 'P'
 */
void handlePing() {
  Serial.write(RESP_PONG);
}

/*
 * CMD_VERSION ('V')
 * Get firmware version information
 *
 * Protocol: 'V'
 * Response: 'I' <major> <minor> <patch>
 */
void handleVersion() {
  Serial.write(RESP_VERSION);
  Serial.write(FW_VERSION_MAJOR);
  Serial.write(FW_VERSION_MINOR);
  Serial.write(FW_VERSION_PATCH);
}

/*
 * CMD_SET_DATA_DIR ('S')
 * Set data pin direction (for reading data back from device)
 *
 * Protocol: 'S' <direction>
 *   direction: 'I' = INPUT, 'O' = OUTPUT
 * Response: 'K'
 *
 * This is needed because the data pins are bidirectional in the parallel port.
 * Normally they're outputs (sending data to PMP300), but when reading data
 * back they need to be inputs.
 */
void handleSetDataDirection() {
  uint8_t dir = waitForByte();

  if (dir == 'I') {
    setDataDirection(INPUT);
    sendOK();
  } else if (dir == 'O') {
    setDataDirection(OUTPUT);
    sendOK();
  } else {
    sendError(ERR_INVALID_PARAM);
  }
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/*
 * Set direction of all data pins
 */
void setDataDirection(uint8_t dir) {
  for (uint8_t i = 0; i < 8; i++) {
    pinMode(dataPins[i], dir);
    if (dir == OUTPUT) {
      digitalWrite(dataPins[i], LOW); // Initialize to low
    }
  }
  dataDirection = dir;
}

/*
 * Wait for a byte to arrive on serial with timeout
 * Returns the byte, or 0 on timeout (with error sent)
 */
uint8_t waitForByte() {
  unsigned long startTime = millis();
  const unsigned long timeout = 1000; // 1 second timeout

  while (!Serial.available()) {
    if (millis() - startTime > timeout) {
      sendError(ERR_TIMEOUT);
      return 0;
    }
  }

  return Serial.read();
}

/*
 * Send OK response
 */
void sendOK() {
  Serial.write(RESP_OK);
}

/*
 * Send value response
 */
void sendValue(uint8_t value) {
  Serial.write(RESP_VALUE);
  Serial.write(value);
}

/*
 * Send error response
 */
void sendError(uint8_t errorCode) {
  Serial.write(RESP_ERROR);
  Serial.write(errorCode);
}

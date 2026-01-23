/*
 * PMP300 USB-to-Parallel Bridge - Optimized Version
 *
 * This Arduino sketch implements a USB-to-5V-parallel-port bridge for interfacing
 * with the Diamond Rio PMP300 MP3 player on modern computers without parallel ports.
 *
 * Hardware: Arduino Uno (or Mega 2560)
 * Interface: USB Serial at 115200 baud
 *
 * License: MIT
 */

// ============================================================================
// BOARD CONFIGURATION
// ============================================================================

#if defined(__AVR_ATmega2560__)
  #define BOARD_TYPE "Mega2560"
  #define DATA0_PIN  22
  #define DATA1_PIN  23
  #define DATA2_PIN  24
  #define DATA3_PIN  25
  #define DATA4_PIN  26
  #define DATA5_PIN  27
  #define DATA6_PIN  28
  #define DATA7_PIN  29
  #define CTRL_STROBE_PIN     30
  #define CTRL_AUTOFEED_PIN   31
  #define CTRL_NINIT_PIN      32
  #define CTRL_NSELECT_PIN    33
  #define STATUS_NERROR_PIN   34
  #define STATUS_SELECT_PIN   35
  #define STATUS_PAPEROUT_PIN 36
  #define STATUS_NACK_PIN     37
  #define STATUS_BUSY_PIN     38
#else
  #define BOARD_TYPE "Uno"
  #define DATA0_PIN  2
  #define DATA1_PIN  3
  #define DATA2_PIN  4
  #define DATA3_PIN  5
  #define DATA4_PIN  6
  #define DATA5_PIN  7
  #define DATA6_PIN  8
  #define DATA7_PIN  9
  #define CTRL_STROBE_PIN     A3
  #define CTRL_AUTOFEED_PIN   A4
  #define CTRL_NINIT_PIN      10
  #define CTRL_NSELECT_PIN    11
  #define STATUS_NERROR_PIN   12
  #define STATUS_SELECT_PIN   13
  #define STATUS_PAPEROUT_PIN A0
  #define STATUS_NACK_PIN     A1
  #define STATUS_BUSY_PIN     A2
#endif

// ============================================================================
// PROTOCOL CONSTANTS
// ============================================================================

#define SERIAL_BAUD_RATE 115200

// Commands (Host -> Arduino)
#define CMD_PING             'P'  // Connection test
#define CMD_VERSION          'V'  // Get version
#define CMD_WRITE_DATA       'W'  // Write to data register
#define CMD_WRITE_CTRL       'C'  // Write to control register
#define CMD_READ_STATUS      'R'  // Read status register
#define CMD_DELAY_MS         'M'  // Delay milliseconds
#define CMD_COMMANDOUT       'c'  // COMMANDOUT(data, ctrl1, ctrl2) - optimized
#define CMD_READ_NIBBLE_BLK  'n'  // Read bytes using nibble protocol
#define CMD_WRITE_PMP_CHUNK  'w'  // Write 528 bytes with PMP300 control toggling

// Responses (Arduino -> Host)
#define RESP_OK      'K'
#define RESP_VALUE   'V'
#define RESP_ERROR   'E'
#define RESP_PONG    'P'
#define RESP_VERSION 'I'

// Error codes
#define ERR_UNKNOWN_CMD   0x01
#define ERR_TIMEOUT       0x02

// Firmware version
#define FW_VERSION_MAJOR  2
#define FW_VERSION_MINOR  0
#define FW_VERSION_PATCH  1

// ============================================================================
// PIN ARRAYS
// ============================================================================

const uint8_t dataPins[8] = {
  DATA0_PIN, DATA1_PIN, DATA2_PIN, DATA3_PIN,
  DATA4_PIN, DATA5_PIN, DATA6_PIN, DATA7_PIN
};

// ============================================================================
// GLOBAL STATE
// ============================================================================

bool dataIsOutput = true;

// ============================================================================
// SETUP
// ============================================================================

void setup() {
  Serial.begin(SERIAL_BAUD_RATE);
  while (!Serial && millis() < 3000);

  // Data pins as outputs
  setDataOutput();

  // Control pins as outputs, idle state (0x04)
  pinMode(CTRL_STROBE_PIN, OUTPUT);
  pinMode(CTRL_AUTOFEED_PIN, OUTPUT);
  pinMode(CTRL_NINIT_PIN, OUTPUT);
  pinMode(CTRL_NSELECT_PIN, OUTPUT);
  writeControl(0x04);

  // Status pins as inputs
  pinMode(STATUS_NERROR_PIN, INPUT);
  pinMode(STATUS_SELECT_PIN, INPUT);
  pinMode(STATUS_PAPEROUT_PIN, INPUT);
  pinMode(STATUS_NACK_PIN, INPUT);
  pinMode(STATUS_BUSY_PIN, INPUT);

  delay(100);
  Serial.print(F("PMP300 Bridge v"));
  Serial.print(FW_VERSION_MAJOR);
  Serial.print('.');
  Serial.print(FW_VERSION_MINOR);
  Serial.print('.');
  Serial.print(FW_VERSION_PATCH);
  Serial.print(F(" ("));
  Serial.print(F(BOARD_TYPE));
  Serial.println(F(") Ready"));
}

// ============================================================================
// MAIN LOOP
// ============================================================================

void loop() {
  if (Serial.available()) {
    uint8_t cmd = Serial.read();

    switch(cmd) {
      case CMD_PING:           handlePing(); break;
      case CMD_VERSION:        handleVersion(); break;
      case CMD_WRITE_DATA:     handleWriteData(); break;
      case CMD_WRITE_CTRL:     handleWriteControl(); break;
      case CMD_READ_STATUS:    handleReadStatus(); break;
      case CMD_DELAY_MS:       handleDelayMs(); break;
      case CMD_COMMANDOUT:     handleCommandOut(); break;
      case CMD_READ_NIBBLE_BLK: handleReadNibbleBlock(); break;
      case CMD_WRITE_PMP_CHUNK: handleWritePMPChunk(); break;
      default:                 sendError(ERR_UNKNOWN_CMD); break;
    }
  }
}

// ============================================================================
// COMMAND HANDLERS
// ============================================================================

void handlePing() {
  Serial.write(RESP_PONG);
}

void handleVersion() {
  Serial.write(RESP_VERSION);
  Serial.write(FW_VERSION_MAJOR);
  Serial.write(FW_VERSION_MINOR);
  Serial.write(FW_VERSION_PATCH);
}

// Write byte to data register
// Protocol: 'W' <byte> -> 'K'
void handleWriteData() {
  uint8_t value = waitForByte();
  if (!dataIsOutput) setDataOutput();
  writeDataByte(value);
  Serial.write(RESP_OK);
}

// Write byte to control register
// Protocol: 'C' <byte> -> 'K'
void handleWriteControl() {
  uint8_t value = waitForByte();
  writeControl(value);
  Serial.write(RESP_OK);
}

// Read status register
// Protocol: 'R' -> 'V' <byte>
void handleReadStatus() {
  Serial.write(RESP_VALUE);
  Serial.write(readStatusByte());
}

// Delay milliseconds
// Protocol: 'M' <high> <low> -> 'K'
void handleDelayMs() {
  uint16_t ms = (waitForByte() << 8) | waitForByte();
  delay(ms);
  Serial.write(RESP_OK);
}

// Optimized COMMANDOUT - executes data, ctrl1, ctrl2 in one call
// Protocol: 'c' <data> <ctrl1> <ctrl2> -> 'K'
// This replaces 3 USB round-trips with 1
void handleCommandOut() {
  uint8_t data = waitForByte();
  uint8_t ctrl1 = waitForByte();
  uint8_t ctrl2 = waitForByte();

  if (!dataIsOutput) setDataOutput();

  writeDataByte(data);
  writeControl(ctrl1);
  writeControl(ctrl2);

  Serial.write(RESP_OK);
}

// Read multiple bytes using PMP300 nibble protocol
// Protocol: 'n' <count_high> <count_low> -> 'K' <data...>
void handleReadNibbleBlock() {
  uint16_t count = (waitForByte() << 8) | waitForByte();

  writeControl(0x04);  // Initial state
  Serial.write(RESP_OK);

  for (uint16_t i = 0; i < count; i++) {
    Serial.write(readNibbleByte());
  }
}

// Write 528 bytes (512 data + 16 end block) with PMP300 control toggling
// Protocol: 'w' <528 bytes> -> 'K'
// Control alternates: even bytes ctrl=0x00, odd bytes ctrl=0x04
void handleWritePMPChunk() {
  if (!dataIsOutput) setDataOutput();

  // Read and write 528 bytes with control toggling
  for (uint16_t i = 0; i < 528; i++) {
    uint8_t value = waitForByte();
    writeDataByte(value);

    // Alternate control: 0x00 for even, 0x04 for odd
    if ((i & 1) == 0) {
      writeControl(0x00);
    } else {
      writeControl(0x04);
    }
  }

  Serial.write(RESP_OK);
}

// ============================================================================
// LOW-LEVEL HELPERS
// ============================================================================

// Write byte to data pins (fast, no function call overhead)
inline void writeDataByte(uint8_t value) {
  digitalWrite(DATA0_PIN, (value >> 0) & 1);
  digitalWrite(DATA1_PIN, (value >> 1) & 1);
  digitalWrite(DATA2_PIN, (value >> 2) & 1);
  digitalWrite(DATA3_PIN, (value >> 3) & 1);
  digitalWrite(DATA4_PIN, (value >> 4) & 1);
  digitalWrite(DATA5_PIN, (value >> 5) & 1);
  digitalWrite(DATA6_PIN, (value >> 6) & 1);
  digitalWrite(DATA7_PIN, (value >> 7) & 1);
}

// Write control register (matches PC parallel port hardware inversion)
inline void writeControl(uint8_t value) {
  digitalWrite(CTRL_STROBE_PIN, ((value >> 0) & 1) ? LOW : HIGH);   // Inverted
  digitalWrite(CTRL_AUTOFEED_PIN, ((value >> 1) & 1) ? LOW : HIGH); // Inverted
  digitalWrite(CTRL_NINIT_PIN, ((value >> 2) & 1) ? HIGH : LOW);    // NOT inverted
  digitalWrite(CTRL_NSELECT_PIN, ((value >> 3) & 1) ? LOW : HIGH);  // Inverted
}

// Read status register
inline uint8_t readStatusByte() {
  uint8_t status = 0;
  if (digitalRead(STATUS_NERROR_PIN))   status |= 0x08;
  if (digitalRead(STATUS_SELECT_PIN))   status |= 0x10;
  if (digitalRead(STATUS_PAPEROUT_PIN)) status |= 0x20;
  if (digitalRead(STATUS_NACK_PIN))     status |= 0x40;
  if (digitalRead(STATUS_BUSY_PIN))     status |= 0x80;
  return status;
}

// Read one byte using PMP300 nibble protocol
// NOTE: PC parallel port hardware inverts the Busy line (bit 7), so C++ code
// XORs with 0x80 to compensate. Arduino reads raw signals with NO hardware
// inversion, so we must NOT XOR with 0x80 here!
uint8_t readNibbleByte() {
  uint8_t result, status;

  // High nibble: control = 0x00
  writeControl(0x00);
  delayMicroseconds(2);
  status = readStatusByte();
  result = (status & 0xF0) >> 4;  // No XOR 0x80 - Arduino has no Busy inversion

  // Low nibble: control = 0x04
  writeControl(0x04);
  delayMicroseconds(2);
  status = readStatusByte();
  result |= (status & 0xF0);  // No XOR 0x80

  // Reverse bits
  result = (result & 0xF0) >> 4 | (result & 0x0F) << 4;
  result = (result & 0xCC) >> 2 | (result & 0x33) << 2;
  result = (result & 0xAA) >> 1 | (result & 0x55) << 1;

  return result;
}

// Set data pins as outputs
void setDataOutput() {
  for (uint8_t i = 0; i < 8; i++) {
    pinMode(dataPins[i], OUTPUT);
    digitalWrite(dataPins[i], LOW);
  }
  dataIsOutput = true;
}

// Set data pins as inputs
void setDataInput() {
  for (uint8_t i = 0; i < 8; i++) {
    pinMode(dataPins[i], INPUT);
  }
  dataIsOutput = false;
}

// Wait for serial byte with timeout
uint8_t waitForByte() {
  unsigned long start = millis();
  while (!Serial.available()) {
    if (millis() - start > 1000) {
      sendError(ERR_TIMEOUT);
      return 0;
    }
  }
  return Serial.read();
}

void sendError(uint8_t code) {
  Serial.write(RESP_ERROR);
  Serial.write(code);
}

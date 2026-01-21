/*
 * PMP300 Hardware Test Sketch
 *
 * This simple sketch helps you verify your wiring is correct before
 * using the full USB-Parallel bridge firmware.
 *
 * Upload this sketch, open Serial Monitor at 115200 baud, and follow
 * the on-screen instructions.
 *
 * Tests:
 * 1. Data pins output (D0-D7)
 * 2. Control pins output (nInitialize, nSelect-In)
 * 3. Status pins input (nError, Select, Paper-Out, nAck, Busy)
 *
 * Use a multimeter or logic analyzer to verify pin voltages.
 */

// ============================================================================
// PIN DEFINITIONS (auto-detect board type)
// ============================================================================

#if defined(__AVR_ATmega2560__)
  // Arduino Mega 2560
  const byte dataPins[8] = {22, 23, 24, 25, 26, 27, 28, 29};
  const byte ctrlPins[2] = {30, 31};  // nInit, nSelect
  const byte statusPins[5] = {32, 33, 34, 35, 36};  // nError, Select, PaperOut, nAck, Busy
  const char* boardName = "Arduino Mega 2560";
#else
  // Arduino Uno
  const byte dataPins[8] = {2, 3, 4, 5, 6, 7, 8, 9};
  const byte ctrlPins[2] = {10, 11};  // nInit, nSelect
  const byte statusPins[5] = {12, 13, A0, A1, A2};  // nError, Select, PaperOut, nAck, Busy
  const char* boardName = "Arduino Uno";
#endif

// ============================================================================
// SETUP
// ============================================================================

void setup() {
  Serial.begin(115200);
  while (!Serial && millis() < 3000) {
    ; // Wait for serial port
  }

  Serial.println("========================================");
  Serial.println("PMP300 Hardware Test");
  Serial.println("========================================");
  Serial.print("Board detected: ");
  Serial.println(boardName);
  Serial.println();

  // Configure data pins as outputs
  for (int i = 0; i < 8; i++) {
    pinMode(dataPins[i], OUTPUT);
    digitalWrite(dataPins[i], LOW);
  }

  // Configure control pins as outputs
  for (int i = 0; i < 2; i++) {
    pinMode(ctrlPins[i], OUTPUT);
    digitalWrite(ctrlPins[i], LOW);
  }

  // Configure status pins as inputs
  for (int i = 0; i < 5; i++) {
    pinMode(statusPins[i], INPUT);
  }

  Serial.println("Pin configuration complete.");
  Serial.println();
  printMenu();
}

// ============================================================================
// MAIN LOOP
// ============================================================================

void loop() {
  if (Serial.available()) {
    char cmd = Serial.read();

    switch (cmd) {
      case '1':
        testDataPins();
        break;
      case '2':
        testControlPins();
        break;
      case '3':
        testStatusPins();
        break;
      case '4':
        testWalking1();
        break;
      case '5':
        testAllOutputs();
        break;
      case 'm':
      case 'M':
        printMenu();
        break;
      default:
        // Ignore whitespace/newlines
        if (cmd != '\n' && cmd != '\r') {
          Serial.print("Unknown command: ");
          Serial.println(cmd);
        }
        break;
    }
  }
}

// ============================================================================
// TEST FUNCTIONS
// ============================================================================

void printMenu() {
  Serial.println("Select a test:");
  Serial.println("  1 - Test data pins (output patterns)");
  Serial.println("  2 - Test control pins (toggle)");
  Serial.println("  3 - Test status pins (read and display)");
  Serial.println("  4 - Walking 1's test (data pins)");
  Serial.println("  5 - Test all outputs HIGH");
  Serial.println("  M - Show this menu");
  Serial.println();
}

// Test data pins with various patterns
void testDataPins() {
  Serial.println("--- Data Pins Test ---");
  Serial.println("Testing patterns on data pins D0-D7");
  Serial.print("Arduino pins: ");
  for (int i = 0; i < 8; i++) {
    Serial.print(dataPins[i]);
    if (i < 7) Serial.print(", ");
  }
  Serial.println();
  Serial.println();

  // Test pattern 1: All LOW
  Serial.println("Pattern 1: All LOW (0x00)");
  writeDataByte(0x00);
  delay(1000);

  // Test pattern 2: All HIGH
  Serial.println("Pattern 2: All HIGH (0xFF)");
  writeDataByte(0xFF);
  delay(1000);

  // Test pattern 3: Alternating 0xAA (10101010)
  Serial.println("Pattern 3: Alternating 0xAA (10101010)");
  writeDataByte(0xAA);
  delay(1000);

  // Test pattern 4: Alternating 0x55 (01010101)
  Serial.println("Pattern 4: Alternating 0x55 (01010101)");
  writeDataByte(0x55);
  delay(1000);

  // Test pattern 5: Count 0-255
  Serial.println("Pattern 5: Counting 0x00 to 0xFF");
  for (int i = 0; i <= 255; i++) {
    writeDataByte(i);
    delay(10);
  }

  // Reset to LOW
  writeDataByte(0x00);
  Serial.println("Data pins reset to LOW");
  Serial.println();
}

// Test control pins
void testControlPins() {
  Serial.println("--- Control Pins Test ---");
  Serial.print("nInitialize pin: ");
  Serial.println(ctrlPins[0]);
  Serial.print("nSelect-In pin: ");
  Serial.println(ctrlPins[1]);
  Serial.println();

  // Test each control pin
  Serial.println("Testing nInitialize (bit 2):");
  Serial.println("  LOW -> HIGH -> LOW");
  digitalWrite(ctrlPins[0], LOW);
  delay(500);
  digitalWrite(ctrlPins[0], HIGH);
  delay(500);
  digitalWrite(ctrlPins[0], LOW);
  delay(500);

  Serial.println("Testing nSelect-In (bit 3):");
  Serial.println("  LOW -> HIGH -> LOW");
  digitalWrite(ctrlPins[1], LOW);
  delay(500);
  digitalWrite(ctrlPins[1], HIGH);
  delay(500);
  digitalWrite(ctrlPins[1], LOW);
  delay(500);

  // Test combinations (control register values used in protocol)
  Serial.println("Testing control combinations:");
  Serial.println("  0x00 (both LOW)");
  writeControlByte(0x00);
  delay(1000);

  Serial.println("  0x04 (bit 2 HIGH)");
  writeControlByte(0x04);
  delay(1000);

  Serial.println("  0x08 (bit 3 HIGH)");
  writeControlByte(0x08);
  delay(1000);

  Serial.println("  0x0C (both HIGH)");
  writeControlByte(0x0C);
  delay(1000);

  // Reset to 0x04 (common idle state)
  writeControlByte(0x04);
  Serial.println("Control pins reset to 0x04");
  Serial.println();
}

// Test status pins
void testStatusPins() {
  Serial.println("--- Status Pins Test ---");
  Serial.println("Pin mapping:");
  Serial.print("  nError (bit 3): ");
  Serial.println(statusPins[0]);
  Serial.print("  Select (bit 4): ");
  Serial.println(statusPins[1]);
  Serial.print("  Paper-Out (bit 5): ");
  Serial.println(statusPins[2]);
  Serial.print("  nAck (bit 6): ");
  Serial.println(statusPins[3]);
  Serial.print("  Busy (bit 7): ");
  Serial.println(statusPins[4]);
  Serial.println();

  Serial.println("Reading status pins continuously...");
  Serial.println("(Press any key to stop)");
  Serial.println();
  Serial.println("Status | nErr Sel POut nAck Busy | Hex");
  Serial.println("-------+---------------------------+-----");

  while (!Serial.available()) {
    byte status = readStatusByte();

    // Print as binary
    for (int i = 7; i >= 0; i--) {
      if (i == 7 || i == 2) Serial.print(" ");
      Serial.print((status >> i) & 1);
    }

    // Print individual bits
    Serial.print(" | ");
    Serial.print((status >> 3) & 1);  // nError
    Serial.print("    ");
    Serial.print((status >> 4) & 1);  // Select
    Serial.print("   ");
    Serial.print((status >> 5) & 1);  // Paper-Out
    Serial.print("    ");
    Serial.print((status >> 6) & 1);  // nAck
    Serial.print("    ");
    Serial.print((status >> 7) & 1);  // Busy

    // Print as hex
    Serial.print(" | 0x");
    if (status < 0x10) Serial.print("0");
    Serial.println(status, HEX);

    delay(100);
  }

  // Clear input
  while (Serial.available()) Serial.read();

  Serial.println();
  Serial.println("Status pin test stopped.");
  Serial.println();
}

// Walking 1's test - helps identify individual pin connections
void testWalking1() {
  Serial.println("--- Walking 1's Test ---");
  Serial.println("Single bit HIGH walks across data pins D0-D7");
  Serial.println("Watch with multimeter or LED to verify each pin");
  Serial.println();

  for (int i = 0; i < 8; i++) {
    byte pattern = 1 << i;
    Serial.print("Bit ");
    Serial.print(i);
    Serial.print(" HIGH (0x");
    if (pattern < 0x10) Serial.print("0");
    Serial.print(pattern, HEX);
    Serial.print(") - Pin ");
    Serial.println(dataPins[i]);

    writeDataByte(pattern);
    delay(1000);
  }

  writeDataByte(0x00);
  Serial.println("Walking 1's test complete.");
  Serial.println();
}

// Test all outputs HIGH
void testAllOutputs() {
  Serial.println("--- All Outputs HIGH Test ---");
  Serial.println("Setting all output pins HIGH for 5 seconds");
  Serial.println("Use multimeter to verify 5V on all pins");
  Serial.println();

  writeDataByte(0xFF);
  writeControlByte(0x0C);

  Serial.println("Data pins: 0xFF (all HIGH)");
  Serial.println("Control pins: 0x0C (both HIGH)");
  Serial.println();

  for (int i = 5; i > 0; i--) {
    Serial.print(i);
    Serial.println(" seconds...");
    delay(1000);
  }

  writeDataByte(0x00);
  writeControlByte(0x04);

  Serial.println("All outputs reset.");
  Serial.println();
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

void writeDataByte(byte value) {
  for (int i = 0; i < 8; i++) {
    digitalWrite(dataPins[i], (value >> i) & 1);
  }
}

void writeControlByte(byte value) {
  digitalWrite(ctrlPins[0], (value >> 2) & 1);  // Bit 2
  digitalWrite(ctrlPins[1], (value >> 3) & 1);  // Bit 3
}

byte readStatusByte() {
  byte result = 0;
  if (digitalRead(statusPins[0])) result |= 0x08;  // Bit 3: nError
  if (digitalRead(statusPins[1])) result |= 0x10;  // Bit 4: Select
  if (digitalRead(statusPins[2])) result |= 0x20;  // Bit 5: Paper-Out
  if (digitalRead(statusPins[3])) result |= 0x40;  // Bit 6: nAck
  if (digitalRead(statusPins[4])) result |= 0x80;  // Bit 7: Busy
  return result;
}

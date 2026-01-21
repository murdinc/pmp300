# PMP300 Arduino Wiring Guide

Visual reference for connecting Arduino to DB-25 female connector.

## DB-25 Female Connector Pinout

Looking at the **solder side** of DB-25 female connector (where you'll solder wires):

```
 13  12  11  10   9   8   7   6   5   4   3   2   1
┌───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┐
│ • │ • │ • │ • │ • │ • │ • │ • │ • │ • │ • │ • │ • │
└─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┘
  │ • │ • │ • │ • │ • │ • │ • │ • │ • │ • │ • │ • │
  └───┴───┴───┴───┴───┴───┴───┴───┴───┴───┴───┴───┘
   25  24  23  22  21  20  19  18  17  16  15  14
```

## Arduino Mega 2560 Wiring

### Data Lines (8 wires)

```
Arduino Pin    →    DB-25 Pin    Signal Name
─────────────────────────────────────────────
    22        →        2         Data 0 (LSB)
    23        →        3         Data 1
    24        →        4         Data 2
    25        →        5         Data 3
    26        →        6         Data 4
    27        →        7         Data 5
    28        →        8         Data 6
    29        →        9         Data 7 (MSB)
```

### Control Lines (2 wires)

```
Arduino Pin    →    DB-25 Pin    Signal Name
─────────────────────────────────────────────
    30        →       16         nInitialize
    31        →       17         nSelect-In
```

### Status Lines (5 wires)

```
Arduino Pin    ←    DB-25 Pin    Signal Name
─────────────────────────────────────────────
    32        ←       15         nError
    33        ←       13         Select
    34        ←       12         Paper-Out
    35        ←       10         nAck
    36        ←       11         Busy
```

### Ground (1 wire minimum, more is better)

```
Arduino Pin    →    DB-25 Pin    Signal Name
─────────────────────────────────────────────
    GND       →    18-25        Ground (connect all)
```

**Total: 16 connections (15 signals + ground)**

---

## Arduino Uno Wiring

### Data Lines (8 wires)

```
Arduino Pin    →    DB-25 Pin    Signal Name
─────────────────────────────────────────────
     2        →        2         Data 0 (LSB)
     3        →        3         Data 1
     4        →        4         Data 2
     5        →        5         Data 3
     6        →        6         Data 4
     7        →        7         Data 5
     8        →        8         Data 6
     9        →        9         Data 7 (MSB)
```

### Control Lines (2 wires)

```
Arduino Pin    →    DB-25 Pin    Signal Name
─────────────────────────────────────────────
    10        →       16         nInitialize
    11        →       17         nSelect-In
```

### Status Lines (5 wires)

```
Arduino Pin    ←    DB-25 Pin    Signal Name
─────────────────────────────────────────────
    12        ←       15         nError
    13        ←       13         Select
    A0        ←       12         Paper-Out
    A1        ←       10         nAck
    A2        ←       11         Busy
```

### Ground (1 wire minimum, more is better)

```
Arduino Pin    →    DB-25 Pin    Signal Name
─────────────────────────────────────────────
    GND       →    18-25        Ground (connect all)
```

**Total: 16 connections (15 signals + ground)**

---

## Visual Wiring Diagram (Arduino Mega)

```
┌─────────────────────────────┐
│   Arduino Mega 2560         │
│                             │
│  [22] ────────────┐         │
│  [23] ──────────┐ │         │
│  [24] ────────┐ │ │         │
│  [25] ──────┐ │ │ │         │
│  [26] ────┐ │ │ │ │         │     ┌──────────────────┐
│  [27] ──┐ │ │ │ │ │         │     │   DB-25 Female   │
│  [28] ┐ │ │ │ │ │ │         │     │   (Solder Side)  │
│  [29] │ │ │ │ │ │ │         │     │                  │
│       │ │ │ │ │ │ │         │     │  [2] Data 0      │◄────┘ │ │ │ │ │ │ │
│  [30] │ │ │ │ │ │ │ ────────┼─────┤  [3] Data 1      │◄──────┘ │ │ │ │ │ │
│  [31] │ │ │ │ │ │ │ ────────┼─────┤  [4] Data 2      │◄────────┘ │ │ │ │ │
│       │ │ │ │ │ │ │         │     │  [5] Data 3      │◄──────────┘ │ │ │ │
│  [32] │ │ │ │ │ │ │ ────────┼─────┤  [6] Data 4      │◄────────────┘ │ │ │
│  [33] │ │ │ │ │ │ │ ────────┼─────┤  [7] Data 5      │◄──────────────┘ │ │
│  [34] │ │ │ │ │ │ │ ────────┼─────┤  [8] Data 6      │◄────────────────┘ │
│  [35] │ │ │ │ │ │ │ ────────┼─────┤  [9] Data 7      │◄──────────────────┘
│  [36] │ │ │ │ │ │ │ ────────┼─────┤                  │
│       │ │ │ │ │ │ │         │     │  [10] nAck       │─────────────────┐
│  [GND]│ │ │ │ │ │ └─────────┼─────┤  [11] Busy       │───────────────┐ │
│       │ │ │ │ │ └───────────┼─────┤  [12] Paper-Out  │─────────────┐ │ │
│       │ │ │ │ └─────────────┼─────┤  [13] Select     │───────────┐ │ │ │
│       │ │ │ └───────────────┼─────┤  [15] nError     │─────────┐ │ │ │ │
│       │ │ └─────────────────┼─────┤  [16] nInitialize│───────┐ │ │ │ │ │
│       │ └───────────────────┼─────┤  [17] nSelect-In │─────┐ │ │ │ │ │ │
│       └─────────────────────┼─────┤                  │     │ │ │ │ │ │ │
│                             │     │  [18-25] GND     │◄──┐ │ │ │ │ │ │ │
└─────────────────────────────┘     └──────────────────┘   │ │ │ │ │ │ │ │
                                                           │ │ │ │ │ │ │ │
                                    Status Inputs ─────────┴─┴─┴─┴─┘ │ │ │
                                    Control Outputs ─────────────────┴─┴─┘
```

---

## Step-by-Step Assembly

### Method 1: Soldering (Permanent)

1. **Prepare wires**
   - Cut 16 wires ~6-12 inches long
   - Strip 1/4" from each end
   - Use different colored wires if possible
   - Label each wire with tape

2. **Solder to DB-25**
   - Tin the wire ends
   - Tin the DB-25 pins
   - Solder wires to appropriate DB-25 pins
   - Let cool completely
   - Test continuity with multimeter

3. **Connect to Arduino**
   - If using headers: crimp or solder Dupont connectors
   - If soldering: solder directly to Arduino pins
   - Test continuity again

4. **Test before connecting PMP300**
   - Upload hardware_test sketch
   - Verify all pins with multimeter
   - Check for shorts between adjacent pins

### Method 2: DB-25 Breakout Board (Easy)

1. **Get DB-25 breakout board** (~$5-10 online)
   - Female DB-25 connector
   - Screw terminals or header pins
   - Makes wiring much easier

2. **Connect with jumper wires**
   - Use male-to-female jumper wires
   - Connect according to tables above
   - Label each connection

3. **Test thoroughly**
   - Upload hardware_test sketch
   - Verify with multimeter
   - Check all connections

### Method 3: Breadboard Prototyping (Testing)

1. **Use DB-25 breakout with headers**
2. **Plug into breadboard**
3. **Use jumper wires to Arduino**
4. **Test and verify**
5. **Move to permanent solution once working**

---

## Wire Color Code Suggestion

Use consistent colors to avoid mistakes:

```
Data Lines:     White, Gray, Purple, Blue, Green, Yellow, Orange, Red
                (D0 → D7, matches resistor color code)

Control Lines:  Brown (nInitialize), Black (nSelect-In)

Status Lines:   Pink (nError), Cyan (Select), Magenta (Paper-Out),
                Lime (nAck), Violet (Busy)

Ground:         Black or Green
```

Or use ribbon cable with different colored stripes.

---

## Testing Your Wiring

### Continuity Test (Before Powering On)

Use multimeter in continuity mode:

1. **Test each signal wire**
   - Touch one probe to Arduino pin
   - Touch other probe to DB-25 pin
   - Should beep/show continuity
   - Verify correct pin numbers

2. **Test for shorts**
   - Test between adjacent Arduino pins (should be open)
   - Test between adjacent DB-25 pins (should be open)
   - Test signal to ground (should be open)

3. **Test ground**
   - Verify ground continuity
   - Check all DB-25 ground pins connected

### Voltage Test (After Uploading Firmware)

Use multimeter in voltage mode:

1. **Upload hardware_test sketch**
2. **Run "All Outputs HIGH" test**
3. **Measure voltage**:
   - Data pins should read ~5V
   - Control pins should read ~5V
   - Ground should read 0V

4. **Run "All Outputs LOW" test**
5. **Measure voltage**:
   - All pins should read ~0V

---

## Common Mistakes

❌ **Wrong connector gender**: Need **female** DB-25 (has holes, not pins)
❌ **Counting pins wrong**: Verify with pinout diagram above
❌ **Missing ground**: Ground connection is critical!
❌ **Swapped data bits**: D0 is LSB (pin 2), D7 is MSB (pin 9)
❌ **Using 3.3V Arduino**: Must use 5V board (Uno, Mega, Nano)
❌ **Bad solder joints**: Test continuity after soldering
❌ **Shorts between pins**: Check carefully with multimeter

---

## Materials Checklist

- [ ] Arduino Mega 2560 or Uno
- [ ] DB-25 female connector (solder cup or breakout board)
- [ ] Wire (22-24 AWG solid or stranded)
- [ ] Wire strippers
- [ ] Soldering iron and solder (if soldering)
- [ ] Multimeter (essential for testing)
- [ ] Labels or tape (for marking wires)
- [ ] Optional: Heat shrink tubing
- [ ] Optional: DB-25 backshell/hood
- [ ] Optional: Project enclosure

---

## Next Steps After Wiring

1. **Test with multimeter** - Verify all connections
2. **Upload hardware_test sketch** - Test each pin
3. **Upload production firmware** - pmp300_usb_parallel_bridge
4. **Run example Go client** - Test communication
5. **Connect PMP300** - Finally plug in your device
6. **Build your software** - Use example as reference

---

## Resources

- Arduino Mega pinout: https://docs.arduino.cc/hardware/mega-2560
- Arduino Uno pinout: https://docs.arduino.cc/hardware/uno-rev3
- DB-25 connector info: See main project README
- Soldering guide: https://learn.adafruit.com/adafruit-guide-excellent-soldering

---

## Questions?

If you encounter issues:
1. Double-check wiring against diagrams above
2. Test continuity with multimeter
3. Verify correct board type in firmware
4. Upload hardware_test sketch for diagnostics
5. Check main project README for troubleshooting

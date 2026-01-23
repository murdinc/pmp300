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

## Resources

- Arduino Mega pinout: https://docs.arduino.cc/hardware/mega-2560
- Arduino Uno pinout: https://docs.arduino.cc/hardware/uno-rev3

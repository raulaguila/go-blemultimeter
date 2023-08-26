# Bluetooth communication protocol

This work was carried using bluetooth to connect to the multimeter and tracking the message bytes in each multimeter function.

Tested with [Owon - OW18E Digital Multimeter](https://owon.com.hk/products_owon_ow18d%7Ce_4_1%7C2_digits__handheld_digital_multimeter)

![](/screenshot/ow18e.png)

### Example message:

- Receive a 6-byte array
- Example read: [ 98, 240, 4, 0, [147, 49](#5th--6th-bytes) ]
- Converting read example into 8-digit binary: [ [01100010](#1st-byte), [11110000](#2nd-byte), [00000100](#3rd-byte), [00000000](#4th-byte), 10010011, 00110001 ]
- Final output struct on code: `<value> <unit> <flags>`
- Final output of example read: `126.91 V [AC, Auto Range]`

### 1st Byte

- 8 bits, in the example read: [01100010](#example-message)

  - Bits [0..1]: Represents the function (In the example read: _01_)
  - Bits [2..4]: Represents the unit of measure (In the example read: _100_)
  - Bits [5..7]: Represents the range of the measured value (In the example read: _010_)

    | Bits [0..1] | func |     | Bits [2..4] | unit |     | Bits [5..7] | range |
    | :---------: | :--: | --- | :---------: | :--: | --- | :---------: | :---: |
    |     00      |  DC  |     |     001     |  n   |     |     000     |  NCV  |
    |     01      |  AC  |     |     010     |  µ   |     |     001     | 2000  |
    |     10      | Diod |     |     011     |  m   |     |     010     |  200  |
    |     11      | Cont |     |     100     |  1   |     |     011     |  20   |
    |      -      |  -   |     |     101     |  k   |     |     100     |   2   |
    |      -      |  -   |     |     110     |  M   |     |     111     |   L   |

### 2nd Byte

- 8 bits, in the example read: [11110000](#example-message)

  | Bit(s) | Value |                Function                |
  | :----: | :---: | :------------------------------------: |
  | [0..3] | 1111  | Apparently they are not used [\*](#ps) |
  |        |       |                                        |
  | [4..5] |  00   | Apparently they are not used [\*](#ps) |
  |        |       |                                        |
  | [6..7] |  00   |                Voltage                 |
  | [6..7] |  01   |               Resistance               |
  | [6..7] |  10   |               Continuity               |
  | [6..7] |  11   |                  NCV                   |

### 3rd Byte

- 8 bits, in the example read: [00000100](#example-message)

  | Bit(s) | Value |                Function                |
  | :----: | :---: | :------------------------------------: |
  | [0..3] | 0000  | Apparently they are not used [\*](#ps) |
  |        |       |                                        |
  |   4    |   1   |              Low Battery               |
  |   4    |   0   |               Battery OK               |
  |        |       |                                        |
  |   5    |   1   |           Auto range enabled           |
  |   5    |   0   |          Auto range disabled           |
  |        |       |                                        |
  |   6    |   1   |         Relative mode enabled          |
  |   6    |   0   |         Relative mode disabled         |
  |        |       |                                        |
  |   7    |   1   |              Hold enabled              |
  |   7    |   0   |             Hold disabled              |

### 4th Byte

- 8 bits, in the example read: [00000000](#example-message)
- Apparently this byte is not used [\*](#ps)

### 5th & 6th Bytes

- In the example read: [\[... **147** **49**\]](#example-message)
- Represents the measurement value
- Use them without converting to binary
- 6th byte counts the overflow of 5th byte
- If the 5th byte >= 128, it is a negative value

### Final function

- Combining the 1st and 2nd byte items, it has the final function.
- [`First two bits of 1st byte`](#1st-byte) + [`last two bits of 2nd byte`](#2nd-byte).

  | [1st Byte](#1st-byte) [0..1] | [2nd Byte](#2nd-byte) [6..7] |   Final function    | Symbol |
  | :--------------------------: | :--------------------------: | :-----------------: | :----: |
  |              DC              |          Continuity          |     Temperature     |   ºC   |
  |              DC              |          Resistance          | Resistance Measure  |   Ω    |
  |              DC              |           Voltage            | DC Voltage Measure  |   V    |
  |              AC              |          Continuity          |     Temperature     |   ºF   |
  |              AC              |             NCV              |     NCV Measure     |   -    |
  |              AC              |          Resistance          | Capacitance Measure |   F    |
  |              AC              |           Voltage            | AC Voltage Measure  |   V    |
  |             Diod             |          Continuity          |     Diode test      |   V    |
  |             Diod             |          Resistance          |  Frequence Measure  |   Hz   |
  |             Diod             |           Voltage            |   Current Measure   |   A    |
  |             Cont             |          Continuity          |   Continuity test   |   Ω    |
  |             Cont             |          Resistance          |  Frequence Measure  |   %    |

#### PS

- Maybe these unused bits are used to represent information that was not tracked, such as MIN, MAX and others.
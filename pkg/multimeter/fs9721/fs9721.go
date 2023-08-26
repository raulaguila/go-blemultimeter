package fs9721

import (
	"fmt"
	"strconv"
	"strings"
)

type Fs9721 struct {
	bytearray     []string
	originalarray []byte
}

func (m *Fs9721) ProccessArray(bytearray []byte) (value float64, unit string, flags []string) {
	switch len(bytearray) {
	case 8:
		m.bytearray = m.bytearray[:0]
		m.originalarray = m.originalarray[:0]
		fallthrough
	case 6:
		for _, b := range bytearray {
			aux := fmt.Sprintf("%08b", b)
			m.bytearray = append(m.bytearray, aux[len(aux)-4:])
			m.originalarray = append(m.originalarray, b)
		}

		if len(m.bytearray) == 14 {
			str := strings.Join(m.bytearray, "")

			value = m.extractValue(str)
			unit = m.extractUnit(str)
			flags = m.extractFlags(str)
		}
	}

	return value, unit, flags
}

func (m *Fs9721) extractValue(str string) (ret float64) {
	digits := map[string]string{
		"1111101": "0",
		"0000101": "1",
		"1011011": "2",
		"0011111": "3",
		"0100111": "4",
		"0111110": "5",
		"1111110": "6",
		"0010101": "7",
		"1111111": "8",
		"0111111": "9",
		"0000000": "",
		"1101000": "L",
	}

	arrDigits := []string{
		str[5:12],  // Digito 01
		str[12:13], // Ponto 01
		str[13:20], // Digito 02
		str[20:21], // Ponto 02
		str[21:28], // Digito 03
		str[28:29], // Ponto 03
		str[29:36], // Digito 04
	}

	measured := "0"
	for i, digit := range arrDigits {
		switch i % 2 {
		case 0:
			if val, exist := digits[digit]; exist {
				measured += val
			}
		case 1:
			if digit == "1" {
				measured += "."
			}
		}
	}

	ret, _ = strconv.ParseFloat(measured, 64)
	if str[4:5] == "1" {
		ret = ret * -1
	}

	return
}

func (m *Fs9721) extractUnit(str string) (unit string) {
	arrUnits := [][2]interface{}{
		{str[37:38] == "1", "n"},  // nano
		{str[36:37] == "1", "µ"},  // micro
		{str[38:39] == "1", "k"},  // kilo
		{str[40:41] == "1", "m"},  // mili
		{str[42:43] == "1", "M"},  // mega
		{str[41:42] == "1", "%"},  // percent
		{str[45:46] == "1", "Ω"},  // ohm
		{str[48:49] == "1", "A"},  // amp
		{str[49:50] == "1", "V"},  // volts
		{str[44:45] == "1", "F"},  // cap
		{str[50:51] == "1", "Hz"}, // hertz
		{str[53:54] == "1", "°C"}, // temp
	}

	for _, item := range arrUnits {
		if item[0].(bool) {
			unit += item[1].(string)
		}
	}

	return
}

func (m *Fs9721) extractFlags(str string) (flags []string) {
	arrFlags := [][2]interface{}{
		{str[0:1] == "1", "AC"},
		{str[1:2] == "1" && str[53:54] == "0", "DC"},
		{str[2:3] == "1", "Auto"},
		{str[39:40] == "1", "Diode test"},
		{str[43:44] == "1", "Conti test"},
		{str[44:45] == "1", "Capacity"},
		{str[46:47] == "1", "Rel"},
		{str[47:48] == "1", "Hold"},
		{str[52:53] == "1", "Min"},
		{str[55:56] == "1", "Max"},
		{str[51:52] == "1", "LowBat"},
		{str[21:28] == "1101000", "L"},
	}

	for _, flag := range arrFlags {
		if flag[0].(bool) {
			flags = append(flags, flag[1].(string))
		}
	}

	return
}

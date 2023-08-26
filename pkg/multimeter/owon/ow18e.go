package owon

import (
	"fmt"
	"log"
)

const (
	// First 2 bits of 1st byte
	dc   string = "00"
	ac   string = "01"
	diod string = "10"
	cont string = "11"

	// Last 2 bits of 2nd byte
	voltage    string = "00"
	resistance string = "01"
	continuity string = "10"
	ncv        string = "11"
)

type OW18E struct{}

func (m *OW18E) getBinArray(bytearray []byte) (str []string) {
	for _, b := range bytearray {
		str = append(str, fmt.Sprintf("%08b", b))
	}
	return
}

func (m *OW18E) ProccessArray(bytearray []byte) (float64, string, []string) {
	binArray := m.getBinArray(bytearray) // Convert byte array to bits string

	mRange := binArray[0][5:]
	unity := binArray[0][2:5]
	finalFunction := binArray[0][:2] + binArray[1][6:]
	function := binArray[0][:2]

	value := m.extractValue(bytearray, mRange)
	unit := m.extractUnit(unity, finalFunction)
	flags := m.extractFlags(binArray, finalFunction, function, mRange)

	return value, unit, flags
}

func (m *OW18E) calcValue(byte5 byte, byte4 byte, div float64, negative bool) (ret float64) {
	ret = float64(byte5)
	if negative {
		ret -= 128
	}

	ret = ((ret * 256) + float64(byte4)) / div

	if negative {
		ret *= -1
	}

	return
}

func (m *OW18E) extractValue(byteArray []byte, mRange string) float64 {
	switch mRange {
	case "100": // range 2
		return m.calcValue(byteArray[5], byteArray[4], 10000, byteArray[5] >= 128)
	case "011": // range 20
		return m.calcValue(byteArray[5], byteArray[4], 1000, byteArray[5] >= 128)
	case "010": // range 200
		return m.calcValue(byteArray[5], byteArray[4], 100, byteArray[5] >= 128)
	case "001": // range 2000
		return m.calcValue(byteArray[5], byteArray[4], 10, byteArray[5] >= 128)
	case "000": // NCV
		return float64(byteArray[4])
	case "111": // L
		return 0
	default:
		log.Printf("\tRange not tracked: %v\n", mRange)
	}

	return 0
}

func (m *OW18E) extractUnit(unity string, finalFunction string) string {
	arrUnits := [][2]interface{}{
		{unity == "001", "n"},                    // nano
		{unity == "010", "µ"},                    // micro
		{unity == "011", "m"},                    // mili
		{unity == "100", ""},                     // 1
		{unity == "101", "k"},                    // kilo
		{unity == "110", "M"},                    // Mega
		{finalFunction == dc+continuity, "ºC"},   // Temp celsius
		{finalFunction == dc+voltage, "V"},       // DC Voltage Measure
		{finalFunction == dc+resistance, "Ω"},    // Resistance Measure
		{finalFunction == ac+continuity, "ºF"},   // Temp fahrenheit
		{finalFunction == ac+resistance, "F"},    // Capacitance Measure
		{finalFunction == ac+voltage, "V"},       // AC Voltage Measure
		{finalFunction == ac+ncv, "NVC"},         // NVC Measure
		{finalFunction == diod+continuity, "V"},  // Diode test
		{finalFunction == diod+resistance, "Hz"}, // Frequence
		{finalFunction == diod+voltage, "A"},     // Current Measure
		{finalFunction == cont+continuity, "Ω"},  // Continuity test
		{finalFunction == cont+resistance, "%"},  // Percentage
	}

	unit := ""
	for _, item := range arrUnits {
		if item[0].(bool) {
			unit += item[1].(string)
		}
	}

	return unit
}

func (m *OW18E) extractFlags(binArray []string, finalFunction, function, mRange string) []string {
	arrFlags := [][2]interface{}{
		{function == dc, "DC"},                                // DC Voltage Measure
		{function == ac, "AC"},                                // AC Voltage Measure
		{mRange == "111", "L"},                                // L
		{finalFunction == dc+continuity, "Temp celsius"},      // Temp celsius
		{finalFunction == ac+continuity, "Temp fahrenheit"},   // Temp fahrenheit
		{finalFunction == ac+resistance, "Capacity"},          // Capacitance Measure
		{finalFunction == ac+ncv, "NCV Measure"},              // NCV Measure
		{finalFunction == diod+continuity, "Diode test"},      // Diode test
		{finalFunction == cont+continuity, "Continuity test"}, // Continuity test
		{finalFunction == cont+resistance, "Percentage"},      // Percentage
		{binArray[2][4] == '1', "Low Battery"},                // Low Battery
		{binArray[2][5] == '1', "Auto Range"},                 // Auto Range
		{binArray[2][6] == '1', "Relative Mode"},              // Relative Mode
		{binArray[2][7] == '1', "Hold"},                       // Hold
	}

	flags := []string{}
	for _, flag := range arrFlags {
		if flag[0].(bool) {
			flags = append(flags, flag[1].(string))
		}
	}

	return flags
}

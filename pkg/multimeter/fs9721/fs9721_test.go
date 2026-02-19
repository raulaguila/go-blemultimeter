package fs9721

import (
	"math"
	"reflect"
	"strconv"
	"testing"
)

var fsDigitBits = map[string]string{
	"0": "1111101",
	"1": "0000101",
	"2": "1011011",
	"3": "0011111",
	"4": "0100111",
	"5": "0111110",
	"6": "1111110",
	"7": "0010101",
	"8": "1111111",
	"9": "0111111",
	"L": "1101000",
}

func TestFS9721ProcessArray_StatefulParsing(t *testing.T) {
	t.Parallel()

	first, second := buildFS9721Frames(t, buildFS9721Bits([4]string{"1", "2", "3", "4"}, 2, false, map[int]byte{
		1:  '1', // DC
		2:  '1', // Auto
		49: '1', // V
	}))

	m := &FS9721{}

	v1, u1, f1 := m.ProcessArray(first)
	if v1 != 0 || u1 != "" || f1 != nil {
		t.Fatalf("first 8-byte chunk should not emit value, got value=%v unit=%q flags=%v", v1, u1, f1)
	}

	v2, u2, f2 := m.ProcessArray(second)
	if math.Abs(v2-12.34) > 1e-9 {
		t.Fatalf("value mismatch: got %.12f want %.12f", v2, 12.34)
	}
	if u2 != "V" {
		t.Fatalf("unit mismatch: got %q want %q", u2, "V")
	}
	if !reflect.DeepEqual(f2, []string{"DC", "Auto"}) {
		t.Fatalf("flags mismatch: got %v want %v", f2, []string{"DC", "Auto"})
	}
}

func TestFS9721ProcessArray_ResetOnNew8ByteHeader(t *testing.T) {
	t.Parallel()

	firstA, secondA := buildFS9721Frames(t, buildFS9721Bits([4]string{"1", "2", "3", "4"}, 2, false, map[int]byte{
		1:  '1',
		49: '1',
	}))
	firstB, secondB := buildFS9721Frames(t, buildFS9721Bits([4]string{"5", "6", "7", "8"}, 2, false, map[int]byte{
		0:  '1', // AC
		50: '1', // Hz
	}))

	m := &FS9721{}
	m.ProcessArray(firstA)

	v, u, f := m.ProcessArray(firstB)
	if v != 0 || u != "" || f != nil {
		t.Fatalf("new 8-byte header must reset aggregation; got value=%v unit=%q flags=%v", v, u, f)
	}

	v, u, f = m.ProcessArray(secondB)
	if math.Abs(v-56.78) > 1e-9 {
		t.Fatalf("value mismatch after reset: got %.12f want %.12f", v, 56.78)
	}
	if u != "Hz" {
		t.Fatalf("unit mismatch after reset: got %q want %q", u, "Hz")
	}
	if !reflect.DeepEqual(f, []string{"AC"}) {
		t.Fatalf("flags mismatch after reset: got %v want %v", f, []string{"AC"})
	}

	_ = secondA
}

func TestFS9721ProcessArray_UnknownChunkLengthIgnored(t *testing.T) {
	t.Parallel()

	m := &FS9721{}
	v, u, f := m.ProcessArray([]byte{1, 2, 3, 4, 5})
	if v != 0 || u != "" || f != nil {
		t.Fatalf("unexpected output for invalid chunk length: value=%v unit=%q flags=%v", v, u, f)
	}
}

func buildFS9721Bits(digits [4]string, decimalPos int, negative bool, setBits map[int]byte) string {
	bits := make([]byte, 56)
	for i := range bits {
		bits[i] = '0'
	}

	setRange(bits, 5, fsDigitBits[digits[0]])
	if decimalPos == 1 {
		bits[12] = '1'
	}

	setRange(bits, 13, fsDigitBits[digits[1]])
	if decimalPos == 2 {
		bits[20] = '1'
	}

	setRange(bits, 21, fsDigitBits[digits[2]])
	if decimalPos == 3 {
		bits[28] = '1'
	}

	setRange(bits, 29, fsDigitBits[digits[3]])

	if negative {
		bits[4] = '1'
	}

	for idx, v := range setBits {
		bits[idx] = v
	}

	return string(bits)
}

func buildFS9721Frames(t *testing.T, bits string) ([]byte, []byte) {
	t.Helper()

	if len(bits) != 56 {
		t.Fatalf("invalid bits length: got %d want 56", len(bits))
	}

	frame := make([]byte, 14)
	for i := 0; i < 14; i++ {
		nibbleBits := bits[i*4 : (i+1)*4]
		n, err := strconv.ParseUint(nibbleBits, 2, 8)
		if err != nil {
			t.Fatalf("failed to parse nibble %q: %v", nibbleBits, err)
		}
		frame[i] = byte(n)
	}

	first := append([]byte(nil), frame[:8]...)
	second := append([]byte(nil), frame[8:]...)
	return first, second
}

func setRange(dst []byte, start int, value string) {
	for i := 0; i < len(value); i++ {
		dst[start+i] = value[i]
	}
}

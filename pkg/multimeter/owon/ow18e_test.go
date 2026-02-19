package owon

import (
	"math"
	"reflect"
	"testing"
)

func TestOW18EProcessArray_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		payload   []byte
		wantValue float64
		wantUnit  string
		wantFlags []string
	}{
		{
			name:      "documented AC voltage frame",
			payload:   []byte{98, 240, 4, 0, 147, 49},
			wantValue: 126.91,
			wantUnit:  "V",
			wantFlags: []string{"AC", "Auto Range"},
		},
		{
			name:      "negative DC voltage frame",
			payload:   []byte{34, 0, 0, 0, 210, 132},
			wantValue: -12.34,
			wantUnit:  "V",
			wantFlags: []string{"DC"},
		},
		{
			name:      "short frame is ignored",
			payload:   []byte{98, 240, 4, 0, 147},
			wantValue: 0,
			wantUnit:  "",
			wantFlags: nil,
		},
		{
			name:      "long frame keeps first 6 bytes",
			payload:   []byte{98, 240, 4, 0, 147, 49, 255, 255},
			wantValue: 126.91,
			wantUnit:  "V",
			wantFlags: []string{"AC", "Auto Range"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &OW18E{}
			gotValue, gotUnit, gotFlags := m.ProcessArray(tc.payload)

			if math.Abs(gotValue-tc.wantValue) > 1e-9 {
				t.Fatalf("value mismatch: got %.12f want %.12f", gotValue, tc.wantValue)
			}

			if gotUnit != tc.wantUnit {
				t.Fatalf("unit mismatch: got %q want %q", gotUnit, tc.wantUnit)
			}

			if !reflect.DeepEqual(gotFlags, tc.wantFlags) {
				t.Fatalf("flags mismatch: got %v want %v", gotFlags, tc.wantFlags)
			}
		})
	}
}

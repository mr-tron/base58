package base58

import (
	"bytes"
	"strings"
	"testing"
)

func TestTrivialBase58DecodingRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{name: "empty", input: "", wantErr: "zero length string"},
		{name: "invalid-digit", input: "0", wantErr: "invalid base58 digit"},
		{name: "non-ascii", input: string([]byte{0xc3, 0xa9}), wantErr: "high-bit set on invalid digit"},
		{name: "trailing-high-bit", input: "z" + string([]byte{0xff}), wantErr: "high-bit set on invalid digit"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("decoder panicked for %q: %v", tc.input, r)
				}
			}()

			_, err := TrivialBase58DecodingAlphabet(tc.input, BTCAlphabet)
			if err == nil {
				t.Fatalf("decoder(%q) returned nil error, want %q", tc.input, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("decoder(%q) error = %q, want substring %q", tc.input, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestTrivialBase58DecodingPreservesLeadingZeros(t *testing.T) {
	tests := []struct {
		input string
		want  []byte
	}{
		{input: "1", want: []byte{0}},
		{input: "11", want: []byte{0, 0}},
		{input: "1111111111", want: make([]byte, 10)},
		{input: "11z", want: []byte{0, 0, 57}},
		{input: "2g", want: []byte{0x61}},
	}

	for _, tc := range tests {
		got, err := TrivialBase58DecodingAlphabet(tc.input, BTCAlphabet)
		if err != nil {
			t.Fatalf("decoder(%q) error = %v", tc.input, err)
		}
		if !bytes.Equal(got, tc.want) {
			t.Fatalf("decoder(%q) = %x, want %x", tc.input, got, tc.want)
		}
	}
}

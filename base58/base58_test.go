package base58

import "testing"

func TestTrivialBase58DecodingEmpty(t *testing.T) {
	got, err := TrivialBase58Decoding("")
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("decoding empty string returned (%#v, %v), want ([]byte{}, nil)", got, err)
	}
}

func TestTrivialBase58DecodingInvalidByte(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
	}{
		{"\xff", "illegal base58 data at input index: 0"},
		{"11é", "illegal base58 data at input index: 2"},
		{"1z0", "illegal base58 data at input index: 2"},
	} {
		if _, err := TrivialBase58Decoding(tc.input); err == nil || err.Error() != tc.want {
			t.Errorf("decoding %q returned error %v, want %q", tc.input, err, tc.want)
		}
	}
}

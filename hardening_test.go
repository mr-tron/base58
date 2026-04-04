package base58

import (
	"bytes"
	"encoding/hex"
	"math"
	"math/big"
	"strings"
	"testing"
)

func TestKnownVectors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		hex string
		enc string
	}{
		{"", ""},
		{"61", "2g"},
		{"626262", "a3gV"},
		{"636363", "aPEr"},
		{"73696d706c792061206c6f6e6720737472696e67", "2cFupjhnEsSn59qHXstmK2ffpLv2"},
		{"516b6fcd0f", "ABnLTmg"},
		{"bf4f89001e670274dd", "3SEo3LWLoPntC"},
		{"572e4794", "3EFU7m"},
		{"ecac89cad93923c02321", "EJDM8drfXA6uyA"},
		{"10c8511e", "Rt5zm"},
		{"00000000000000000000", "1111111111"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.enc, func(t *testing.T) {
			data := mustHex(t, tc.hex)

			if got := Encode(data); got != tc.enc {
				t.Fatalf("Encode() = %q, want %q", got, tc.enc)
			}
			if got := FastBase58Encoding(data); got != tc.enc {
				t.Fatalf("FastBase58Encoding() = %q, want %q", got, tc.enc)
			}
			if got := TrivialBase58Encoding(data); got != tc.enc {
				t.Fatalf("TrivialBase58Encoding() = %q, want %q", got, tc.enc)
			}

			if tc.enc == "" {
				return
			}

			decoded, err := Decode(tc.enc)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if !bytes.Equal(decoded, data) {
				t.Fatalf("Decode() = %x, want %x", decoded, data)
			}

			decoded, err = FastBase58Decoding(tc.enc)
			if err != nil {
				t.Fatalf("FastBase58Decoding() error = %v", err)
			}
			if !bytes.Equal(decoded, data) {
				t.Fatalf("FastBase58Decoding() = %x, want %x", decoded, data)
			}

			decoded, err = TrivialBase58Decoding(tc.enc)
			if err != nil {
				t.Fatalf("TrivialBase58Decoding() error = %v", err)
			}
			if !bytes.Equal(decoded, data) {
				t.Fatalf("TrivialBase58Decoding() = %x, want %x", decoded, data)
			}
		})
	}
}

func TestLeadingZerosArePreservedAcrossChunkBoundaries(t *testing.T) {
	t.Parallel()

	alphabetNames := map[string]*Alphabet{
		"btc":    BTCAlphabet,
		"flickr": FlickrAlphabet,
		"custom": NewAlphabet(reverseString(btcDigits)),
	}
	tails := [][]byte{
		nil,
		{1},
		{1, 2, 3, 4, 5},
		mustHex(t, "0102030405060708090a0b0c0d0e0f10"),
		mustHex(t, "ffffffffffffffffffffffffffffffff"),
	}

	for name, alphabet := range alphabetNames {
		name, alphabet := name, alphabet
		t.Run(name, func(t *testing.T) {
			for _, zeroCount := range []int{1, 2, 7, 8, 9, 10, 11, 32, 64} {
				for _, tail := range tails {
					payload := append(make([]byte, zeroCount), tail...)
					wantPrefix := strings.Repeat(string(alphabet.encode[0]), zeroCount)
					wantTail := ""
					if len(tail) > 0 {
						wantTail = TrivialBase58EncodingAlphabet(tail, alphabet)
					}

					encoded := FastBase58EncodingAlphabet(payload, alphabet)
					if encoded != wantPrefix+wantTail {
						t.Fatalf("zeroCount=%d payload=%x encoded=%q want=%q", zeroCount, payload, encoded, wantPrefix+wantTail)
					}

					decoded, err := FastBase58DecodingAlphabet(encoded, alphabet)
					if err != nil {
						t.Fatalf("FastBase58DecodingAlphabet(%q) error = %v", encoded, err)
					}
					if !bytes.Equal(decoded, payload) {
						t.Fatalf("FastBase58DecodingAlphabet(%q) = %x, want %x", encoded, decoded, payload)
					}
				}
			}
		})
	}
}

func TestDecodeRejectsMalformedInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{name: "empty", input: "", wantErr: "zero length string"},
		{name: "zero", input: "0", wantErr: "invalid base58 digit"},
		{name: "capital-o", input: "O", wantErr: "invalid base58 digit"},
		{name: "capital-i", input: "I", wantErr: "invalid base58 digit"},
		{name: "lower-l", input: "l", wantErr: "invalid base58 digit"},
		{name: "space", input: "12 3", wantErr: "invalid base58 digit"},
		{name: "high-bit", input: string([]byte{0x80}), wantErr: "high-bit set on invalid digit"},
		{name: "high-bit-after-leading-zero", input: "11" + string([]byte{0xff}), wantErr: "high-bit set on invalid digit"},
	}

	decoders := []struct {
		name string
		fn   func(string) ([]byte, error)
	}{
		{name: "Decode", fn: Decode},
		{name: "FastBase58Decoding", fn: FastBase58Decoding},
		{name: "TrivialBase58Decoding", fn: TrivialBase58Decoding},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for _, decoder := range decoders {
				decoder := decoder
				t.Run(decoder.name, func(t *testing.T) {
					assertDecodeError(t, decoder.fn, tc.input, tc.wantErr)
				})
			}
		})
	}
}

func TestRoundTripOnDeterministicBoundaryPayloads(t *testing.T) {
	t.Parallel()

	alphabetNames := map[string]*Alphabet{
		"btc":    BTCAlphabet,
		"flickr": FlickrAlphabet,
		"custom": NewAlphabet(btcDigits[17:] + btcDigits[:17]),
	}
	lengths := []int{1, 2, 7, 8, 9, 10, 11, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 129, 255}
	zeroCounts := []int{0, 1, 2, 7, 8, 9}

	for name, alphabet := range alphabetNames {
		name, alphabet := name, alphabet
		t.Run(name, func(t *testing.T) {
			for _, length := range lengths {
				for _, zeroCount := range zeroCounts {
					payload := append(make([]byte, zeroCount), deterministicBytes(length)...)

					fastEncoded := FastBase58EncodingAlphabet(payload, alphabet)
					trivialEncoded := TrivialBase58EncodingAlphabet(payload, alphabet)
					if fastEncoded != trivialEncoded {
						t.Fatalf("length=%d zeroCount=%d fast=%q trivial=%q", length, zeroCount, fastEncoded, trivialEncoded)
					}

					fastDecoded, err := FastBase58DecodingAlphabet(fastEncoded, alphabet)
					if err != nil {
						t.Fatalf("FastBase58DecodingAlphabet(length=%d, zeroCount=%d) error = %v", length, zeroCount, err)
					}
					if !bytes.Equal(fastDecoded, payload) {
						t.Fatalf("FastBase58DecodingAlphabet(length=%d, zeroCount=%d) = %x, want %x", length, zeroCount, fastDecoded, payload)
					}

					trivialDecoded, err := TrivialBase58DecodingAlphabet(trivialEncoded, alphabet)
					if err != nil {
						t.Fatalf("TrivialBase58DecodingAlphabet(length=%d, zeroCount=%d) error = %v", length, zeroCount, err)
					}
					if !bytes.Equal(trivialDecoded, payload) {
						t.Fatalf("TrivialBase58DecodingAlphabet(length=%d, zeroCount=%d) = %x, want %x", length, zeroCount, trivialDecoded, payload)
					}
				}
			}
		})
	}
}

func TestMulAddBase58WordsLEMatchesBigInt(t *testing.T) {
	t.Parallel()

	wordSets := [][]uint64{
		nil,
		{},
		{0},
		{1},
		{57},
		{math.MaxUint64},
		{0, 1},
		{1, 0, 1},
		{math.MaxUint64, 0, math.MaxUint64},
		{0x0123456789abcdef, 0xfedcba9876543210},
	}
	addends := []uint64{0, 1, 57, 58, 123456789, math.MaxUint32, math.MaxUint64}

	for digits := 1; digits <= base58EncodeChunkDigits; digits++ {
		mul := base58ChunkPowers[digits]
		for _, originalWords := range wordSets {
			for _, add := range addends {
				got := append([]uint64(nil), originalWords...)
				want := wordsLEToBigInt(originalWords)
				want.Mul(want, new(big.Int).SetUint64(mul))
				want.Add(want, new(big.Int).SetUint64(add))

				mulAddBase58WordsLE(&got, mul, add)

				if wordsLEToBigInt(got).Cmp(want) != 0 {
					t.Fatalf("words=%#v mul=%d add=%d got=%#v want=%s", originalWords, mul, add, got, want.String())
				}
			}
		}
	}
}

func assertDecodeError(t *testing.T, fn func(string) ([]byte, error), input, wantErr string) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("decoder panicked for %q: %v", input, r)
		}
	}()

	_, err := fn(input)
	if err == nil {
		t.Fatalf("decoder(%q) returned nil error, want %q", input, wantErr)
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("decoder(%q) error = %q, want substring %q", input, err.Error(), wantErr)
	}
}

func deterministicBytes(n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte((i*37+n*17)%251 + 1)
	}
	return out
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()

	if s == "" {
		return nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex.DecodeString(%q): %v", s, err)
	}
	return b
}

func wordsLEToBigInt(words []uint64) *big.Int {
	n := new(big.Int)
	for i := len(words) - 1; i >= 0; i-- {
		n.Lsh(n, 64)
		n.Add(n, new(big.Int).SetUint64(words[i]))
	}
	return n
}

func reverseString(s string) string {
	b := []byte(s)
	for i := 0; i < len(b)/2; i++ {
		j := len(b) - 1 - i
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

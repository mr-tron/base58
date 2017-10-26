package base58

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

type testValues struct {
	dec, enc string // decoded hex value
}

var n = 5000000
var testPairs = make([]testValues, 0, n)

func initTestPairs() {
	if len(testPairs) > 0 {
		return
	}
	// pre-make the test pairs, so it doesn't take up benchmark time...
	data := make([]byte, 32)
	for i := 0; i < n; i++ {
		rand.Read(data)
		testPairs = append(testPairs, testValues{dec: hex.EncodeToString(data), enc: FastBase58Encoding(data)})
	}
}

func TestFastEqTrivialEncodingAndDecoding(t *testing.T) {
	for j := 1; j < 256; j++ {
		var b = make([]byte, j)
		for i := 0; i < 100; i++ {
			rand.Read(b)
			fe := FastBase58Encoding(b)
			te := TrivialBase58Encoding(b)

			if fe != te {
				t.Errorf("encoding err: %#v", hex.EncodeToString(b))
			}

			fd, ferr := FastBase58Decoding(fe)
			if ferr != nil {
				t.Errorf("fast error: %v", ferr)
			}
			td, terr := TrivialBase58Decoding(te)
			if terr != nil {
				t.Errorf("trivial error: %v", terr)
			}

			if hex.EncodeToString(fd) != hex.EncodeToString(td) {
				t.Errorf("decoding err: [%x] %s != %s", b, hex.EncodeToString(fd), hex.EncodeToString(td))
			}
		}
	}
}

func BenchmarkTrivialBase58Encoding(b *testing.B) {
	data := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		rand.Read(data)
		TrivialBase58Encoding(data)
	}
}

func BenchmarkFastBase58Encoding(b *testing.B) {
	data := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		rand.Read(data)
		FastBase58Encoding(data)
	}
}

func BenchmarkTrivialBase58Decoding(b *testing.B) {
	initTestPairs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		TrivialBase58Decoding(testPairs[i].enc)
	}
}

func BenchmarkFastBase58Decoding(b *testing.B) {
	initTestPairs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		FastBase58Decoding(testPairs[i].enc)
	}
}

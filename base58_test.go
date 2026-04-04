package base58

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"
)

type testValues struct {
	dec []byte
	enc string
}

var n = 5000000
var testPairs = make([]testValues, 0, n)
var testPairsByLength = make(map[int][]testValues)

func init() {
	// If we do not seed the prng - it will default to a seed of (1)
	// https://golang.org/pkg/math/rand/#Seed
	rand.Seed(time.Now().UTC().UnixNano())
}

func initTestPairs() {
	if len(testPairs) > 0 {
		return
	}
	testPairs = initTestPairsWithSize(32)
}

func initTestPairsWithSize(size int) []testValues {
	if pairs, ok := testPairsByLength[size]; ok && len(pairs) > 0 {
		return pairs
	}

	// pre-make the test pairs, so it doesn't take up benchmark time...
	pairs := make([]testValues, 0, n)
	for i := 0; i < n; i++ {
		data := make([]byte, size)
		rand.Read(data)
		pairs = append(pairs, testValues{dec: data, enc: FastBase58Encoding(data)})
	}

	testPairsByLength[size] = pairs
	return pairs
}

func randAlphabet() *Alphabet {
	// Permutes [0, 127] and returns the first 58 elements.
	var randomness [128]byte
	rand.Read(randomness[:])

	var bts [128]byte
	for i, r := range randomness {
		j := int(r) % (i + 1)
		bts[i] = bts[j]
		bts[j] = byte(i)
	}
	return NewAlphabet(string(bts[:58]))
}

var btcDigits = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func TestInvalidAlphabetTooShort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on alphabet being too short did not occur")
		}
	}()

	_ = NewAlphabet(btcDigits[1:])
}

func TestInvalidAlphabetTooLong(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on alphabet being too long did not occur")
		}
	}()

	_ = NewAlphabet("0" + btcDigits)
}

func TestInvalidAlphabetNon127(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on alphabet containing non-ascii chars did not occur")
		}
	}()

	_ = NewAlphabet("\xFF" + btcDigits[1:])
}

func TestInvalidAlphabetDup(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on alphabet containing duplicate chars did not occur")
		}
	}()

	_ = NewAlphabet("z" + btcDigits[1:])
}

func TestFastEqTrivialEncodingAndDecoding(t *testing.T) {
	for k := 0; k < 10; k++ {
		testEncDecLoop(t, randAlphabet())
	}
	testEncDecLoop(t, BTCAlphabet)
	testEncDecLoop(t, FlickrAlphabet)
}

func testEncDecLoop(t *testing.T, alph *Alphabet) {
	for j := 1; j < 256; j++ {
		var b = make([]byte, j)
		for i := 0; i < 100; i++ {
			rand.Read(b)
			fe := FastBase58EncodingAlphabet(b, alph)
			te := TrivialBase58EncodingAlphabet(b, alph)

			if fe != te {
				t.Errorf("encoding err: %#v", hex.EncodeToString(b))
			}

			fd, ferr := FastBase58DecodingAlphabet(fe, alph)
			if ferr != nil {
				t.Errorf("fast error: %v", ferr)
			}
			td, terr := TrivialBase58DecodingAlphabet(te, alph)
			if terr != nil {
				t.Errorf("trivial error: %v", terr)
			}

			if hex.EncodeToString(b) != hex.EncodeToString(td) {
				t.Errorf("decoding err: %s != %s", hex.EncodeToString(b), hex.EncodeToString(td))
			}
			if hex.EncodeToString(b) != hex.EncodeToString(fd) {
				t.Errorf("decoding err: %s != %s", hex.EncodeToString(b), hex.EncodeToString(fd))
			}
		}
	}
}

func BenchmarkTrivialBase58Encoding(b *testing.B) {
	benchmarkTrivialBase58EncodingWithSize(b, 32)
}

func BenchmarkFastBase58Encoding(b *testing.B) {
	benchmarkFastBase58EncodingWithSize(b, 32)
}

func BenchmarkTrivialBase58Decoding(b *testing.B) {
	benchmarkTrivialBase58DecodingWithSize(b, 32)
}

func BenchmarkFastBase58Decoding(b *testing.B) {
	benchmarkFastBase58DecodingWithSize(b, 32)
}

func BenchmarkTrivialBase58Encoding32(b *testing.B) {
	benchmarkTrivialBase58EncodingWithSize(b, 32)
}

func BenchmarkFastBase58Encoding32(b *testing.B) {
	benchmarkFastBase58EncodingWithSize(b, 32)
}

func BenchmarkTrivialBase58Encoding36(b *testing.B) {
	benchmarkTrivialBase58EncodingWithSize(b, 36)
}

func BenchmarkFastBase58Encoding36(b *testing.B) {
	benchmarkFastBase58EncodingWithSize(b, 36)
}

func BenchmarkTrivialBase58Encoding64(b *testing.B) {
	benchmarkTrivialBase58EncodingWithSize(b, 64)
}

func BenchmarkFastBase58Encoding64(b *testing.B) {
	benchmarkFastBase58EncodingWithSize(b, 64)
}

func BenchmarkTrivialBase58Encoding256(b *testing.B) {
	benchmarkTrivialBase58EncodingWithSize(b, 256)
}

func BenchmarkFastBase58Encoding256(b *testing.B) {
	benchmarkFastBase58EncodingWithSize(b, 256)
}

func BenchmarkTrivialBase58Decoding32(b *testing.B) {
	benchmarkTrivialBase58DecodingWithSize(b, 32)
}

func BenchmarkFastBase58Decoding32(b *testing.B) {
	benchmarkFastBase58DecodingWithSize(b, 32)
}

func BenchmarkTrivialBase58Decoding36(b *testing.B) {
	benchmarkTrivialBase58DecodingWithSize(b, 36)
}

func BenchmarkFastBase58Decoding36(b *testing.B) {
	benchmarkFastBase58DecodingWithSize(b, 36)
}

func BenchmarkTrivialBase58Decoding64(b *testing.B) {
	benchmarkTrivialBase58DecodingWithSize(b, 64)
}

func BenchmarkFastBase58Decoding64(b *testing.B) {
	benchmarkFastBase58DecodingWithSize(b, 64)
}

func BenchmarkTrivialBase58Decoding256(b *testing.B) {
	benchmarkTrivialBase58DecodingWithSize(b, 256)
}

func BenchmarkFastBase58Decoding256(b *testing.B) {
	benchmarkFastBase58DecodingWithSize(b, 256)
}

func benchmarkTrivialBase58EncodingWithSize(b *testing.B, size int) {
	pairs := initTestPairsWithSize(size)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		TrivialBase58Encoding(pairs[i%len(pairs)].dec)
	}
}

func benchmarkFastBase58EncodingWithSize(b *testing.B, size int) {
	pairs := initTestPairsWithSize(size)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		FastBase58Encoding(pairs[i%len(pairs)].dec)
	}
}

func benchmarkTrivialBase58DecodingWithSize(b *testing.B, size int) {
	pairs := initTestPairsWithSize(size)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		TrivialBase58Decoding(pairs[i%len(pairs)].enc)
	}
}

func benchmarkFastBase58DecodingWithSize(b *testing.B, size int) {
	pairs := initTestPairsWithSize(size)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		FastBase58Decoding(pairs[i%len(pairs)].enc)
	}
}

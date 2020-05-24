package base58

import (
	"fmt"
)

// Encode encodes the passed bytes into a base58 encoded string.
func Encode(bin []byte) string {
	return FastBase58EncodingAlphabet(bin, BTCAlphabet)
}

// EncodeAlphabet encodes the passed bytes into a base58 encoded string with the
// passed alphabet.
func EncodeAlphabet(bin []byte, alphabet *Alphabet) string {
	return FastBase58EncodingAlphabet(bin, alphabet)
}

// FastBase58Encoding encodes the passed bytes into a base58 encoded string.
func FastBase58Encoding(bin []byte) string {
	return FastBase58EncodingAlphabet(bin, BTCAlphabet)
}

// FastBase58EncodingAlphabet encodes the passed bytes into a base58 encoded
// string with the passed alphabet.
func FastBase58EncodingAlphabet(bin []byte, alphabet *Alphabet) string {
	zero := alphabet.encode[0]

	binsz := len(bin)
	var i, j, zcount, high int
	var carry uint32

	for zcount < binsz && bin[zcount] == 0 {
		zcount++
	}

	size := ((binsz-zcount)*138/100 + 1)

	// allocate one big buffer up front
	buf := make([]byte, size*2+zcount)

	// use the second half for the temporary buffer
	tmp := buf[size+zcount:]

	high = size - 1
	for i = zcount; i < binsz; i++ {
		j = size - 1
		for carry = uint32(bin[i]); j > high || carry != 0; j-- {
			carry = carry + 256*uint32(tmp[j])
			tmp[j] = byte(carry % 58)
			carry /= 58
		}
		high = j
	}

	for j = 0; j < size && tmp[j] == 0; j++ {
	}

	// Use the first half for the result
	b58 := buf[:size-j+zcount]

	if zcount != 0 {
		for i = 0; i < zcount; i++ {
			b58[i] = zero
		}
	}

	for i = zcount; j < size; i++ {
		b58[i] = alphabet.encode[tmp[j]]
		j++
	}

	return string(b58)
}

// Decode decodes the base58 encoded bytes.
func Decode(str string) ([]byte, error) {
	return FastBase58DecodingAlphabet(str, BTCAlphabet)
}

// DecodeAlphabet decodes the base58 encoded bytes using the given b58 alphabet.
func DecodeAlphabet(str string, alphabet *Alphabet) ([]byte, error) {
	return FastBase58DecodingAlphabet(str, alphabet)
}

// FastBase58Decoding decodes the base58 encoded bytes.
func FastBase58Decoding(str string) ([]byte, error) {
	return FastBase58DecodingAlphabet(str, BTCAlphabet)
}

// FastBase58DecodingAlphabet decodes the base58 encoded bytes using the given
// b58 alphabet.
func FastBase58DecodingAlphabet(str string, alphabet *Alphabet) ([]byte, error) {
	if len(str) == 0 {
		return nil, fmt.Errorf("zero length string")
	}

	zero := alphabet.encode[0]
	b58sz := len(str)

	var zcount int
	for i := 0; i < b58sz && str[i] == zero; i++ {
		zcount++
	}

	var t, c uint64

	// the 32bit algo stretches the result up to 2 times
	binu := make([]byte, 2*((b58sz*406/555)+1))
	outi := make([]uint32, (b58sz+3)/4)

	for _, r := range str {
		if r > 127 {
			return nil, fmt.Errorf("high-bit set on invalid digit")
		}
		if alphabet.decode[r] == -1 {
			return nil, fmt.Errorf("invalid base58 digit (%q)", r)
		}

		c = uint64(alphabet.decode[r])

		for j := len(outi) - 1; j >= 0; j-- {
			t = uint64(outi[j])*58 + c
			c = t >> 32
			outi[j] = uint32(t & 0xffffffff)
		}
	}

	// initial mask depends on b58sz, on further loops it always starts at 24 bits
	mask := (uint(b58sz%4) * 8)
	if mask == 0 {
		mask = 32
	}
	mask -= 8
	var j, cnt int
	for j, cnt = 0, 0; j < len(outi); j++ {
		for mask < 32 { // loop relies on uint overflow
			binu[cnt] = byte(outi[j] >> mask)
			mask -= 8
			cnt++
		}
		mask = 24
	}

	for n := zcount; n < len(binu); n++ {
		if binu[n] > 0 {
			return binu[n-zcount : cnt], nil
		}
	}
	return binu[:cnt], nil
}

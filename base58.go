package main

import (
	"fmt"
	"math/big"
)

const b58set = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var decodeMap [256]int8

func init() {
	for i := range decodeMap {
		decodeMap[i] = -1
	}
	for i, b := range b58set {
		decodeMap[b] = int8(i)
	}
}

var (
	bn0  = big.NewInt(0)
	bn58 = big.NewInt(58)
)

func FastBase58Encoding(bin []byte) string {
	binsz := len(bin)
	var i, j, high, zcount, carry int

	for zcount < binsz && bin[zcount] == 0 {
		zcount++
	}

	size := (binsz-zcount)*138/100 + 1
	var buf = make([]byte, size)

	high = size - 1
	for i = zcount; i < binsz; i++ {
		j = size - 1
		for carry = int(bin[i]); j > high || carry != 0; j-- {
			carry = carry + 256*int(buf[j])
			buf[j] = byte(carry % 58)
			carry /= 58
		}
		high = j
	}

	for j = 0; j < size && buf[j] == 0; j++ {
	}

	var b58 = make([]byte, size-j+zcount)

	if zcount != 0 {
		for i = 0; i < zcount; i++ {
			b58[i] = '1'
		}
	}

	for i = zcount; j < size; i++ {
		b58[i] = b58set[buf[j]]
		j += 1
	}

	return string(b58)
}

func TrivialBase58Encoding(a []byte) string {
	idx := len(a)*138/100 + 1
	buf := make([]byte, idx)
	bn := new(big.Int).SetBytes(a)
	var mo *big.Int
	for bn.Cmp(bn0) != 0 {
		bn, mo = bn.DivMod(bn, bn58, new(big.Int))
		idx--
		buf[idx] = b58set[mo.Int64()]
	}
	for i := range a {
		if a[i] != 0 {
			break
		}
		idx--
		buf[idx] = b58set[0]
	}
	return string(buf[idx:])
}

func FastBase58Decoding(str string) ([]byte, error) {
	if len(str) == 0 {
		return nil, fmt.Errorf("zero length string")
	}

	var (
		t        uint64
		zmask, c uint32
		zcount   int

		b58u  = []rune(str)
		b58sz = len(b58u)

		binsz     = len(b58u)
		outisz    = (binsz + 3) / 4 // check to see if we need to change this buffer size to optimize
		binu      = make([]byte, (binsz+3)*3)
		bytesleft = binsz % 4
	)

	if bytesleft > 0 {
		zmask = (0xffffffff << uint32(bytesleft*8))
	} else {
		bytesleft = 4
	}

	var outi = make([]uint32, outisz)

	var i = 0
	for ; i < b58sz && b58u[i] == '1'; i++ {
		zcount++
	}

	for ; i < b58sz; i++ {
		if b58u[i]&0x80 != 0 {
			return nil, fmt.Errorf("High-bit set on invalid digit")
		}

		if decodeMap[b58u[i]] == -1 {
			return nil, fmt.Errorf("Invalid base58 digit (%q)", b58u[i])
		}

		c = uint32(decodeMap[b58u[i]])

		for j := (outisz - 1); j >= 0; j-- {
			t = uint64(outi[j])*58 + uint64(c)
			c = uint32((t & 0x3f00000000) >> 32)
			outi[j] = uint32(t & 0xffffffff)
		}

		if c > 0 {
			return nil, fmt.Errorf("Output number too big (carry to the next int32)")
		}

		if outi[0]&zmask != 0 {
			return nil, fmt.Errorf("Output number too big (last int32 filled too far)")
		}
	}

	// the nested for-loop below is the same as the original code:
	// switch (bytesleft) {
	// 	case 3:
	// 		*(binu++) = (outi[0] & 0xff0000) >> 16;
	// 		//-fallthrough
	// 	case 2:
	// 		*(binu++) = (outi[0] & 0xff00) >>  8;
	// 		//-fallthrough
	// 	case 1:
	// 		*(binu++) = (outi[0] & 0xff);
	// 		++j;
	// 		//-fallthrough
	// 	default:
	// 		break;
	// }
	//
	// for (; j < outisz; ++j)
	// {
	// 	*(binu++) = (outi[j] >> 0x18) & 0xff;
	// 	*(binu++) = (outi[j] >> 0x10) & 0xff;
	// 	*(binu++) = (outi[j] >>    8) & 0xff;
	// 	*(binu++) = (outi[j] >>    0) & 0xff;
	// }
	var j, cnt int
	for j, cnt = 0, 0; j < outisz; j++ {
		for mask := byte(bytesleft-1) * 8; mask <= 0x18; mask, cnt = mask-8, cnt+1 {
			binu[cnt] = byte(outi[j] >> mask)
		}
		if j == 0 {
			bytesleft = 4 // because it could be less than 4 the first time through
		}
	}

	for n, v := range binu {
		if v > 0 {
			start := n - zcount
			if start < 0 {
				start = 0
			}
			return binu[start:cnt], nil
		}
	}

	return binu[:j], nil
}

// Decode decodes the base58 encoded bytes.
// based
func TrivialBase58Decoding(str string) ([]byte, error) {
	var zcount int
	for i := 0; i < len(str) && str[i] == '1'; i++ {
		zcount++
	}
	leading := make([]byte, zcount)

	var padChar rune = -1
	src := []byte(str)
	j := 0
	for ; j < len(src) && src[j] == byte(padChar); j++ {
	}

	n := new(big.Int)
	for i := range src[j:] {
		c := decodeMap[src[i]]
		if c == -1 {
			return nil, fmt.Errorf("illegal base58 data at input index: %d", i)
		}
		n.Mul(n, bn58)
		n.Add(n, big.NewInt(int64(c)))
	}
	return append(leading, n.Bytes()...), nil
}

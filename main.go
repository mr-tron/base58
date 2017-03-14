package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

const b58digits_ordered string = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func FastBase58Encoding(bin []byte) string {
	binsz := len(bin)
	var i, j, high, zcount, carry int

	for zcount < binsz && bin[zcount] == 0 {
		zcount++
	}

	size := (binsz-zcount)*138/100 + 1
	var buf = make([]byte, size)

	high = size - 1
	for i = zcount; i < binsz; i += 1 {
		j = size - 1
		for carry = int(bin[i]); j > high || carry != 0; j -= 1 {
			carry = carry + 256*int(buf[j])
			buf[j] = byte(carry % 58)
			carry /= 58
		}
		high = j
	}

	for j = 0; j < size && buf[j] == 0; j += 1 {
	}

	var b58 = make([]byte, size-j+zcount)

	if zcount != 0 {
		for i = 0; i < zcount; i++ {
			b58[i] = '1'
		}
	}

	for i = zcount; j < size; i += 1 {
		b58[i] = b58digits_ordered[buf[j]]
		j += 1
	}

	return string(b58)
}

var b58set string = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
var bn0 *big.Int = big.NewInt(0)
var bn58 *big.Int = big.NewInt(58)

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

func benchmark() {
	var b = make([]byte, 32)

	t := time.Now()
	for i := 0; i < 1000000; i++ {
		rand.Read(b)
		_ = FastBase58Encoding(b)
	}
	fmt.Println("One million operations with fast algorithm:", time.Since(t))

	t = time.Now()
	for i := 0; i < 1000000; i++ {
		rand.Read(b)
		_ = TrivialBase58Encoding(b)
	}
	fmt.Println("One million operations with trivial algorithm:", time.Since(t))
}

func main() {
	benchmark()
	
	for j := 0; j < 256; j++ {
		var b = make([]byte, j)
		for i := 0; i < 100; i++ {
			rand.Read(b)
			if FastBase58Encoding(b) != TrivialBase58Encoding(b) {
				fmt.Errorf(hex.EncodeToString(b))
				return
			}
		}
	}
	fmt.Println("Test passed")
}

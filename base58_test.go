package main

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestFastEqTrivialEncoding(t *testing.T) {
	for j := 0; j < 256; j++ {
		var b = make([]byte, j)
		for i := 0; i < 100; i++ {
			rand.Read(b)
			if FastBase58Encoding(b) != TrivialBase58Encoding(b) {
				t.Errorf(hex.EncodeToString(b))
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

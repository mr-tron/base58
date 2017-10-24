# go-base58 A Fast Implementation of Base58 encoding used in Bitcoin

Fast implementation of base58 encoding in Go (Golang). 

Base algorithm is copied from https://github.com/trezor/trezor-crypto/blob/master/base58.c

To import libarary

```go
	import (
		"github.com/pschlump/go-base58/base58"
	)
```

Without big.Int divisions it works more than 4 times faster.

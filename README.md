# go-base58

Fast implementation of base58 encoding on golang. 

Base algorithm is copied from https://github.com/trezor/trezor-crypto/blob/master/base58.c

If your know how to improve it, please make pull request.

Without big.Int divisions it works more than 4 times faster.

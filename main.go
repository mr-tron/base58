package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	var (
		err error
		bin []byte

		help    = flag.Bool("h", false, "display this message")
		lnBreak = flag.Int("b", 76, "break encoded string into num character lines. Use 0 to disable line wrapping")
		input   = flag.String("i", "-", `input file (use: "-" for stdin)`)
		output  = flag.String("o", "-", `output file (use: "-" for stdout)`)
		decode  = flag.Bool("d", false, `decode input`)
	)

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	fin, fout := os.Stdin, os.Stdout
	if *input != "-" {
		if fin, err = os.Open(*input); err != nil {
			fmt.Fprintf(os.Stderr, "input file err: %v\n", err)
			os.Exit(1)
		}
	}

	if *output != "-" {
		if fout, err = os.Create(*output); err != nil {
			fmt.Fprintf(os.Stderr, "output file err: %v\n", err)
			os.Exit(1)
		}
	}

	if bin, err = ioutil.ReadAll(fin); err != nil {
		fmt.Fprintf(os.Stderr, "read input err: %v\n", err)
		os.Exit(1)
	}

	if *decode {
		decoded, err := FastBase58Decoding(string(bin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode input err: %v\n", err)
			os.Exit(1)
		}
		io.Copy(fout, bytes.NewReader(decoded))
		os.Exit(0)
	}

	encoded := FastBase58Encoding(bin)
	if *lnBreak > 0 {
		lines := (len(encoded) / *lnBreak) + 1
		for i := 0; i < lines; i++ {
			start := i * *lnBreak
			end := start + *lnBreak
			if i == lines-1 {
				fmt.Fprintln(fout, encoded[start:])
				return
			}
			fmt.Fprintln(fout, encoded[start:end])
		}
	}
	fmt.Fprintln(fout, encoded)
}

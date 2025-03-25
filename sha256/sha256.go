package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
)

func main() {
	WIF := false
	if len(os.Args) > 1 && os.Args[1] == "-wif" {
		WIF = true
	}

	h := sha256.New()
	if WIF {
		h.Write([]byte{0x80})
	}
	n, err := io.Copy(h, os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, n, "bytes hashed")
	hs := h.Sum(nil)

	if WIF {
		fmt.Println(base58.Encode(hs))
	} else {
		fmt.Println(hex.EncodeToString(hs))
	}
}

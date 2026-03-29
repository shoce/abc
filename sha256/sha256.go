// GoGet GoFmt GoBuildNull
// GoBuild
// GoRun <readme.text

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"os"

	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
)

const (
	NL  = "\n"
	SEP = ","
)

var (
	WIF bool
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-wif" {
		WIF = true
	}

	h := sha256.New()
	if WIF {
		h.Write([]byte{0x80})
	}
	n, err := io.Copy(h, os.Stdin)
	if err != nil {
		perr("ERROR %v", err)
		os.Exit(1)
	}

	perr("<%s> bytes hashed", seps(uint64(n), 3))
	hs := h.Sum(nil)

	if WIF {
		pout(base58.Encode(hs))
	} else {
		pout(hex.EncodeToString(hs))
	}
}

func perr(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}

func pout(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stdout, msg+NL)
	} else {
		fmt.Fprintf(os.Stdout, msg+NL, args...)
	}
}

func seps(i uint64, e uint64) string {
	ee := uint64(math.Pow(10, float64(e)))
	if i < ee {
		return fmt.Sprintf("%d", i%ee)
	} else {
		f := fmt.Sprintf("0%dd", e)
		return fmt.Sprintf("%s"+SEP+"%"+f, seps(i/ee, e), i%ee)
	}
}

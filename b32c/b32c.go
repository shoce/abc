/*
HISTORY
2020/3/31 v1
2026/10/21 decode
*/
/*
GoGet
GoBuildNull
GoBuild
*/
package main
import (
	"encoding/base32"
	"fmt"
	"io"
	"os"
	"path/filepath"
)
const (
	CrockfordUpper = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	CrockfordLower = "0123456789abcdefghjkmnpqrstvwxyz"
	NL = "\n"
)
var (
	exit = os.Exit
	F = fmt.Sprintf
	pout = fmt.Print
)
func perr(msg string) { fmt.Fprint(os.Stderr, msg) }
func main() {
	var err error
	var DECODE bool
	encod := base32.NewEncoding(CrockfordLower)
	
	if filepath.Base(os.Args[0]) == "b32cd" { DECODE = true }
	args := os.Args[1:]
	for _, a := range args { switch a { case "-d", "d", "dec", "decode": DECODE = true } }
	if DECODE {
		de := base32.NewDecoder(encod, os.Stdin)
		_, err = io.Copy(os.Stdout, de)
		if err!=nil { perr(F("ERROR %v"+NL, err)) ; exit(1) }
		exit(0)
	}
	
	bb := make([]byte, 50)
	var n int
	for {
		// https://pkg.go.dev/io/
		n, err = os.Stdin.Read(bb)
		if err!=nil && err!=io.EOF { perr(F("ERROR %v"+NL, err)) ; exit(1) }
		if n==0 && err==io.EOF { break }
		if n==0 { continue }
		// https://pkg.go.dev/encoding/base32/
		pout(encod.EncodeToString(bb[:n])+NL)
	}
}


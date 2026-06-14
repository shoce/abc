// escape
/*
HISTORY
026/0620 v1
*/
/*
GoGet 
GoFmt 
GoBuildNull 
GoBuild
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	N = ""
	NL   = "\n"
	SPAC = "    "
	SEP = ","

	Rdigits = `\d+|\D+`

	ScannerBufferSize = 1 << 20 //ae:>>
)

var (
	R *regexp.Regexp

	F = fmt.Sprintf
	pout = fmt.Print
)

func init() {
	R = regexp.MustCompile(Rdigits)
}

func main() {
	// https://pkg.go.dev/bufio#Scanner
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(nil, ScannerBufferSize)

	var line string

	// https://pkg.go.dev/regexp#Regexp.FindAllString
	for scanner.Scan() {
		line = scanner.Text()
		ss := R.FindAllString(line, -1) 
		for i := range ss {
			if ss[i][0] >= '0' && ss[i][0] <= '9' {
				ss[i] = sepss(ss[i], 3)
			}
		}
		pout(strings.Join(ss, N)+NL)
	}

	if err := scanner.Err(); err != nil {
		perr(F("ERROR reading input %v", err))
		os.Exit(1)
	}
}

func sepss(q string, e int) (s string) {
	if e < 1 { return SEP+"🖕🏽"+SEP }
	qrr := []rune(q)
	var rr []rune = make([]rune, len(q)+len(q)/e+1)
	var sep rune = []rune(SEP)[0]
	j := len(rr)-1
	for i := len(qrr)-1 ; i >= 0 ; i-- {
		if (len(qrr)-1-i)%e==0 { rr[j] = sep ; j-- ; }
		rr[j] = qrr[i] ; j-- ;
	}
	return string(rr[j+1:])
}

func perr(msg string) { fmt.Fprint(os.Stderr, msg+NL) }


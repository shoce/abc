/*
USAGE
sre abc def <readme.text #>
srer '[0-9]' '#' <readme.text #>
*/
/*
INSTALL
ln -s sre srer
*/

// GoGet GoFmt GoBuildNull GoBuild

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	NL   = "\n"
	SPAC = "    "

	ScannerBufferSize = 900 << 10 //ae:>>
)

var (
	S1, S2 string
	R1     *regexp.Regexp

	ScannerBuffer []byte
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr,
			"USAGE sre S1 [S2]"+NL+
				SPAC+"S1 and S2 are literal strings"+NL+
				"USAGE srer R1 [S2]"+NL+
				SPAC+"R1 is a regexp, S2 is a string with $n for submatches"+NL,
		)
		os.Exit(1)
	}
	S1 = os.Args[1]
	if len(os.Args) == 3 {
		S2 = os.Args[2]
	}

	if filepath.Base(os.Args[0]) == "srer" {
		var err error
		if R1, err = regexp.Compile(S1); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR compile provided regular expression"+NL+"%v"+NL, err)
			os.Exit(1)
		}
	}

	// https://pkg.go.dev/bufio#Scanner
	scanner := bufio.NewScanner(os.Stdin)
	ScannerBuffer = make([]byte, ScannerBufferSize)
	scanner.Buffer(ScannerBuffer, ScannerBufferSize)
	//scanner.Split(scanner.Scan
	/*
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			//return scanner.ScanLines(data, atEOF)
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if i := bytes.IndexByte(data, '\n'); i >= 0 {
				return i + 1, dropCR(data[0:i]), nil
			}
			if atEOF {
				return len(data), dropCR(data), nil
			}
			return 0, nil, nil
		})
	*/

	var line1, line2 string

	if R1 == nil {
		for scanner.Scan() {
			line1 = scanner.Text()
			// https://pkg.go.dev/strings#ReplaceAll
			line2 = strings.ReplaceAll(line1, S1, S2)
			fmt.Println(line2)
		}
	} else {
		for scanner.Scan() {
			line1 = scanner.Text()
			// https://pkg.go.dev/regexp#Regexp.ReplaceAllString
			line2 = R1.ReplaceAllString(line1, S2)
			fmt.Println(line2)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR reading input %v"+NL, err)
		os.Exit(1)
	}
}

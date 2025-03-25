/*
history:
2020/04/20 v1

go get -a -u -v
go mod tidy

GoFmt GoBuildNull GoBuild GoRun
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	NL = "\n"
)

func main() {
	var err error
	var ijk []int
	var l string
	var ww []string

	if len(os.Args) < 2 {
		fmt.Fprintf(
			os.Stderr,
			"usage: wo i [j] [k] [...]"+NL+
				"	i, j, k, ... = position of the word to select, negative to select from the end of the line"+NL,
		)
		os.Exit(1)
	}

	for _, nn := range os.Args[1:] {
		n, err := strconv.Atoi(nn)
		if err != nil || n == 0 {
			fmt.Fprintf(os.Stderr, "invalid number '%s': should be digital and non-zero", nn)
			os.Exit(1)
		}
		ijk = append(ijk, n)
	}

	sp := regexp.MustCompile(`[^ \t]+([ \t]+|$)`)
	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {
		l = sc.Text()
		l = strings.TrimSpace(l)

		ww = sp.FindAllString(l, -1)

		for _, n := range ijk {
			if n > 0 && len(ww) < n {
				continue
			}
			if n < 0 && len(ww) < -n {
				continue
			}

			if n > 0 {
				fmt.Print(ww[n-1])
			} else if n < 0 {
				fmt.Print(ww[len(ww)+n])
			}
		}
		fmt.Println()
	}
	if err = sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read stdin: %v", err)
	}
}

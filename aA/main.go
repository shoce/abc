/*
history:
025/0521 v1

GoFmt
GoBuild

ln -s aA Aa
*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var transform func(string) string
	progname := filepath.Base(os.Args[0])
	switch progname {
	case "aA":
		transform = strings.ToUpper
	case "Aa":
		transform = strings.ToLower
	default:
		fmt.Fprintf(os.Stderr, "incorrect os.Args[0]=`%s`", os.Args[0])
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	buf := make([]byte, 1000)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			s := transform(string(buf[:n]))
			writer.WriteString(s)
		}
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "read error: %v", err)
			os.Exit(1)
		}
	}
}

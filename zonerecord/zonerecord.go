/*
history:
2020/4/3 v1

GoFmt
GoBuildNull
go build -o $HOME/bin/zonerecord
*/

package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	Zone, Record string
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: zonerecord dns.name.com")
		os.Exit(1)
	}

	dnsname := os.Args[1]
	dnsname = strings.TrimRight(dnsname, ".")
	words := strings.Split(dnsname, ".")
	if len(words) < 2 {
		fmt.Fprintln(os.Stderr, "argument must contain at least two words separated by dot")
		os.Exit(1)
	}

	p := len(words) - 2
	Zone = strings.Join(words[p:], ".")
	Record = strings.Join(words[:p], ".")
	if Record == "" {
		Record = "@"
	}
	fmt.Printf("Zone=%s Record=%s\n", Zone, Record)

	return
}

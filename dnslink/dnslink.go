/*
history:
2020/4/3 v1

GoFmt GoBuildNull GoBuild
*/

package main

import (
	"fmt"
	"os"
	"strings"
)

const NL = "\n"

var (
	Zone, Record, DnslinkRecord string
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dnslink dns.name.com")
		os.Exit(1)
	}

	dnsname := os.Args[1]
	dnsname = strings.TrimRight(dnsname, ".")
	words := strings.Split(dnsname, ".")
	if len(words) < 2 {
		fmt.Fprintln(os.Stderr, "argument must contain at least two words separated by dot")
		os.Exit(1)
	}

	words = append([]string{"_dnslink"}, words...)
	p := len(words) - 2
	Zone = strings.Join(words[p:], ".")
	Record = strings.Join(words[1:p], ".")
	if Record == "" {
		Record = "@"
	}
	DnslinkRecord = strings.Join(words[:p], ".")
	fmt.Printf("Zone:'%s' Record:'%s' DnslinkRecord:'%s'"+NL, Zone, Record, DnslinkRecord)

	return
}

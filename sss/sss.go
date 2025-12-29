/*
history:
20/1123 v1
*/

// GoGet GoFmt GoBuildNull GoBuild GoRun

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	psproc "github.com/shirou/gopsutil/v4/process"
)

const (
	SP = " "
	NL = "\n"
)

var (
	VERSION string
)

func fmtdur(d time.Duration) (s string) {
	days := d / (time.Hour * 24)
	secs := d % (time.Hour * 24) / time.Second
	s = fmt.Sprintf("%ds", secs)
	if days > 0 {
		s = fmt.Sprintf("%dd·", days) + s
	}
	return s
}

func init() {
	if len(os.Args) == 2 && (os.Args[1] == "version" || os.Args[1] == "-version") {
		fmt.Println(VERSION)
		os.Exit(0)
	}
}

func main() {
	procs, err := psproc.Processes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error psproc.Processes %v", err)
		os.Exit(1)
	}

	for _, p := range procs {
		pname, err := p.Name()
		if err != nil {
			pname = ""
		}

		pcreatetime, err := p.CreateTime()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error p.CreateTime %v", err)
			os.Exit(1)
		}
		puptime := time.Since(time.Unix(pcreatetime/1000, 0))

		pconns, err := p.Connections()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error p.Connections %v", err)
			os.Exit(1)
		}

		sort.Slice(pconns, func(i, j int) bool {
			return pconns[i].Laddr.Port < pconns[j].Laddr.Port
		})

		var plistens []string
		for _, c := range pconns {
			if pname == "docker-proxy" {
				continue
			}
			if c.Status != "LISTEN" {
				continue
			}
			claddr := c.Laddr.IP
			/*
				if strings.HasPrefix(claddr, "127.0.0.") || claddr == "::1" {
					continue
				}
			*/
			if claddr == "0.0.0.0" || claddr == "::" || claddr == "*" {
				claddr = ""
			}
			if c.Laddr.Port != 0 {
				claddr = fmt.Sprintf("%s:%d", claddr, c.Laddr.Port)
			}
			craddr := c.Raddr.IP
			if craddr == "0.0.0.0" || craddr == "::" || craddr == "*" {
				craddr = ""
			}
			if c.Raddr.Port != 0 {
				craddr = fmt.Sprintf("%s:%d", craddr, c.Raddr.Port)
			}
			l := claddr
			if craddr != "" {
				l += "/" + craddr
			}
			l = "[" + l + "]"
			add := true
			for _, p := range plistens {
				if l == p {
					add = false
				}
			}
			if add {
				plistens = append(plistens, l)
			}
		}

		if len(plistens) == 0 {
			continue
		}

		if len(os.Args) > 1 {
			print := false
			for _, a := range os.Args[1:] {
				if strings.Contains(pname, a) {
					print = true
				}
				for _, l := range plistens {
					if strings.HasSuffix(l, ":"+a) {
						print = true
					}
				}
			}
			if !print {
				continue
			}
		}

		fmt.Printf(
			"{ @pid <%d> @name [%s] @uptime <%s> @listens (%v) }"+NL,
			p.Pid, pname, fmtdur(puptime), strings.Join(plistens, SP),
		)
	}
}

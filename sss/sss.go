/*
Sockets Sockets Sockets

history:
20/1123 v1
*/

// GoGet GoFmt GoBuildNull GoBuild GoRun

package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	psproc "github.com/shirou/gopsutil/v4/process"
)

const (
	SP  = " "
	NL  = "\n"
	SEP = ","
)

var (
	VERSION string
)

func init() {
	if len(os.Args) == 2 && os.Args[1] == "-version" {
		fmt.Print(VERSION + NL)
		os.Exit(0)
	}
}

func main() {
	procs, err := psproc.Processes()
	if err != nil {
		perr("ERROR psproc.Processes %v", err)
		os.Exit(1)
	}

	for _, p := range procs {
		pname, err := p.Name()
		if err != nil {
			pname = ""
		}

		pcreatetime, err := p.CreateTime()
		if err != nil {
			perr("ERROR p.CreateTime %v", err)
			os.Exit(1)
		}
		puptime := time.Since(time.Unix(pcreatetime/1000, 0))

		pconns, err := p.Connections()
		if err != nil {
			perr("ERROR p.Connections %v", err)
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
			"<%d> up<%s> [%s] listens( %v )"+NL,
			p.Pid, fmtdur(puptime), pname, strings.Join(plistens, SP),
		)
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

func fmtdur(d time.Duration) (s string) {
	days := d / (time.Hour * 24)
	secs := d % (time.Hour * 24) / time.Second
	s = seps(uint64(secs), 2) + "s"
	if days > 0 {
		s = seps(uint64(days), 2) + "d" + SEP + s
	}
	return s
}

func perr(msg string, args ...interface{}) {
	msgtext := msg
	if len(args) > 0 {
		msgtext = fmt.Sprintf(msg, args...)
	}
	fmt.Fprint(os.Stderr, msgtext+NL)
}

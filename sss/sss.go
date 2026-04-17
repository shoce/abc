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

	F = fmt.Sprintf
)

func init() {
	if len(os.Args) == 2 && os.Args[1] == "-version" {
		pout(VERSION)
		os.Exit(0)
	}
}

func main() {

	// https://pkg.go.dev/github.com/shirou/gopsutil/v4/process#Processes
	procs, err := psproc.Processes()
	if err != nil {
		perr("ERROR psproc.Processes %v", err)
		os.Exit(1)
	}

	listens := make(map[string][]string)

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

		// https://pkg.go.dev/github.com/shirou/gopsutil/v4/process#Process.Connections
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
				claddr = F("%s:%d", claddr, c.Laddr.Port)
			}
			craddr := c.Raddr.IP
			if craddr == "0.0.0.0" || craddr == "::" || craddr == "*" {
				craddr = ""
			}
			if c.Raddr.Port != 0 {
				craddr = F("%s:%d", craddr, c.Raddr.Port)
			}
			l := claddr
			if craddr != "" {
				l += "/" + craddr
			}
			l = "[" + l + "]"

			add := true
			for _, p := range plistens {
				if p == l {
					add = false
				}
			}
			if add {
				plistens = append(plistens, l)
				listens[l] = append(listens[l], F("<%d>[%s]<%s>", p.Pid, pname, fmtdur(puptime)))
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
			pout(
				"<%d> up<%s> [%s] listens( %v )",
				p.Pid, fmtdur(puptime), pname, strings.Join(plistens, SP),
			)
		}

	}

	// https://pkg.go.dev/slices#SortFunc
	//slices.SortFunc(listens, cmp.Compare)

	for l, pp := range listens {
		pout(
			"%s ( %s )",
			l, strings.Join(pp, SP),
		)
	}

}

func seps(i uint64, e uint64) string {
	ee := uint64(math.Pow(10, float64(e)))
	if i < ee {
		return F("%d", i%ee)
	} else {
		f := F("0%dd", e)
		return F("%s"+SEP+"%"+f, seps(i/ee, e), i%ee)
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
		msgtext = F(msg, args...)
	}
	fmt.Fprint(os.Stderr, msgtext+NL)
}

func pout(msg string, args ...interface{}) {
	msgtext := msg
	if len(args) > 0 {
		msgtext = F(msg, args...)
	}
	fmt.Fprint(os.Stdout, msgtext+NL)
}

/*
Sockets Sockets Sockets

history:
20/1123 v1
*/

// GoGet GoFmt GoBuildNull GoBuild GoRun

package main

import (
	"cmp"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	psnet "github.com/shirou/gopsutil/v4/net"
	psproc "github.com/shirou/gopsutil/v4/process"
	"golang.org/x/exp/slices"
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
	listensstats := make(map[string]psnet.ConnectionStat)

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
			/*
				if strings.HasPrefix(c.Laddr.IP, "127.0.0.") || c.Laddr.IP == "::1" {
					continue
				}
			*/
			switch c.Laddr.IP {
			case "0.0.0.0":
				c.Laddr.IP = "0"
			case "::":
				//c.Laddr.IP = ""
			case "*":
				//c.Laddr.IP = ""
			}
			claddr := c.Laddr.IP
			if c.Laddr.Port != 0 {
				claddr = F("%s:%d", claddr, c.Laddr.Port)
			}
			switch c.Raddr.IP {
			case "0.0.0.0":
				c.Raddr.IP = "0"
			case "::":
				//c.Laddr.IP = ""
			case "*":
				//c.Laddr.IP = ""
			}
			craddr := c.Raddr.IP
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
				listensstats[l] = c
				listens[l] = append(listens[l], F(
					"<%d><%s>[%s]",
					p.Pid, fmtdur(puptime), pname,
				))
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

	var listenskk []string
	for l, _ := range listens {
		listenskk = append(listenskk, l)
	}
	// https://pkg.go.dev/slices#Sort
	slices.SortFunc(listenskk, func(a, b string) int {
		ca := listensstats[a]
		cb := listensstats[b]
		cmplip := cmp.Compare(ca.Laddr.IP, cb.Laddr.IP)
		cmplport := cmp.Compare(ca.Laddr.Port, cb.Laddr.Port)
		cmprip := cmp.Compare(ca.Raddr.IP, cb.Raddr.IP)
		cmprport := cmp.Compare(ca.Raddr.Port, cb.Raddr.Port)
		if cmplip != 0 {
			return cmplip
		}
		if cmplport != 0 {
			return cmplport
		}
		if cmprip != 0 {
			return cmprip
		}
		return cmprport
	})

	for _, l := range listenskk {
		pout(
			"%s ( %s )",
			l, strings.Join(listens[l], SP),
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

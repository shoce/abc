/*
history:
20/1123 v1

GoFmt GoBuild GoRelease
*/

package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	psproc "github.com/shirou/gopsutil/process"
)

const (
	NL = "\n"
)

var (
	VERSION string
)

func fmtdur(d time.Duration) string {
	days := d / (time.Minute * 1440)
	mins := d % (time.Minute * 1440) / time.Minute
	s := ""
	if mins > 0 {
		s = fmt.Sprintf("%dm%s", mins, s)
	}
	if days > 0 {
		s = fmt.Sprintf("%dd%s", days, s)
	}
	if s == "" {
		s = "0m"
	}
	return s
}

func init() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(VERSION)
		os.Exit(0)
	}
}

func main() {
	procs, err := psproc.Processes()
	if err != nil {
		log.Fatalf("psproc.Processes: %s", err)
	}

	for _, p := range procs {
		pname, err := p.Name()
		if err != nil {
			//log.Fatalf("p.Name: %s", err)
			pname = ""
		}

		pcreatetime, err := p.CreateTime()
		if err != nil {
			log.Fatalf("p.CreateTime: %s", err)
		}
		puptime := time.Since(time.Unix(pcreatetime/1000, 0))

		pconns, err := p.Connections()
		if err != nil {
			log.Fatalf("p.Connections: %s", err)
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
			"/proc/%d %s uptime=%s listens=%v"+NL,
			p.Pid, pname, fmtdur(puptime), plistens,
		)
	}
}

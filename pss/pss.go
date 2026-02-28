/*
history:
20/290 v1
20/293 option "0" to show the tree of processes starting with with pid=0
20/301 first arg is pid to specify the root process
20/307 proper sorting to build visual process tree
20/307 accept any number of arguments as filters by process id or by process name

go mod init github.com/shoce/pss
go get -a -u -v
go mod tidy

GoFmt
GoBuildNull
GoBuild
GoRun
*/

package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	ps "github.com/shoce/go-ps"
	sysconf "github.com/tklauser/go-sysconf"
)

const (
	SP  = " "
	TAB = "\t"
	NL  = "\n"
	SEP = ","
)

type Process struct {
	Pid  int64
	Ppid int64
	Pids []int64

	Name    string
	Cmdline string

	Utime uint64
	Stime uint64

	Starttime uint64

	Vsize uint64
	Rss   uint32

	Cgroup  string
	Kubepod bool
}

type Filter struct {
	Pid  int64
	Name string
}

var (
	VERSION string

	ClkTck   int64
	PageSize int

	pp []Process
	ff []Filter
)

func init() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(VERSION)
		os.Exit(0)
	}
}

func main() {
	var err error

	ClkTck, err = sysconf.Sysconf(sysconf.SC_CLK_TCK)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR %v"+NL, err)
		os.Exit(1)
	}

	PageSize = os.Getpagesize()

	for _, a := range os.Args[1:] {
		a = strings.TrimSpace(a)
		filtername := a
		filterpid, err := strconv.Atoi(a)
		if err != nil {
			filtername = a
		}
		ff = append(ff, Filter{Pid: int64(filterpid), Name: filtername})
	}
	if len(ff) == 0 {
		ff = []Filter{Filter{Pid: 1}}
	}

	// https://pkg.go.dev/github.com/shoce/go-ps#Processes
	pp0, err := ps.Processes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR %v"+NL, err)
		os.Exit(1)
	}

	for _, p0 := range pp0 {
		p := Process{
			Pid:       int64(p0.Pid()),
			Ppid:      int64(p0.PPid()),
			Name:      p0.Executable(),
			Cmdline:   p0.Cmdline(),
			Utime:     p0.Utime(),
			Stime:     p0.Stime(),
			Starttime: p0.Starttime(),
			Vsize:     p0.Vsize(),
			Rss:       p0.Rss(),
			Cgroup:    p0.Cgroup(),
			Kubepod:   strings.Contains(p0.Cgroup(), "/kubepods/"),
		}
		pp = append(pp, p)
	}

	sort.Slice(pp, func(i, j int) bool {
		if pp[i].Ppid < pp[j].Ppid {
			return true
		}
		if pp[i].Ppid > pp[j].Ppid {
			return false
		}
		return pp[i].Pid < pp[j].Pid
	})

	for i, p := range pp {
		pp[i].Pids = []int64{p.Pid}
		if p.Pid == p.Ppid || p.Ppid == 0 {
			continue
		}
		for _, q := range pp {
			if q.Pid == p.Ppid {
				pp[i].Pids = append(q.Pids, pp[i].Pids...)
			}
		}
	}

	sort.Slice(pp, func(i, j int) bool {
		ml := len(pp[i].Pids)
		if len(pp[j].Pids) < ml {
			ml = len(pp[j].Pids)
		}
		for k := 0; k < ml; k++ {
			if pp[i].Pids[k] < pp[j].Pids[k] {
				return true
			}
			if pp[i].Pids[k] > pp[j].Pids[k] {
				return false
			}
		}
		if len(pp[i].Pids) < len(pp[j].Pids) {
			return true
		}
		return false
	})

	for _, p := range pp {
		skip := true

		for _, f := range ff {
			if f.Name == "0" {
				skip = false
				break
			}
			if f.Name != "" && strings.Contains(p.Name, f.Name) {
				skip = false
				break
			}
			if f.Pid == 0 {
				continue
			}
			for _, pid := range p.Pids {
				if pid == f.Pid {
					skip = false
					break
				}
			}
		}

		if skip {
			continue
		}

		pidss := ""
		for _, pid := range p.Pids {
			pidss += fmt.Sprintf("%d", pid) + TAB
		}
		kubepods := ""
		if p.Kubepod {
			kubepods = "(k)" + TAB
		}

		procstats := ""
		if p.Utime > 0 || p.Vsize > 0 {
			procstats = fmt.Sprintf(
				"utime<%ss>stime<%ss>vsize<%skb>rss<%skb>",
				seps(p.Utime/uint64(ClkTck), 2),
				seps(p.Stime/uint64(ClkTck), 2),
				seps(p.Vsize/1024, 3),
				seps(uint64(p.Rss)*uint64(PageSize)/1024, 3),
			) + TAB
		}
		fmt.Printf(
			"%s"+"%s"+"%s"+"%s"+NL,
			pidss,
			kubepods,
			procstats,
			p.Cmdline,
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

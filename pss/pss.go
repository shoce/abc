/*
history:
20/290 v1
20/293 option "0" to show the tree of processes starting with with pid=0
20/301 first arg is pid to specify the root process
20/307 proper sorting to build visual process tree
20/307 accept any number of arguments as filters by process id or by process name
*/

// GoFixDiff GoFixFix
// GoGet GoFmt GoBuildNull GoBuild
// GoRun

package main

import (
	"fmt"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	sysconf "github.com/tklauser/go-sysconf"
)

const (
	SP  = " "
	TAB = "\t"
	NL  = "\n"
	SEP = ","
	N   = ""

	TagKubepod = "kubepod"
)

var (
	DEBUG bool
)

type Process struct {
	Pid   int64   // process id
	Ppid  int64   // parent process id
	Pids  []int64 // all ancestors process ids and process id
	Pgid  int64   // process group id
	Sid   int64   // session id
	TtyNr int     // controlling tty number
	Tpgid int

	State   rune     // process state
	Name    string   // process name
	Cmdline []string // command line

	utimeticks     uint64 // user cpu time in cpu ticks
	stimeticks     uint64 // system cpu time in cpu ticks
	starttimeticks uint64 // start time in cpu ticks

	Utime     time.Duration // user cpu time
	Stime     time.Duration // system cpu time
	Starttime time.Time     // start time

	Vsize int64 // virtual size
	Rss   int64 // resident set size

	Cgroup  string // cgroup
	Kubepod bool
}

type ProcessFilter struct {
	Pid  int64
	Text string
}

var (
	VERSION string

	BootTime time.Time
	ClkTck   int64
	PageSize int

	PP []Process
	FF []ProcessFilter
)

func init() {
	var err error

	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if os.Getenv("DEBUG") != "" {
		DEBUG = true
	}

	BootTime, err = GetBootTime()
	if err != nil {
		perr("ERROR GetBootTime %v", err)
		os.Exit(1)
	}
	perr("DEBUG BootTime <%s>", BootTime.Format("2006:0102:150405"))

	ClkTck, err = GetClkTck()
	if err != nil {
		perr("ERROR GetClkTck %v", err)
		os.Exit(1)
	}
	perr("DEBUG ClkTck <%d>", ClkTck)

	PageSize = os.Getpagesize()
	perr("DEBUG PageSize <%d>", PageSize)
}

func GetClkTck() (int64, error) {
	return sysconf.Sysconf(sysconf.SC_CLK_TCK)
}

func main() {
	var err error

	for _, a := range os.Args[1:] {
		a = strings.TrimSpace(a)
		filtertext := a
		filterpid, _ := strconv.Atoi(a)
		FF = append(FF, ProcessFilter{Pid: int64(filterpid), Text: filtertext})
	}
	if len(FF) == 0 {
		FF = []ProcessFilter{ProcessFilter{Pid: 1}}
	}

	PP, err = GetProcesses()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR GetProcesses %v"+NL, err)
		os.Exit(1)
	}

	sort.Slice(PP, func(i, j int) bool {
		if PP[i].Ppid < PP[j].Ppid {
			return true
		}
		if PP[i].Ppid > PP[j].Ppid {
			return false
		}
		return PP[i].Pid < PP[j].Pid
	})

	for i, p := range PP {
		PP[i].Pids = []int64{p.Pid}
		if p.Pid == p.Ppid || p.Ppid == 0 {
			continue
		}
		for _, q := range PP {
			if q.Pid == p.Ppid {
				PP[i].Pids = append(q.Pids, PP[i].Pids...)
			}
		}
	}

	sort.Slice(PP, func(i, j int) bool {
		ml := min(len(PP[i].Pids), len(PP[j].Pids))
		for k := 0; k < ml; k++ {
			if PP[i].Pids[k] < PP[j].Pids[k] {
				return true
			}
			if PP[i].Pids[k] > PP[j].Pids[k] {
				return false
			}
		}
		if len(PP[i].Pids) < len(PP[j].Pids) {
			return true
		}
		return false
	})

	for _, p := range PP {

		skip := true

		for _, f := range FF {
			if f.Text != "" {
				if strings.Contains(p.Name, f.Text) {
					skip = false
					break
				}
				for _, a := range p.Cmdline {
					if strings.Contains(a, f.Text) {
						skip = false
						break
					}
				}
			}
			if f.Pid > 0 {
				if slices.Contains(p.Pids, f.Pid) {
					skip = false
				}
			}
			if !skip {
				break
			}
		}

		if skip {
			continue
		}

		pids := make([]string, 0, 12)
		for _, pid := range p.Pids {
			pids = append(pids, fmt.Sprintf("%d", pid))
		}
		for i := range pids {
			pids[i] = fmt.Sprintf("<%s>", pids[i])
		}
		pidss := strings.Join(pids, N)

		var tags []string
		if p.Kubepod {
			tags = append(tags, TagKubepod)
		}
		for i := range tags {
			tags[i] = fmt.Sprintf("[%s]", tags[i])
		}
		tagss := strings.Join(tags, N)

		procstatss := []string{}
		if !BootTime.IsZero() && !p.Starttime.IsZero() {
			procstatss = append(procstatss,
				fmt.Sprintf("up<%s>",
					fmttime(uint64(time.Since(p.Starttime).Seconds()))),
			)
		}
		if p.Utime+p.Stime > 0 {
			procstatss = append(procstatss,
				fmt.Sprintf("cpu<%s>",
					fmttime(uint64((p.Utime+p.Stime).Seconds()))),
			)
		}
		if p.Vsize > 0 {
			procstatss = append(procstatss,
				fmt.Sprintf("rss<%skb>",
					seps(uint64(p.Rss)*uint64(PageSize)/1024, 3)),
			)
		}
		procstats := strings.Join(procstatss, SP)

		cmd := "[" + p.Cmdline[0] + "]"
		cmdargs := make([]string, 0, len(p.Cmdline))
		for _, a := range p.Cmdline[1:] {
			if strings.Contains(a, NL) {
				cmdargs = append(cmdargs, "[-"+NL+a+NL+"-]")
			} else {
				cmdargs = append(cmdargs, "["+a+"]")
			}
		}

		procinfo := pidss
		if tagss != "" {
			procinfo += tagss
		}
		procinfo += SP
		if procstats != "" {
			procinfo += procstats + SP
		}
		procinfo += cmd
		procinfo += strings.Join(cmdargs, N)
		procinfo += NL
		fmt.Print(procinfo)

	}
}

func perr(msg string, args ...interface{}) {
	if strings.HasPrefix(msg, "DEBUG ") && !DEBUG {
		return
	}
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}

func fmttime(t uint64) string {
	tdays, tsecs := t/(24*3600), t%(24*3600)
	ts := fmt.Sprintf("%ds", tsecs)
	if tdays > 0 {
		ts = fmt.Sprintf("%dd"+SEP, tdays) + ts
	}
	return ts
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

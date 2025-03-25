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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	ps "github.com/mitchellh/go-ps"
)

type Process struct {
	Pid  int64
	Ppid int64
	Pids []int64
	Name string
}

type Filter struct {
	Pid  int64
	Name string
}

var (
	Version string

	pp []Process
	ff []Filter
)

func init() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(Version)
		os.Exit(0)
	}
}

func main() {
	var err error

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

	pp0, err := ps.Processes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	for _, p0 := range pp0 {
		pid := p0.Pid()
		ppid := p0.PPid()
		name := p0.Executable()
		if runtime.GOOS == "linux" {
			namebb, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", p0.Pid()))
			if err == nil && len(namebb) > 0 {
				namebb = bytes.ReplaceAll(namebb, []byte{0}, []byte(" "))
				namebb = bytes.ReplaceAll(namebb, []byte("\n"), []byte(" "))
				name = string(namebb)
			}
		}
		p := Process{Pid: int64(pid), Ppid: int64(ppid), Name: name}
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

		pidss := ""
		for _, pid := range p.Pids {
			pidss += fmt.Sprintf("%d\t", pid)
		}
		if skip {
			continue
		}
		fmt.Printf("%s%s\n", pidss, p.Name)
	}
}

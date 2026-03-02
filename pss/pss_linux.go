//go:build linux

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetBootTime() (time.Time, error) {
	var boottime time.Time
	psbb, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}, err
	}
	for _, line := range strings.Split(string(psbb), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "btime" {
			btime, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return time.Time{}, err
			}
			boottime = time.Unix(btime, 0).UTC()
		}
	}
	if boottime.IsZero() {
		return time.Time{}, fmt.Errorf("/proc/stat btime not found")
	}
	return boottime, nil
}

func GetProcesses() ([]Process, error) {
	procdir, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer procdir.Close()
	ff, err := procdir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	pp := make([]Process, 0, 1000)
	for _, f := range ff {
		if !f.IsDir() {
			continue
		}
		pid, err := strconv.ParseInt(f.Name(), 10, 0)
		if err != nil {
			continue
		}

		p := Process{Pid: pid}

		pstatpath := fmt.Sprintf("/proc/%d/stat", p.Pid)
		pstatbb, err := ioutil.ReadFile(pstatpath)
		if err != nil {
			perr("ERROR read %s %v", pstatpath, err)
			continue
		}

		// First, parse out the image name
		pstats := string(pstatbb)
		pcommstart := strings.IndexByte(pstats, '(')
		pcommend := strings.LastIndexByte(pstats, ')')
		p.Name = pstats[pcommstart+1 : pcommend]

		// Move past the image name and start parsing the rest
		pstats = pstats[pcommend+2:]
		var skip int64
		// https://pkg.go.dev/fmt#Sscanf
		_, err = fmt.Sscanf(pstats,
			"%c %d %d %d "+
				"%d %d "+
				"%d %d %d %d %d "+
				"%d %d %d %d "+
				"%d %d %d %d "+
				"%d "+
				"%d %d",
			&p.State, &p.Ppid, &p.Pgid, &p.Sid,
			&p.TtyNr, &p.Tpgid,
			&skip, &skip, &skip, &skip, &skip,
			&p.utimeticks, &p.stimeticks,
			&skip, &skip,
			&skip, &skip, &skip, &skip,
			&p.starttimeticks,
			&p.Vsize, &p.Rss,
		)
		if err != nil {
			return nil, err
		}

		p.Utime = time.Duration(p.utimeticks) * time.Second / time.Duration(ClkTck)
		p.Stime = time.Duration(p.stimeticks) * time.Second / time.Duration(ClkTck)
		p.Starttime = BootTime.Add(time.Duration(p.starttimeticks/uint64(ClkTck)) * time.Second)

		cmdlinepath := fmt.Sprintf("/proc/%d/cmdline", p.Pid)
		cmdlinebb, err := ioutil.ReadFile(cmdlinepath)
		if err != nil {
			perr("ERROR read %s %v", cmdlinepath, err)
			continue
		}
		cmdlinebbb := bytes.Split(cmdlinebb, []byte{0})
		for len(cmdlinebbb) > 0 && len(cmdlinebbb[len(cmdlinebbb)-1]) == 0 {
			cmdlinebbb = cmdlinebbb[:len(cmdlinebbb)-1]
		}
		for _, a := range cmdlinebbb {
			p.Cmdline = append(p.Cmdline, string(a))
		}

		cgrouppath := fmt.Sprintf("/proc/%d/cgroup", p.Pid)
		cgroupbb, err := ioutil.ReadFile(cgrouppath)
		if err != nil {
			perr("ERROR read %s %v", cgrouppath, err)
			continue
		}
		p.Cgroup = string(cgroupbb)
		p.Kubepod = strings.Contains(p.Cgroup, "/kubepods/")

		pp = append(pp, p)
	}

	return pp, nil
}

/*
history:
20/1030 v1
20/1106 suffix every line with shortened hostname
23/0827 github.com/shirou/gopsutil/v3
*/

// GoGet GoFmt GoBuildNull
// GoBuild GoRun

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	pscpu "github.com/shirou/gopsutil/v4/cpu"
	psdisk "github.com/shirou/gopsutil/v4/disk"
	pshost "github.com/shirou/gopsutil/v4/host"
	psmem "github.com/shirou/gopsutil/v4/mem"
)

const (
	NL  = "\n"
	TAB = "\t"

	VisualRatio    = 5
	HostnameMaxLen = 32
)

var (
	VERSION string

	Hostname     string
	PollInterval time.Duration
	TimeLimit    time.Duration
)

func tsnow() string {
	t := time.Now().Local()
	return fmt.Sprintf(
		"%03d:%02d%02d:%02d%02d",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
}

func perr(msg string, args ...interface{}) {
	ts := tsnow()
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, ts+" "+msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, ts+" "+msg+NL, args...)
	}
}

func print() {
	ts := tsnow()

	cpuInterval := PollInterval
	if cpuInterval == 0 {
		cpuInterval = time.Second / 10
	}
	cpupercents, err := pscpu.Percent(cpuInterval, false)
	if err != nil {
		perr("pscpu.Percent %v", err)
		os.Exit(1)
	}
	cpupercent := int(cpupercents[0])
	cpugauge := (strings.Repeat("=", cpupercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-cpupercent/VisualRatio))
	cpunumber, err := pscpu.Counts(false)
	if err != nil {
		perr("pscpu.Counts %v", err)
		os.Exit(1)
	}

	mem, err := psmem.VirtualMemory()
	if err != nil {
		perr("psmem.VirtualMemory %v", err)
		os.Exit(1)
	}
	memsizegb := int(mem.Total / (1 << 30))
	mempercent := int(mem.UsedPercent)
	memgauge := (strings.Repeat("=", mempercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-mempercent/VisualRatio))

	swap, err := psmem.SwapMemory()
	if err != nil {
		perr("psmem.SwapMemory %v", err)
		os.Exit(1)
	}
	swapsizegb := int(swap.Total / (1 << 30))
	swappercent := int(swap.UsedPercent)
	var swapgauge string
	if swap.Total > 0 {
		swapgauge = (strings.Repeat("=", swappercent/VisualRatio) +
			strings.Repeat("-", 100/VisualRatio-swappercent/VisualRatio))
	} else {
		swapgauge = strings.Repeat(" ", 100/VisualRatio)
	}

	disk, err := psdisk.Usage("/")
	if err != nil {
		perr("psdisk.Usage %v", err)
		os.Exit(1)
	}
	disksizegb := int(disk.Total / (1 << 30))
	diskpercent := int(disk.UsedPercent)
	diskgauge := (strings.Repeat("=", diskpercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-diskpercent/VisualRatio))

	uptime, err := pshost.Uptime()
	if err != nil {
		perr("pshost.Uptime %v", err)
		os.Exit(1)
	}
	uptimedays, uptimesecs := uptime/(24*3600), uptime%(24*3600)
	uptimefmt := fmt.Sprintf("%ds", uptimesecs)
	if uptimedays > 0 {
		uptimefmt = fmt.Sprintf("%dd"+"·", uptimedays) + uptimefmt
	}

	fmt.Printf(
		"%s %s"+TAB+"cpu%s%d mem%s%dgb swap%s%dgb disk%s%dgb uptime<%s>"+NL,
		ts, Hostname,
		cpugauge, cpunumber,
		memgauge, memsizegb,
		swapgauge, swapsizegb,
		diskgauge, disksizegb,
		uptimefmt,
	)
}

func init() {
	if len(os.Args) == 2 && (os.Args[1] == "-version" || os.Args[1] == "version") {
		fmt.Println(VERSION)
		os.Exit(0)
	}
}

func main() {
	var err error

	Hostname, err = os.Hostname()
	if err != nil {
		perr("Hostname %v", err)
		os.Exit(1)
	}
	//Hostname = strings.TrimSuffix(Hostname, ".local")
	if len(Hostname) > HostnameMaxLen {
		Hostname = Hostname[:HostnameMaxLen-7] + "~" + Hostname[len(Hostname)-6:]
	}

	if len(os.Args) > 1 {
		ri, err := strconv.Atoi(os.Args[1])
		if err != nil {
			perr("invalid integer [%s] for repeat interval in seconds", os.Args[1])
			os.Exit(1)
		}
		PollInterval = time.Duration(ri) * time.Second

		if len(os.Args) > 2 {
			tl, err := strconv.Atoi(os.Args[2])
			if err != nil {
				perr("invalid integer [%s] for time limit in seconds", os.Args[2])
				os.Exit(1)
			}
			TimeLimit = time.Duration(tl) * time.Second
		}
	}

	if PollInterval > 0 {
		st := time.Now()
		for {
			print()
			time.Sleep(PollInterval)
			if TimeLimit > 0 && time.Since(st) > TimeLimit {
				break
			}
		}
	} else {
		print()
	}
}

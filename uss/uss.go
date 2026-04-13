/*
history:
20/1030 v1
20/1106 suffix every line with shortened hostname
23/0827 github.com/shirou/gopsutil/v3
*/

// GoGet GoFmt GoBuildNull GoBuild GoRun

package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	// https://pkg.go.dev/github.com/shirou/gopsutil/v4/
	pscpu "github.com/shirou/gopsutil/v4/cpu"
	psdisk "github.com/shirou/gopsutil/v4/disk"
	pshost "github.com/shirou/gopsutil/v4/host"
	psmem "github.com/shirou/gopsutil/v4/mem"
	psnet "github.com/shirou/gopsutil/v4/net"
	psproc "github.com/shirou/gopsutil/v4/process"
	slices "golang.org/x/exp/slices"
)

const (
	N   = ""
	SP  = " "
	TAB = "\t"
	NL  = "\n"
	SEP = ","

	VisualRatio    = 8
	HostnameMaxLen = 32
)

var (
	VERSION string

	Hostname     string
	PollInterval time.Duration
	TimeLimit    time.Duration
)

func print() {
	ts := fmttime(time.Now().Local())

	cpuInterval := PollInterval
	if cpuInterval == 0 {
		cpuInterval = time.Second / 2
	}
	cpupercents, err := pscpu.Percent(cpuInterval, false)
	if err != nil {
		perr("ERROR pscpu.Percent %v", err)
		os.Exit(1)
	}
	cpupercent := int(cpupercents[0])
	cpugauge := (strings.Repeat("=", cpupercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-cpupercent/VisualRatio))
	cpunumber, err := pscpu.Counts(false)
	if err != nil {
		perr("ERROR pscpu.Counts %v", err)
		os.Exit(1)
	}

	cpufreq_path := "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_cur_freq"
	cpufreqbb, err := os.ReadFile(cpufreq_path)
	if err != nil {
		//perr("WARNING ReadFile [%s] %v", cpufreq_path, err)
	}
	cpufreq := strings.TrimSpace(string(cpufreqbb))
	// https://pkg.go.dev/strconv#ParseUint
	cpufreqhz, err := strconv.ParseUint(cpufreq, 10, 64)
	if err != nil {
		//perr("WARNING ParseUint [%s] %v", cpufreqs, err)
	}
	if cpufreqhz > 0 {
		cpufreq = fmt.Sprintf("%d", cpufreqhz/1000) + "mhz"
	}
	if cpufreq != "" {
		cpufreq = "<" + cpufreq + ">"
	}

	mem, err := psmem.VirtualMemory()
	if err != nil {
		perr("ERROR psmem.VirtualMemory %v", err)
		os.Exit(1)
	}
	memsizemb := mem.Total / (1 << 20)
	mempercent := int(mem.UsedPercent)
	memgauge := (strings.Repeat("=", mempercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-mempercent/VisualRatio))

	swap, err := psmem.SwapMemory()
	if err != nil {
		perr("ERROR psmem.SwapMemory %v", err)
		os.Exit(1)
	}
	swapsizemb := swap.Total / (1 << 20)
	swappercent := int(swap.UsedPercent)
	var swapgauge string
	if swap.Total > 0 {
		swapgauge = (strings.Repeat("=", swappercent/VisualRatio) +
			strings.Repeat("-", 100/VisualRatio-swappercent/VisualRatio))
	} else {
		swapgauge = strings.Repeat("_", 100/VisualRatio)
	}

	disk, err := psdisk.Usage("/")
	if err != nil {
		perr("ERROR psdisk.Usage %v", err)
		os.Exit(1)
	}
	disksizegb := int(disk.Total / (1 << 30))
	diskpercent := int(disk.UsedPercent)
	diskgauge := (strings.Repeat("=", diskpercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-diskpercent/VisualRatio))

	var diskrdt, diskwrt uint64
	diskstats, err := psdisk.IOCounters()
	if err != nil {
		perr("ERROR psdisk.IOCounters %v", err)
		os.Exit(1)
	}
	for _, dss := range diskstats {
		diskrdt += dss.ReadTime
		diskwrt += dss.WriteTime
	}

	/*
		users, err := pshost.Users()
		if err != nil {
			perr("ERROR pshost.Users %v", err)
			//os.Exit(1)
		}
		//perr("DEBUG users %+v", users)
	*/

	procs, err := psproc.Processes()
	if err != nil {
		perr("ERROR psproc.Processes %v", err)
		os.Exit(1)
	}

	var listens []string
	inet4conns, err := psnet.Connections("inet4")
	if err != nil {
		perr("ERROR psnet.Connections inet4 %v", err)
		os.Exit(1)
	}
	//perr("inet4conns<%d>", len(inet4conns))
	inet6conns, err := psnet.Connections("inet6")
	if err != nil {
		perr("ERROR psnet.Connections inet6 %v", err)
		os.Exit(1)
	}
	//perr("inet6conns<%d>", len(inet6conns))
	for _, c := range append(inet4conns, inet6conns...) {
		if c.Status != "LISTEN" {
			continue
		}
		claddrip := c.Laddr.IP
		if strings.HasPrefix(claddrip, "10.") || strings.HasPrefix(claddrip, "127.") || strings.HasPrefix(claddrip, "192.168.") {
			continue
		}
		if claddrip == "::1" {
			continue
		}
		if claddrip == "0.0.0.0" || claddrip == "::" {
			claddrip = "*"
		}
		p, err := psproc.NewProcess(c.Pid)
		if err != nil {
			perr("ERROR psproc.NewProcess <%d> %v", c.Pid, err)
		}
		pname, err := p.Name()
		if err != nil {
			perr("ERROR p.Name <%d> %v", p.Pid, err)
		}
		cdesc := fmt.Sprintf("[%s:%s:%d]", pname, claddrip, c.Laddr.Port)
		if !slices.Contains(listens, cdesc) {
			listens = append(listens, cdesc)
		}
	}

	uptime, err := pshost.Uptime()
	if err != nil {
		perr("ERROR pshost.Uptime %v", err)
		os.Exit(1)
	}
	uptimefmt := fmtdursec(uptime)

	boot_id_path := "/proc/sys/kernel/random/boot_id"
	bootidbb, err := os.ReadFile(boot_id_path)
	if err != nil {
		//perr("WARNING ReadFile [%s] %v", boot_id_path, err)
	}
	bootid := string(bootidbb)
	if len(bootid) > 4 {
		bootid = bootid[:4]
	}

	pout(
		"<%s> [%s] cpu%s<%d>%s mem%s<%smb> swap%s<%smb> disk%s<%dgb> uptime<%s> bootid[%s] read<%s> write<%s> nprocs<%s> listens(%s)",
		ts, Hostname,
		cpugauge, cpunumber, cpufreq,
		memgauge, seps(memsizemb, 3),
		swapgauge, seps(swapsizemb, 3),
		diskgauge, disksizegb,
		uptimefmt, bootid,
		fmtdursec(diskrdt/1000), fmtdursec(diskwrt/1000),
		seps(uint64(len(procs)), 3),
		strings.Join(listens, N),
	)
}

func init() {
}

func main() {
	var err error

	args := os.Args[1:]
	//perr("DEBUG args %#v", args)
	n := 0
	for _, a := range args {
		if a != "" {
			args[n] = a
			n++
		}
	}
	args = args[:n]
	//perr("DEBUG n <%d> args %#v", n, args)

	if len(args) == 1 && args[0] == "version" {
		fmt.Print(VERSION + NL)
		os.Exit(0)
	}

	if len(args) > 0 {
		ri, err := strconv.Atoi(args[0])
		if err != nil {
			perr("ERROR invalid integer [%s] for repeat interval in seconds", args[0])
			os.Exit(1)
		}
		PollInterval = time.Duration(ri) * time.Second
	}

	if len(args) > 1 {
		tl, err := strconv.Atoi(args[1])
		if err != nil {
			perr("ERROR invalid integer [%s] for time limit in seconds", args[1])
			os.Exit(1)
		}
		TimeLimit = time.Duration(tl) * time.Second
	}

	Hostname, err = os.Hostname()
	if err != nil {
		perr("ERROR Hostname %v", err)
		os.Exit(1)
	}
	//Hostname = strings.TrimSuffix(Hostname, ".local")
	if len(Hostname) > HostnameMaxLen {
		Hostname = Hostname[:HostnameMaxLen-7] + "~" + Hostname[len(Hostname)-6:]
	}

	if PollInterval > 0 {
		st := time.Now()
		for {
			sti := time.Now()
			print()
			sli := PollInterval - time.Since(sti)
			if sli > 0 {
				time.Sleep(sli)
			}
			if TimeLimit > 0 && time.Since(st) > TimeLimit {
				break
			}
		}
	} else {
		print()
	}
}

func fmttime(t time.Time) string {
	return fmt.Sprintf(
		"%d:%02d%02d:%02d%02d",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
}

func fmtdursec(t uint64) string {
	tdays, tsecs := t/(24*3600), t%(24*3600)
	ts := seps(tsecs, 2) + "s"
	if tdays > 0 {
		ts = seps(tdays, 2) + "d" + SEP + ts
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

func perr(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}

func pout(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stdout, msg+NL)
	} else {
		fmt.Fprintf(os.Stdout, msg+NL, args...)
	}
}

/*
history:
2019/11/1 v2

GoGet GoFmt GoBuildNull GoBuild
GoRun
*/

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	psdisk "github.com/shirou/gopsutil/disk"
	pshost "github.com/shirou/gopsutil/host"
	psload "github.com/shirou/gopsutil/load"
	psmem "github.com/shirou/gopsutil/mem"
	psnet "github.com/shirou/gopsutil/net"
	psproc "github.com/shirou/gopsutil/process"
)

const (
	NL = "\n"
)

var (
	PRINTALL bool
)

func fmtdur(d time.Duration) (s string) {
	days := d / (time.Minute * 1440)
	mins := d % (time.Minute * 1440) / time.Minute
	s = fmt.Sprintf("%dm", mins)
	if days > 0 {
		s = fmt.Sprintf("%dd", days) + s
	}
	return s
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-a" || os.Args[1] == "-all" {
			PRINTALL = true
		}

	}

	loadavg, err := psload.Avg()
	if err != nil {
		plog("psload.Avg %s", err)
		os.Exit(1)
	}

	vmem, err := psmem.VirtualMemory()
	if err != nil {
		plog("psmem.VirtualMemory %s", err)
		os.Exit(1)
	}
	swapmem, err := psmem.SwapMemory()
	if err != nil {
		plog("psmem.SwapMemory %s", err)
		os.Exit(1)
	}

	diskstat, err := psdisk.Usage("/")
	if err != nil {
		plog("psdisk.Usage %s", err)
		os.Exit(1)
	}

	parts, err := psdisk.Partitions(true)
	if err != nil {
		plog("psdisk.Partitions %s", err)
		os.Exit(1)
	}
	rootpart := ""
	for _, p := range parts {
		if p.Mountpoint == "/" {
			rootpart = p.Device
			break
		}
	}
	if rootpart == "" {
		plog("Could not find root partition device name")
		os.Exit(1)
	}
	//pout("@rootpart %s"+NL, rootpart)

	diskcounts1map, err := psdisk.IOCounters(rootpart)
	if err != nil {
		plog("psdisk.IOCounters %s", err)
		os.Exit(1)
	}
	netcounts1map, err := psnet.IOCounters(false)
	if err != nil {
		plog("psnet.IOCounters %s", err)
		os.Exit(1)
	}

	time.Sleep(time.Second)

	diskcounts2map, err := psdisk.IOCounters(rootpart)
	if err != nil {
		plog("psdisk.IOCounters %s", err)
		os.Exit(1)
	}
	netcounts2map, err := psnet.IOCounters(false)
	if err != nil {
		plog("psnet.IOCounters %s", err)
		os.Exit(1)
	}

	var diskcounts1, diskcounts2 psdisk.IOCountersStat
	for _, c := range diskcounts1map {
		diskcounts1 = c
		break
	}
	for _, c := range diskcounts2map {
		diskcounts2 = c
		break
	}
	//pout("@diskcounts1map %+v @diskcounts1 %+v @diskcounts2map %+v @diskcounts2 %+v"+NL,
	//	diskcounts1map, diskcounts1, diskcounts2map, diskcounts2)
	diskread := diskcounts2.ReadBytes - diskcounts1.ReadBytes
	diskwrite := diskcounts2.WriteBytes - diskcounts1.WriteBytes

	ip4conns, err := psnet.Connections("inet4")
	if err != nil {
		plog("psnet.Connections %s", err)
		os.Exit(1)
	}
	ip6conns, err := psnet.Connections("inet6")
	if err != nil {
		plog("psnet.Connections %s", err)
		os.Exit(1)
	}

	var netcounts1, netcounts2 psnet.IOCountersStat
	for _, c := range netcounts1map {
		netcounts1 = c
		break
	}
	for _, c := range netcounts2map {
		netcounts2 = c
		break
	}
	netrecv := netcounts2.BytesRecv - netcounts1.BytesRecv
	netsent := netcounts2.BytesSent - netcounts1.BytesSent

	users, err := pshost.Users()
	if err != nil {
		plog("pshost.Users %s", err)
		os.Exit(1)
	}

	procs, err := psproc.Processes()
	if err != nil {
		plog("psproc.Processes %s", err)
		os.Exit(1)
	}

	boottimeunix, err := pshost.BootTime()
	if err != nil {
		plog("pshost.BootTime %s", err)
		os.Exit(1)
	}
	boottime := time.Unix(int64(boottimeunix), 0)

	/*

		pout("@cpu1m <%.0f%%> @cpu15m <%.0f%%> @mem <%.0f%%> @swap <%.0f%%>"+NL,
			loadavg.Load1*100, loadavg.Load15*100, vmem.UsedPercent, swapmem.UsedPercent)
		pout("@disk <%.0f%%> @diskread <%dkbps> @diskwrite <%dkbps>"+NL,
			diskstat.UsedPercent, diskread>>10, diskwrite>>10)
		pout("@ip4conns <%d> @ip6conns <%d> @netrecv <%dkbps> @netsent <%dkbps>"+NL,
			len(ip4conns), len(ip6conns), netrecv>>10, netsent>>10)
		pout("@users <%d> @procs <%d> @boot <%s>"+NL,
			len(users), len(procs), boottime.Format("Jan/2"))

	*/

	if PRINTALL || loadavg.Load1*100 > 100 {
		pout("@cpu1m <%.0f%%> ", loadavg.Load1*100)
	}
	if PRINTALL || loadavg.Load15*100 > 80 {
		pout("@cpu15m <%.0f%%> ", loadavg.Load15*100)
	}
	if PRINTALL || vmem.UsedPercent > 60 {
		pout("@mem <%.0f%%> ", vmem.UsedPercent)
	}
	if PRINTALL || swapmem.UsedPercent > 60 {
		pout("@swap <%.0f%%> ", swapmem.UsedPercent)
	}

	if PRINTALL || diskstat.UsedPercent > 60 {
		pout("@disk <%.0f%%> ", diskstat.UsedPercent)
	}
	if PRINTALL || diskread>>10 > 1000 {
		pout("@diskread <%dkbps> ", diskread>>10)
	}
	if PRINTALL || diskwrite>>10 > 1000 {
		pout("@diskwrite <%dkbps> ", diskwrite>>10)
	}

	if PRINTALL || len(ip4conns) > 20 {
		pout("@ip4conns <%d> ", len(ip4conns))
	}
	if PRINTALL || len(ip6conns) > 20 {
		pout("@ip6conns <%d> ", len(ip6conns))
	}
	if PRINTALL || netrecv>>10 > 1000 {
		pout("@netrecv <%dkbps> ", netrecv>>10)
	}
	if PRINTALL || netsent>>10 > 1000 {
		pout("@netsent <%dkbps> ", netsent>>10)
	}

	if PRINTALL || len(procs) > 150 {
		pout("@procs <%d> ", len(procs))
	}

	uptime := time.Since(boottime).Truncate(time.Minute)
	if PRINTALL || uptime < 100*time.Minute {
		pout("@uptime <%s> ", fmtdur(uptime))
	}

	pout(NL)

	for _, u := range users {
		ustarted := time.Unix(int64(u.Started), 0)
		pout("- @user %s @host %s @duration <%s>"+NL, u.User, u.Host,
			fmtdur(time.Since(ustarted)),
		)
	}

	for _, p := range procs {
		pcreatetime, err := p.CreateTime()
		if err != nil {
			plog("p.CreateTime %s", err)
			os.Exit(1)
		}
		puptime := time.Since(time.Unix(pcreatetime/1000, 0))
		pcpu, err := p.CPUPercent()
		if err != nil {
			plog("p.CPUPercent %s", err)
			os.Exit(1)
		}
		pmem, err := p.MemoryPercent()
		if err != nil {
			plog("p.MemoryPercent %s", err)
			os.Exit(1)
		}
		pname, err := p.Name()
		if err != nil {
			plog("p.Name %s", err)
			os.Exit(1)
		}
		pfiles, err := p.OpenFiles()
		if err != nil {
			plog("p.OpenFiles %s", err)
			os.Exit(1)
		}

		pconns, err := p.Connections()
		if err != nil {
			plog("p.Connections %s", err)
			os.Exit(1)
		}

		var plistens []string
		for _, c := range pconns {
			if pname == "docker-proxy" {
				continue
			}
			if c.Status != "LISTEN" {
				continue
			}
			claddr := c.Laddr.IP
			if strings.HasPrefix(claddr, "127.0.0.") || claddr == "::1" {
				continue
			}
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
			plistens = append(plistens, fmt.Sprintf("%s/%s", claddr, craddr))
		}

		if puptime > 10*time.Minute {
			continue
		}
		if pcpu < 20 && pmem < 20 && len(pfiles) < 100 && len(pconns) < 100 && len(plistens) == 0 {
			continue
		}

		pout("- @proc %s @uptime <%s> @cpu <%.0f%%> @mem <%.0f%%> @files <%d> @conns <%d> @listens (%s)"+NL,
			pname, fmtdur(puptime), pcpu, pmem, len(pfiles), len(pconns), strings.Join(plistens, " "),
		)
	}

}

func pout(text string, args ...interface{}) (int, error) {
	return fmt.Printf(text, args...)
}

func plog(text string, args ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, text, args...)
}

/*
history:
2019/11/1 v2

GoFmt GoBuildNull GoBuild GoRun
*/

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	psdisk "github.com/shirou/gopsutil/disk"
	pshost "github.com/shirou/gopsutil/host"
	psload "github.com/shirou/gopsutil/load"
	psmem "github.com/shirou/gopsutil/mem"
	psnet "github.com/shirou/gopsutil/net"
	psproc "github.com/shirou/gopsutil/process"
)

const NL = "\n"

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

func main() {
	loadavg, err := psload.Avg()
	if err != nil {
		log.Fatalf("psload.Avg: %s", err)
	}

	vmem, err := psmem.VirtualMemory()
	if err != nil {
		log.Fatalf("psmem.VirtualMemory: %s", err)
	}
	swapmem, err := psmem.SwapMemory()
	if err != nil {
		log.Fatalf("psmem.SwapMemory: %s", err)
	}

	diskstat, err := psdisk.Usage("/")
	if err != nil {
		log.Fatalf("psdisk.Usage: %s", err)
	}

	parts, err := psdisk.Partitions(true)
	if err != nil {
		log.Fatalf("psdisk.Partitions: %s", err)
	}
	rootpart := ""
	for _, p := range parts {
		if p.Mountpoint == "/" {
			rootpart = p.Device
			break
		}
	}
	if rootpart == "" {
		log.Fatal("Could not find root partition device name")
	}
	//fmt.Printf("rootpart=%v\n", rootpart)

	diskcounts1map, err := psdisk.IOCounters(rootpart)
	if err != nil {
		log.Fatalf("psdisk.IOCounters: %s", err)
	}
	netcounts1map, err := psnet.IOCounters(false)
	if err != nil {
		log.Fatalf("psnet.IOCounters: %s", err)
	}

	time.Sleep(time.Second)

	diskcounts2map, err := psdisk.IOCounters(rootpart)
	if err != nil {
		log.Fatalf("psdisk.IOCounters: %s", err)
	}
	netcounts2map, err := psnet.IOCounters(false)
	if err != nil {
		log.Fatalf("psnet.IOCounters: %s", err)
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
	//fmt.Printf("diskcounts1map=%+v diskcounts1=%+v diskcounts2map=%+v diskcounts2=%+v\n",
	//	diskcounts1map, diskcounts1, diskcounts2map, diskcounts2)
	diskread := diskcounts2.ReadBytes - diskcounts1.ReadBytes
	diskwrite := diskcounts2.WriteBytes - diskcounts1.WriteBytes

	ip4conns, err := psnet.Connections("inet4")
	if err != nil {
		log.Fatalf("psnet.Connections: %s", err)
	}
	ip6conns, err := psnet.Connections("inet6")
	if err != nil {
		log.Fatalf("psnet.Connections: %s", err)
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
		log.Fatalf("pshost.Users: %s", err)
	}

	procs, err := psproc.Processes()
	if err != nil {
		log.Fatalf("psproc.Processes: %s", err)
	}

	boottimeunix, err := pshost.BootTime()
	if err != nil {
		log.Fatalf("pshost.BootTime: %s", err)
	}
	boottime := time.Unix(int64(boottimeunix), 0)

	/*

		fmt.Printf("cpu1m=%.0f%% cpu15m=%.0f%% mem=%.0f%% swap=%.0f%%\n",
			loadavg.Load1*100, loadavg.Load15*100, vmem.UsedPercent, swapmem.UsedPercent)
		fmt.Printf("disk=%.0f%% diskread=%dkbps diskwrite=%dkbps\n",
			diskstat.UsedPercent, diskread>>10, diskwrite>>10)
		fmt.Printf("ip4conns=%d ip6conns=%d netrecv=%dkbps netsent=%dkbps\n",
			len(ip4conns), len(ip6conns), netrecv>>10, netsent>>10)
		fmt.Printf("users=%d procs=%d boot=%s\n",
			len(users), len(procs), boottime.Format("Jan/2"))

	*/

	if loadavg.Load1*100 > 100 {
		fmt.Printf("cpu1m:%.0f%% ", loadavg.Load1*100)
	}
	if loadavg.Load15*100 > 80 {
		fmt.Printf("cpu15m:%.0f%% ", loadavg.Load15*100)
	}
	if vmem.UsedPercent > 60 {
		fmt.Printf("mem:%.0f%% ", vmem.UsedPercent)
	}
	if swapmem.UsedPercent > 60 {
		fmt.Printf("swap:%.0f%% ", swapmem.UsedPercent)
	}

	if diskstat.UsedPercent > 60 {
		fmt.Printf("disk:%.0f%% ", diskstat.UsedPercent)
	}
	if diskread>>10 > 1000 {
		fmt.Printf("diskread:%dkbps ", diskread>>10)
	}
	if diskwrite>>10 > 1000 {
		fmt.Printf("diskwrite:%dkbps ", diskwrite>>10)
	}

	if len(ip4conns) > 20 {
		fmt.Printf("ip4conns:%d ", len(ip4conns))
	}
	if len(ip6conns) > 20 {
		fmt.Printf("ip6conns:%d ", len(ip6conns))
	}
	if netrecv>>10 > 1000 {
		fmt.Printf("netrecv:%dkbps ", netrecv>>10)
	}
	if netsent>>10 > 1000 {
		fmt.Printf("netsent:%dkbps ", netsent>>10)
	}

	if len(procs) > 150 {
		fmt.Printf("procs:%d ", len(procs))
	}

	uptime := time.Since(boottime).Truncate(time.Minute)
	if uptime < 100*time.Minute {
		fmt.Printf("uptime:%s ", fmtdur(uptime))
	}

	fmt.Printf("\n")

	for _, u := range users {
		ustarted := time.Unix(int64(u.Started), 0)
		fmt.Printf("- user:%s host:%s duration:%s"+NL, u.User, u.Host,
			fmtdur(time.Since(ustarted)),
		)
	}

	for _, p := range procs {
		pcreatetime, err := p.CreateTime()
		if err != nil {
			log.Fatalf("p.CreateTime: %s", err)
		}
		puptime := time.Since(time.Unix(pcreatetime/1000, 0))
		pcpu, err := p.CPUPercent()
		if err != nil {
			log.Fatalf("p.CPUPercent: %s", err)
		}
		pmem, err := p.MemoryPercent()
		if err != nil {
			log.Fatalf("p.MemoryPercent: %s", err)
		}
		pname, err := p.Name()
		if err != nil {
			log.Fatalf("p.Name: %s", err)
		}
		pfiles, err := p.OpenFiles()
		if err != nil {
			log.Fatalf("p.OpenFiles: %s", err)
		}

		pconns, err := p.Connections()
		if err != nil {
			log.Fatalf("p.Connections: %s", err)
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

		fmt.Printf("- proc:%s uptime:%s cpu:%.0f%% mem:%.0f%% files:%d conns:%d listens:%v"+NL,
			pname, fmtdur(puptime), pcpu, pmem, len(pfiles), len(pconns), plistens,
		)
	}

}

// history:
// 2015-1209 v1
// 2026-0117 aton
// go run server_info.go
// go fmt server_info.go
// GoGet GoFmt GoBuildNull GoBuild GoRun

package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	SP  = " "
	TAB = "	"
	NL  = "\n"
)

type ServerInfo struct {
	CPUInfo        string
	CPUCount       uint64
	HWCPUCount     uint64
	MemTotalBytes  uint64
	MemFreeBytes   uint64
	SwapTotalBytes uint64
	SwapFreeBytes  uint64
	Disks          []Disk
	PrivateIP      string
}
type Disk struct {
	Device     string
	FSType     string
	MountPoint string
	Opts       []string

	Total   uint64
	Free    uint64
	Used    uint64
	Percent float64
}

func main() {
	var serverInfo ServerInfo
	dpp, err := disk.Partitions(false)
	if err != nil {
		perr("ERROR disk.DiskPartitions %v", err)
		os.Exit(1)
	}
	for _, dp := range dpp {
		dps, err := disk.Usage(dp.Device)
		if err != nil {
			//perr("ERROR disk.DiskUsage %s %s", dp.Device, err)
			continue
		}
		var p Disk
		p.Device = dp.Device
		p.MountPoint = dp.Mountpoint
		p.FSType = dp.Fstype
		p.Opts = dp.Opts
		p.Total = dps.Total
		p.Free = dps.Free
		p.Used = dps.Used
		p.Percent = dps.UsedPercent
		serverInfo.Disks = append(serverInfo.Disks, p)
	}

	cpuCount, err := cpu.Counts(true)
	if err != nil {
		perr("ERROR cpu.CPUCounts true %v", err)
		os.Exit(1)
	}
	serverInfo.CPUCount = uint64(cpuCount)
	hwCpuCount, err := cpu.Counts(false)
	if err != nil {
		perr("cpu.CPUCounts false %v", err)
		os.Exit(1)
	}
	serverInfo.HWCPUCount = uint64(hwCpuCount)

	cpuInfo, err := cpu.Info()
	if err != nil {
		perr("ERROR cpu.CPUInfo %v", err)
		os.Exit(1)
	}
	if len(cpuInfo) == 0 {
		perr("ERROR cpu.CPUInfo empty result")
		os.Exit(1)
	}
	serverInfo.CPUInfo = cpuInfo[0].ModelName

	vm, err := mem.VirtualMemory()
	if err != nil {
		perr("ERROR mem.VirtualMemory %v", err)
		os.Exit(1)
	}
	serverInfo.MemTotalBytes = vm.Total
	//serverInfo.MemFreeBytes = vm.Available
	serverInfo.MemFreeBytes = vm.Free + vm.Buffers + vm.Cached

	sm, err := mem.SwapMemory()
	if err != nil {
		perr("ERROR mem.SwapMemory %v", err)
		os.Exit(1)
	}
	serverInfo.SwapTotalBytes = sm.Total
	serverInfo.SwapFreeBytes = sm.Free

	t := reflect.TypeOf(serverInfo)
	v := reflect.ValueOf(serverInfo)
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if ft.PkgPath != "" {
			continue
		}
		fv := v.Field(i)
		pout("@%s %+v"+NL, ft.Name, fv.Interface())
	}
}

func pout(text string, args ...interface{}) (int, error) {
	return fmt.Printf(text, args...)
}

func perr(text string, args ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, text+NL, args...)
}

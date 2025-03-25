// history: 2015-1209 v1
// go run server_info.go
// go fmt server_info.go

package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

const TAB = "	"

type ServerInfo struct {
	CPUInfo        string `json:"cpu_info"`
	CPUCount       uint64      `json:"cpu_count"`
	HWCPUCount     uint64      `json:"hw_cpu_count"`
	MemTotalBytes  uint64      `json:"mem_total_bytes"`
	MemFreeBytes   uint64      `json:"mem_free_bytes"`
	SwapTotalBytes uint64      `json:"swap_total_bytes"`
	SwapFreeBytes  uint64      `json:"swap_free_bytes"`
	Disks          []Disk `json:"disks"`
	PrivateIP       string `json:"private_ip"`
}
type Disk struct {
	Device string `json:"device"`
	FSType string `json:"fstype"`
	MountPoint string `json:"mountpoint"`
	Opts string `json:"opts"`

	Total uint64 `json:"total"`
	Free uint64 `json:"free"`
	Used uint64 `json:"used"`
	Percent float64 `json:"percent"`
}

func main() {
	var serverInfo ServerInfo
	dpp, err := disk.Partitions(false)
	if err != nil {
		log.Fatal("disk.DiskPartitions: ", err)
	}
	for _, dp := range dpp {
		dps, err := disk.Usage(dp.Device)
		if err != nil {
			//log.Printf("disk.DiskUsage `%s`: %s", dp.Device, err)
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
		log.Fatal("cpu.CPUCounts true: ", err)
	}
	serverInfo.CPUCount = uint64(cpuCount)
	hwCpuCount, err := cpu.Counts(false)
	if err != nil {
		log.Fatal("cpu.CPUCounts false: ", err)
	}
	serverInfo.HWCPUCount = uint64(hwCpuCount)

	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Fatal("cpu.CPUInfo: ", err)
	}
	if len(cpuInfo) == 0 {
		log.Fatal("cpu.CPUInfo: empty result")
	}
	serverInfo.CPUInfo = cpuInfo[0].ModelName

	vm, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal("mem.VirtualMemory: ", err)
	}
	serverInfo.MemTotalBytes = vm.Total
	//serverInfo.MemFreeBytes = vm.Available
	serverInfo.MemFreeBytes = vm.Free + vm.Buffers + vm.Cached

	sm, err := mem.SwapMemory()
	if err != nil {
		log.Fatal("mem.SwapMemory: ", err)
	}
	serverInfo.SwapTotalBytes = sm.Total
	serverInfo.SwapFreeBytes = sm.Free

	serverInfoEncoded, err := json.MarshalIndent(map[string]ServerInfo{"server_info": serverInfo}, "", TAB)
	if err != nil {
		log.Fatal("json.MarshalIndent: ", err)
	}
	os.Stdout.Write(serverInfoEncoded)
}

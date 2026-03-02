//go:build darwin

package main

import (
	"bytes"
	"encoding/binary"
	"syscall"
	"time"
	"unsafe"
)

func GetBootTime() (time.Time, error) {
	return time.Time{}, nil
	/*
		const _KERN_BOOTTIME     = 7
		mib := []int32{_CTRL_KERN, _KERN_BOOTTIME}
		var timeval struct {
			Sec  int64
			Usec int32
			_    int32 // padding
		}
		size := uintptr(unsafe.Sizeof(tv))
		_, _, errno := syscall.Syscall6(
			syscall.SYS___SYSCTL,
			uintptr(unsafe.Pointer(&mib[0])),
			uintptr(len(mib)),
			uintptr(unsafe.Pointer(&timeval)),
			uintptr(unsafe.Pointer(&size)),
			0,
			0,
		)
		if errno != 0 {
			return time.Time{}, fmt.Errorf("Syscall6 errno %v", errno)
		}
		return time.Unix(tv.Sec, int64(tv.Usec)*1000).UTC(), nil
	*/
}

const (
	_CTRL_KERN         = 1
	_KERN_PROC         = 14
	_KERN_PROC_ALL     = 0
	_KINFO_STRUCT_SIZE = 648
)

type kinfoProc struct {
	_         [8]byte
	StartSec  int64
	StartUsec int32
	_         [20]byte
	Pid       int32
	_         [160]byte
	Uticks    uint32
	Sticks    uint32
	_         [31]byte
	Comm      [16]byte
	_         [301]byte
	Ppid      int32
	_         [84]byte
}

func darwinSyscallProcAll() (*bytes.Buffer, error) {
	mib := [4]int32{_CTRL_KERN, _KERN_PROC, _KERN_PROC_ALL, 0}
	size := uintptr(0)
	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if errno != 0 {
		return nil, errno
	}
	bs := make([]byte, size)
	_, _, errno = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(&bs[0])),
		uintptr(unsafe.Pointer(&size)),
		0,
		0)
	if errno != 0 {
		return nil, errno
	}
	return bytes.NewBuffer(bs[0:size]), nil
}

func darwinCstring(bb [16]byte) string {
	for i := range bb {
		if bb[i] == 0 {
			return string(bb[:i])
		}
	}
	return string(bb[:])
}

func GetProcesses() ([]Process, error) {
	buf, err := darwinSyscallProcAll()
	if err != nil {
		return nil, err
	}
	kpp := make([]*kinfoProc, 0, 1000)
	k := 0
	for i := _KINFO_STRUCT_SIZE; i < buf.Len(); i += _KINFO_STRUCT_SIZE {
		kp := new(kinfoProc)
		err = binary.Read(bytes.NewBuffer(buf.Bytes()[k:i]), binary.LittleEndian, kp)
		if err != nil {
			return nil, err
		}
		k = i
		kpp = append(kpp, kp)
	}

	pp := make([]Process, len(kpp))
	for i, kp := range kpp {
		pgid, err := syscall.Getpgid(int(kp.Pid))
		if err != nil {
			return nil, err
		}
		sid, err := syscall.Getsid(int(kp.Pid))
		if err != nil {
			return nil, err
		}
		comm := darwinCstring(kp.Comm)
		vsize := 0
		rss := 0
		pp[i] = Process{
			Pid:       int64(kp.Pid),
			Ppid:      int64(kp.Ppid),
			Pgid:      int64(pgid),
			Sid:       int64(sid),
			Name:      comm,
			Cmdline:   []string{comm},
			Utime:     time.Duration(int64(kp.Uticks)/ClkTck) * time.Second,
			Stime:     time.Duration(int64(kp.Sticks)/ClkTck) * time.Second,
			Starttime: time.Unix(kp.StartSec, int64(kp.StartUsec)*1000),
			Vsize:     int64(vsize),
			Rss:       int64(rss),
		}
	}

	return pp, nil
}

// history: 2015-1210 v1
// go run svstat.go

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const (
	DOWN   = "down"
	UP     = "up"
	FINISH = "finish"

	NORMALLY_DOWN = "normally down"
	NORMALLY_UP   = "normally up"
	PAUSED        = "paused"
	WANT_UP       = "want up"
	WANT_DOWN     = "want down"
	GOT_TERM      = "got term"

	START = "u"
	PAUSE = "p"
	ALARM = "a"
	TERMINATE = "t"
	EXIT = "x"
	KILL = "k"
	USER1 = "1"
	USER2 = "2"
	QUIT = "q"
	INTERRUPT = "i"
	HANGUP = "h"
	CONT = "c"
	ONCE = "o"
	STOP = "d"

	EPOCH = 4611686018427387914
)

type Service struct {
	Path string

	Seconds int64
	Nano    int32
	PID     int32
	Paused  byte
	Want    byte
	Term    byte
	Finish  byte

	Status     string
	Action     string
	NormallyUp bool
}

func (s Service) ControlPath() string {
	return s.Path + "/supervise/control"
}

func (s Service) StatusPath() string {
	return s.Path + "/supervise/status"
}

func (s Service) DownPath() string {
	return s.Path + "/down"
}

func (s Service) WriteControl(c string) error {
	f, err := os.OpenFile(s.ControlPath(), os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return fmt.Errorf("os.OpenFile `%s`: %s", s.ControlPath(), err)
	}
	b := []byte(c)
	n, err := f.Write(b)
	if err != nil {
		f.Close()
		return fmt.Errorf("File.Write: %s", err)
	}
	if n != len(b) {
		f.Close()
		return fmt.Errorf("File.Write wrote %d bytes instead of %d bytes", n, len(b))
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("File.Close: %s", err)
	}
	return nil
}

func (s *Service) ReadStatus() error {
	if _, err := os.Stat(s.DownPath()); os.IsNotExist(err) {
		s.NormallyUp = true
	}

	b := make([]byte, 20)
	f, err := os.Open(s.StatusPath())
	if err != nil {
		return fmt.Errorf("os.Open `%s`: %s", s.StatusPath(), err)
	}
	n, err := f.Read(b)
	if err != nil {
		f.Close()
		return fmt.Errorf("File.Read: %s", err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("File.Close: %s", err)
	}

	//fmt.Fprintf(os.Stderr, "% x\n", b)
	if n < 18 {
		return fmt.Errorf("Service.Read returned %d bytes: %s", n, err)
	}
	s.Seconds = int64(binary.BigEndian.Uint64(b[0:8]))
	now := time.Now().UTC().Unix() + EPOCH
	if now < s.Seconds {
		s.Seconds = 0
	} else {
		s.Seconds = now - s.Seconds
	}
	s.Nano = int32(binary.BigEndian.Uint32(b[8:12]))
	s.PID = int32(binary.LittleEndian.Uint32(b[12:16]))
	s.Paused = b[16]
	s.Want = b[17]
	if n >= 20 {
		s.Term = b[18]
		s.Finish = b[19]
	}

	if s.PID > 0 {
		if s.Finish == 2 {
			s.Status = FINISH
		} else {
			s.Status = UP
		}
	} else {
		s.Status = DOWN
	}

	if s.PID > 0 && !s.NormallyUp {
		s.Action = NORMALLY_DOWN
	}
	if s.PID <= 0 && s.NormallyUp {
		s.Action = NORMALLY_UP
	}
	if s.PID > 0 && s.Paused > 0 {
		s.Action = PAUSED
	}
	if s.PID <= 0 && s.Want == 'u' {
		s.Action = WANT_UP
	}
	if s.PID > 0 && s.Want == 'd' {
		s.Action = WANT_DOWN
	}
	if s.PID > 0 && s.Term > 0 {
		s.Action = GOT_TERM
	}

	return nil
}

func main() {
	var sp string
	if len(os.Args) > 1 {
		sp = os.Args[1]
	} else {
		sp = "."
	}
	s := Service{Path: sp}
	err := s.ReadStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Service.ReadStatus: %s\n", err)
		os.Exit(1)
	}
	//fmt.Fprintf(os.Stderr, "%+v\n", s)
	//fmt.Printf("%s: %s (pid %d) %d seconds, %s\n", s.Path, s.Status, s.PID, s.Seconds, s.Action)
	fmt.Printf("Path=%s Status=%s PID=%d Seconds=%d Action=%s\n", s.Path, s.Status, s.PID, s.Seconds, s.Action)
}

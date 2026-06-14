/*
HISTORY
2015-1210 v1
*/

/*
GoFmt 
GoBuildNull 
GoBuild
GoRun stat /etc/service/*
*/

package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"time"
)

const (
	NL  = "\n"
	SEP = ","

	USAGE = `USAGE
	COMMANDS
		stat
		up
		down
		term
`

	STAT_DOWN   = "down"
	STAT_UP     = "up"
	STAT_FINISH = "finish"

	STAT_NORMALLY_DOWN = "normally-down"
	STAT_NORMALLY_UP   = "normally-up"
	STAT_PAUSED        = "paused"
	STAT_WANT_UP       = "want-up"
	STAT_WANT_DOWN     = "want-down"
	STAT_GOT_TERM      = "got-term"

	CTRL_UP     = "u"
	CTRL_DOWN      = "d"
	CTRL_ONCE      = "o"
	CTRL_PAUSE     = "p"
	CTRL_CONT      = "c"
	CTRL_HANGUP    = "h"
	CTRL_INT = "i"
	CTRL_TERM = "t"
	CTRL_ALARM = "a"
	CTRL_EXIT      = "x"
	CTRL_KILL      = "k"
	CTRL_USR1     = "1"
	CTRL_USR2     = "2"
	CTRL_QUIT      = "q"

	EPOCH = 4611686018427387914
)

var (
	F = fmt.Sprintf
	EF = fmt.Errorf
	pout = fmt.Print
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
		return EF("os.OpenFile %s %v", s.ControlPath(), err)
	}
	b := []byte(c)
	n, err := f.Write(b)
	if err != nil {
		f.Close()
		return EF("File.Write %v", err)
	}
	if n != len(b) {
		f.Close()
		return EF("File.Write wrote <%s> bytes instead of <%s> bytes", seps(uint64(n), 3), seps(uint64(len(b)), 3))
	}
	err = f.Close()
	if err != nil {
		return EF("File.Close %v", err)
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
		return EF("os.Open %s %s", s.StatusPath(), err)
	}
	n, err := f.Read(b)
	if err != nil {
		f.Close()
		return EF("File.Read %s", err)
	}
	err = f.Close()
	if err != nil {
		return EF("File.Close %s", err)
	}

	//perr("DEBUG % x", b)
	if n < 18 {
		return EF("Service.Read returned %d bytes %s", n, err)
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
			s.Status = STAT_FINISH
		} else {
			s.Status = STAT_UP
		}
	} else {
		s.Status = STAT_DOWN
	}

	if s.PID > 0 && !s.NormallyUp {
		s.Action = STAT_NORMALLY_DOWN
	}
	if s.PID <= 0 && s.NormallyUp {
		s.Action = STAT_NORMALLY_UP
	}
	if s.PID > 0 && s.Paused > 0 {
		s.Action = STAT_PAUSED
	}
	if s.PID <= 0 && s.Want == 'u' {
		s.Action = STAT_WANT_UP
	}
	if s.PID > 0 && s.Want == 'd' {
		s.Action = STAT_WANT_DOWN
	}
	if s.PID > 0 && s.Term > 0 {
		s.Action = STAT_GOT_TERM
	}

	return nil
}

func main() {
	
	var err error
	var args []string
	args = os.Args[1:]
	
	var cmd string
	var cmdargs []string
	if len(args) < 1 {
		perr(USAGE)
		os.Exit(1)
	}
	cmd = args[0]
	cmdargs = args[1:]
	
	switch cmd {
		case "stat":
			err = svstat(cmdargs)
		case "up":
			err = svup(cmdargs)
		case "down":
			err = svdown(cmdargs)
		case "once":
			err = svonce(cmdargs)
		case "pause":
			err = svpause(cmdargs)
		case "continue":
			err = svcontinue(cmdargs)
		case "hangup":
			err = svhangup(cmdargs)
		case "int":
			err = svint(cmdargs)
		case "term":
			err = svterm(cmdargs)
		case "kill":
			err = svkill(cmdargs)
		case "usr1":
			err = svusr1(cmdargs)
		case "usr2":
			err = svusr2(cmdargs)
		default:
			perr(F("ERROR unknown command [%s]", cmd))
			os.Exit(1)
	}

	if err != nil {
		perr(F("ERROR %s %v", cmd, err))
		os.Exit(1)
	}
	
	return
	
}

func svstat(args []string) error {
	var spp []string
	if len(args) > 0 {
		spp = args[:]
	} else {
		spp = []string{"."}
	}

	for _, sp := range spp {
		s := Service{Path: sp}
		err := s.ReadStatus()
		if err != nil {
			return EF("Service.ReadStatus %v", err)
		}

		durdays, dursecs := s.Seconds/(24*3600), s.Seconds%(24*3600)
		durfmt := F("%ss", seps(uint64(dursecs), 2))
		if durdays > 0 {
			durfmt = F("%sd"+SEP, seps(uint64(durdays), 2)) + durfmt
		}

		/*
			// classic format
			pout(F(
				"%s: %s pid<%d> <%ds>, %s",
				s.Path, s.Status, s.PID, s.Seconds, s.Action,
			)+NL)
		*/
		pout(F(
			"path[%s] status[%s] pid<%d> seconds<%s> action[%s]",
			s.Path, s.Status, s.PID, durfmt, s.Action,
		)+NL)
	}
	
	return nil
}

func svup(args []string) error {
	var spp []string
	if len(args) > 0 {
		spp = args[:]
	} else {
		spp = []string{"."}
	}
	
	for _, sp := range spp {
		s := Service{Path: sp}
		err := s.WriteControl(CTRL_UP)
		if err != nil {
			return EF("[%s] WriteControl %v", sp, err)
		}
	}
	
	return nil
}

func svdown(args []string) error {
	var spp []string
	if len(args) > 0 {
		spp = args[:]
	} else {
		spp = []string{"."}
	}
	
	for _, sp := range spp {
		s := Service{Path: sp}
		err := s.WriteControl(CTRL_DOWN)
		if err != nil {
			return EF("[%s] WriteControl %v", sp, err)
		}
	}
	
	return nil
}

func svonce(args []string) error {
	return EF("not implemented")
}

func svpause(args []string) error {
	return EF("not implemented")
}

func svcontinue(args []string) error {
	return EF("not implemented")
}

func svhangup(args []string) error {
	return EF("not implemented")
}

func svint(args []string) error {
	return EF("not implemented")
}

func svterm(args []string) error {
	var spp []string
	if len(args) > 0 {
		spp = args[:]
	} else {
		spp = []string{"."}
	}
	
	for _, sp := range spp {
		s := Service{Path: sp}
		err := s.WriteControl(CTRL_TERM)
		if err != nil {
			return EF("[%s] WriteControl %v", sp, err)
		}
	}
	
	return nil
}

func svkill(args []string) error {
	return EF("not implemented")
}

func svusr1(args []string) error {
	return EF("not implemented")
}

func svusr2(args []string) error {
	return EF("not implemented")
}

func perr(msg string) {
	fmt.Fprint(os.Stderr, msg+NL)
}

func seps(i uint64, e uint64) string {
	ee := uint64(math.Pow(10, float64(e)))
	if i < ee {
		return F("%d"+SEP, i)
	} else {
		return F("%s"+"%"+F("0%dd"+SEP, e), seps(i/ee, e), i%ee)
	}
}


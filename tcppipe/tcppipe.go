/*
history:
2016-0203 v1
*/

// GoFmt GoBuildNull GoBuild

package main

import (
	"expvar"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

const (
	USAGE = `tcppipe: creates pipe for a tcp address:port to another address:port
usage: tcppipe accept/dial addr1 accept/dial addr2
example: tcppipe accept 127.1:11465 dial 1.2.3.4:465
example: tcppipe dial 127.1:9022 dial 127.1:22
example: tcppipe accept 127.1:8022 accept 127.1:9022
env vars:
	Timeout [` + TimeoutStringDef + `] - timeout for tcp connections and between dials
`

	SP  = " "
	TAB = "\t"
	NL  = "\n"

	TimeoutStringDef = "30s"
)

var (
	DEBUG bool

	Timeout time.Duration

	IPFilter []string

	expAllow1 *expvar.Int
	expAllow2 *expvar.Int
	expOpen1  *expvar.Int
	expOpen2  *expvar.Int
	expClose1 *expvar.Int
	expClose2 *expvar.Int
	expAddr1  *expvar.Map
	expAddr2  *expvar.Map
)

func allowAccept(addr string) (allow chan bool, connch chan *net.Conn, err error) {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return
	}
	allow = make(chan bool)
	connch = make(chan *net.Conn)
	go func(allow chan bool, l net.Listener, connch chan *net.Conn) {
		for {
			<-allow
			l.(*net.TCPListener).SetDeadline(time.Now().Add(Timeout))
			conn, err := l.Accept()
			if err == nil {
				connch <- &conn
			} else {
				connch <- nil
			}
		}
	}(allow, l, connch)
	return
}

func allowDial(addr string) (allow chan bool, connch chan *net.Conn, err error) {
	allow = make(chan bool)
	connch = make(chan *net.Conn)
	go func(allow chan bool, addr string, connch chan *net.Conn) {
		for {
			<-allow
			conn, err := net.Dial("tcp4", addr)
			if err == nil {
				connch <- &conn
			} else {
				connch <- nil
			}
			time.Sleep(Timeout)
		}
	}(allow, addr, connch)
	return
}

func allowConn(cmd string, addr string) (allow chan bool, connch chan *net.Conn, err error) {
	switch cmd {
	case "accept":
		allow, connch, err = allowAccept(addr)
	case "dial":
		allow, connch, err = allowDial(addr)
	default:
		err = fmt.Errorf("cannon parse command [%s] should be [accept]/[dial]", cmd)
	}
	return
}

func main() {
	var err error

	if os.Getenv("DEBUG") != "" {
		DEBUG = true
	}

	TimeoutString := os.Getenv("Timeout")
	if TimeoutString == "" {
		TimeoutString = TimeoutStringDef
	}
	Timeout, err = time.ParseDuration(TimeoutString)
	if err != nil {
		perr("ERROR ParseDuration Timeout [%s] %v", TimeoutString, err)
		os.Exit(1)
	}
	perr("Timeout <%s>", Timeout)

	if v := os.Getenv("IPFilter"); v != "" {
		IPFilter = strings.Fields(v)
	}
	ipfilteraton := "("
	for _, i := range IPFilter {
		ipfilteraton += "[" + i + "]"
	}
	ipfilteraton += ")"
	perr("IPFilter %s", ipfilteraton)

	args := os.Args[1:]

	if len(args) != 4 {
		perr("ERROR invalid number of arguments")
		perr("args (%s)", strings.Join(args, SP))
		perr(NL + USAGE)
		os.Exit(1)
	}
	cmd1, addr1 := args[0], args[1]
	cmd2, addr2 := args[2], args[3]

	al1, ch1, err := allowConn(cmd1, addr1)
	if err != nil {
		perr("ERROR allowConn ([%s] [%s]) %v", cmd1, addr1, err)
		os.Exit(1)
	}

	al2, ch2, err := allowConn(cmd2, addr2)
	if err != nil {
		perr("ERROR allowConn ([%s] [%s]) %v", cmd2, addr2, err)
		os.Exit(1)
	}

	expAllow1 = expvar.NewInt("Allow1")
	expAllow2 = expvar.NewInt("Allow2")
	expOpen1 = expvar.NewInt("Open1")
	expOpen2 = expvar.NewInt("Open2")
	expClose1 = expvar.NewInt("Close1")
	expClose2 = expvar.NewInt("Close2")
	expAddr1 = expvar.NewMap("Accept1")
	expAddr2 = expvar.NewMap("Accept2")

	for {
		al1 <- true
		expAllow1.Add(1)
		conn1 := <-ch1
		if conn1 == nil {
			continue
		}
		localaddr := (*conn1).LocalAddr().String()
		remoteaddr := (*conn1).RemoteAddr().String()
		remoteip, _, err := net.SplitHostPort(remoteaddr)
		if err != nil {
			perr("WARNING SplitHostPort [%s] %v", remoteaddr, err)
		}
		if len(IPFilter) > 0 && !slices.Contains(IPFilter, remoteip) {
			perr("remote[%s] local[%s] -> filtered", remoteaddr, localaddr)
			continue
		}
		perr("remote[%s] local[%s] ->", remoteaddr, localaddr)
		expOpen1.Add(1)
		expAddr1.Add(remoteaddr, 1)

		go func(conn1 *net.Conn) {
			defer func() {
				(*conn1).Close()
				expClose1.Add(1)
			}()

			al2 <- true
			expAllow2.Add(1)
			conn2 := <-ch2
			if conn2 == nil {
				return
			}
			perr("remote[%s] local[%s] -> local[%s] remote[%s]", (*conn1).RemoteAddr(), (*conn1).LocalAddr(), (*conn2).LocalAddr(), (*conn2).RemoteAddr())
			expOpen2.Add(1)
			expAddr2.Add((*conn2).RemoteAddr().String(), 1)
			defer func() {
				(*conn2).Close()
				expClose2.Add(1)
			}()

			tconn1 := timeoutConn{*conn1}
			tconn2 := timeoutConn{*conn2}
			go io.Copy(*conn2, tconn1)
			io.Copy(*conn1, tconn2)
		}(conn1)
	}
}

type timeoutConn struct {
	Conn net.Conn
}

func (c timeoutConn) Read(buf []byte) (int, error) {
	c.Conn.SetReadDeadline(time.Now().Add(Timeout))
	return c.Conn.Read(buf)
}

func (c timeoutConn) Write(buf []byte) (int, error) {
	c.Conn.SetWriteDeadline(time.Now().Add(Timeout))
	return c.Conn.Write(buf)
}

func fmttime(t time.Time) string {
	return fmt.Sprintf(
		"%d:%02d%02d:%02d%02d",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
}

func perr(msg string, args ...interface{}) {
	if strings.HasPrefix(msg, "DEBUG ") && !DEBUG {
		return
	}
	tnow := time.Now().Local()
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "<"+fmttime(tnow)+">"+SP+msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, "<"+fmttime(tnow)+">"+SP+msg+NL, args...)
	}
}

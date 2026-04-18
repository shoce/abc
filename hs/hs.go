/*
Hs

history:
020/0605 v1
020/1016 repl
020/302 2020/10/28 stdin reading support
020/302 interrupt signal (ctrl+c) catching so only child processes get it
still not working with root sessions:
Oct 28 21:37:28 ci sshd[3685911]: error: session_signal_req: session signalling requires privilege separation
020/357 UserKeyFile support
021/0502 InReaderBufferSize
021/1117 SILENT
023/0827 VERBOSE
023/0827 keepalive
025/0108 sighup
025/0823 Status TermInverse
*/

// GoGet GoFmt GoBuildNull GoBuild
// GoRun -- put a '<' <readme.text
// Kill GoRun

/*

Variables:
Host variable checked before every command execution: if it is empty then run locally; otherwise run via ssh.
User variable stores user name if run via ssh.
Status variable tells exit status of the last command executed.

//
notes for possible future scripting language:
Reserved words:
ls, ll [path] / list directory of file by path
cd [dir] / change $dir
if X { one } else { two } / branch execution
for x, y := / loop
! / negate $status
~ regexp string / match string against regexp
exit [status] / exit shell with status
//

*/

package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

const (
	SP  = " "
	TAB = "\t"
	NL  = "\n"
	SEP = ","

	PortDefault = "22"

	InReaderBufferSize = 100 * 1000

	CmdHostname = `hostname -f`
	CmdBootTime = `cat /proc/stat`
	CmdBootId   = `cat /proc/sys/kernel/random/boot_id`
	CmdPwd      = `pwd`

	CmdAllPathCmds = `dd="" ; for d in ${PATH//:/ } ; do test -L "$d" && d=$(readlink -f "$d") ; test -d "$d" && dd="$dd $d" ; done ; ddd="" ; for d in $dd ; do for di in $ddd ; do test "$di" = "$d" && continue 2 ; done ; ddd="$ddd $d" ; done ; for d in $ddd ; do find "$d/" -maxdepth 1 -type f -executable -print -o -type l -exec sh -c 'test -x "{}"' \; -print | LC_ALL=C sort ; done ;`
	CmdAllFiles    = `find "%s" -maxdepth 1 -print | LC_ALL=C sort`
)

var (
	VERSION string

	LogBeatTime bool

	DEBUG   bool
	VERBOSE bool

	TERM string

	Proxy       string // proxy chain separated by semicolons
	ProxyChain  = []string{}
	ProxyDialer proxy.Dialer
	ProxyConn   net.Conn

	Host string // host network address to run commands on: empty or localhost to run with exec() and hostname[:port] to use ssh transport

	Hostname string
	BootTime int64
	BootId   string
	Pwd      string

	AllPathCmds []string
	AllFiles    []string

	SshKeepAliveInterval time.Duration = 12 * time.Second

	SshClientConfig *ssh.ClientConfig
	SshConn         *ssh.Conn
	SshClient       *ssh.Client

	User string // user name

	UserPassword   string
	UserKeyFile    string
	UserKey        string
	UserSigner     ssh.Signer
	UserAuthMethod ssh.AuthMethod

	Status string // status of the last run command

	InterruptChan chan bool

	TzBiel *time.Location = time.FixedZone("Biel", 60*60)

	F = fmt.Sprintf
)

func init() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if os.Getenv("DEBUG") != "" {
		DEBUG = true
	}
	if os.Getenv("VERBOSE") != "" {
		VERBOSE = true
	}

	if v := os.Getenv("TERM"); v != "" {
		TERM = v
	}

	var err error

	Proxy = os.Getenv("Proxy")
	perr("VERBOSE Proxy [%s]", Proxy)
	ProxyChain = strings.FieldsFunc(Proxy, func(c rune) bool { return c == ';' })
	perr("VERBOSE ProxyChain <%d> %v", len(ProxyChain), ProxyChain)
	ProxyDialer = proxy.Direct

	Host = os.Getenv("Host")
	perr("VERBOSE Host [%s]", Host)
	if Host == "" {
		perr("ERROR Host env var empty")
		os.Exit(1)
	}

	User = os.Getenv("User")
	perr("VERBOSE User [%s]", User)
	if User == "" {
		perr("ERROR User env var empty")
		os.Exit(1)
	}

	UserPassword = os.Getenv("UserPassword")
	if UserPassword != "" {
		UserAuthMethod = ssh.Password(UserPassword)
	}
	perr("VERBOSE UserPassword [%s]", UserPassword)

	if os.Getenv("home") == "" {
		err = os.Setenv("home", os.Getenv("HOME"))
		if err != nil {
			perr("WARNING Setenv home %v", err)
		}
	}

	UserKeyFile = os.ExpandEnv(os.Getenv("UserKeyFile"))
	if UserKeyFile != "" {
		userkeybb, err := ioutil.ReadFile(UserKeyFile)
		if err != nil {
			perr("ERROR Read UserKeyFile %v", err)
			os.Exit(1)
		}
		UserKey = string(userkeybb)
	}

	if os.Getenv("UserKey") != "" {
		UserKey = os.Getenv("UserKey")
	}

	if UserKey != "" {
		UserSigner, err = ssh.ParsePrivateKey([]byte(UserKey))
		if err != nil {
			perr("ERROR ParsePrivateKey %v", err)
			os.Exit(1)
		}
		UserAuthMethod = ssh.PublicKeys(UserSigner)
	}
	perr("VERBOSE UserKey [%s]", UserKey)

	if UserAuthMethod == nil {
		perr("ERROR no user auth method provided: no password and no user key")
		os.Exit(1)
	}

	SshClientConfig = &ssh.ClientConfig{
		User:    User,
		Auth:    []ssh.AuthMethod{UserAuthMethod},
		Timeout: 10 * time.Second,
		//ClientVersion: "hs", // NewClientConn: ssh: handshake failed: ssh: invalid packet length, packet too large
		//HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			perr("VERBOSE SSH server public key: type:%s hex:%s", key.Type(), hex.EncodeToString(key.Marshal()))
			return nil
		},
		BannerCallback: func(msg string) error {
			msg = strings.TrimSpace(msg)
			sep := SP
			if strings.Contains(msg, "\n") {
				sep = NL
			}
			perr("SSH server banner %s%s", sep, msg)
			return nil
		},
	}
}

func main() {
	var err error

	sigintchan := make(chan os.Signal, 1)
	signal.Notify(sigintchan, syscall.SIGINT)
	go func() {
		for {
			s := <-sigintchan
			switch s {
			case syscall.SIGINT:
				perr(NL)
				perr("interrupt signal")
				if InterruptChan != nil {
					InterruptChan <- true
				}
			}
		}
	}()

	sighupchan := make(chan os.Signal, 1)
	signal.Notify(sighupchan, syscall.SIGHUP)
	go func() {
		for {
			s := <-sighupchan
			switch s {
			case syscall.SIGHUP:
				perr(NL)
				perr("hangup signal")
				os.Exit(2)
			}
		}
	}()

	args := os.Args[1:]

	if Host == "" {

		Hostname, err = os.Hostname()
		if err != nil {
			perr("ERROR Hostname %v", err)
			os.Exit(1)
		}
		Hostname = strings.TrimSuffix(Hostname, ".local")
		//perr("Hostname [%s]", Hostname)

		u, err := user.Current()
		if err != nil {
			perr("WARNING user.Current %v", err)
		}
		User = u.Username

	} else {

		if len(ProxyChain) > 0 {
			for _, p := range ProxyChain {
				proxyurl, err := url.Parse(p)
				if err != nil {
					perr("ERROR proxy url [%s]` %v", p, err)
					os.Exit(1)
				}
				pd, err := proxy.FromURL(proxyurl, ProxyDialer)
				if err != nil {
					perr("ERROR proxy from url [%s] %v", p, err)
					os.Exit(1)
				}
				ProxyDialer = pd
			}
		}

		var addrerr *net.AddrError
		if _, _, err := net.SplitHostPort(Host); err != nil {
			if errors.As(err, &addrerr) && addrerr.Err == "missing port in address" {
				if len(Host) > 2 && Host[0] == '[' && Host[len(Host)-1] == ']' && net.ParseIP(Host[1:len(Host)-1]) != nil {
					Host = Host[1 : len(Host)-1]
				}
				if ip := net.ParseIP(Host); ip != nil {
					perr("DEBUG ip [%s]", ip.String())
					Host = net.JoinHostPort(ip.String(), PortDefault)
				} else {
					Host = net.JoinHostPort(Host, PortDefault)
				}
			} else {
				perr("ERROR SplitHostPort %v", err)
				os.Exit(1)
			}
		}

		perr("DEBUG Host [%s]", Host)
		err = connectssh()
		if err != nil {
			//perr("ERROR connect ssh %v", err)
		}
		if SshClient != nil {
			defer SshClient.Close()
		}
	}

	inreader := bufio.NewReaderSize(os.Stdin, InReaderBufferSize)

	if len(args) > 0 && args[0] != "--" {
		perr("ERROR the first argument should be `--`, example `hs -- id`")
		os.Exit(1)
	}

	if len(args) > 1 {
		cmd := args[1:]
		cmds := strings.Join(cmd, " ")

		if cmd[len(cmd)-1] == "<" {
			cmd = cmd[:len(cmd)-1]
			cmds = strings.Join(cmd, " ")
			perr("%s stdin: ", cmds)
		}

		Status, err = run(cmds, cmd, inreader)
		if err != nil {
			perr("host[%s] user[%s] [%s] ERROR %v", Host, User, cmds, err)
			os.Exit(1)
		}
		if Status != "" {
			perr("host[%s] user[%s] [%s] %s", Host, User, cmds, TermInverse(F("status[%s]", Status)))
		}
		os.Exit(0)
	}

	var stdinbb []byte
	for {
		logstatus()

		cmds, err := inreader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				perr("EOF")
				break
			}
			perr("WARNING ReadString %v", err)
			continue
		}

		cmds = strings.TrimSuffix(cmds, NL)
		perr("DEBUG cmds [%s]", cmds)
		if strings.HasSuffix(cmds, TAB) {
			perr(strings.ReplaceAll(cmds, TAB, "<TAB>"))
			// https://pkg.go.dev/strings#TrimPrefix
			cmds = strings.TrimSuffix(cmds, TAB)
			cmdsff := strings.Split(cmds, SP)
			if len(cmdsff) == 1 {
				cmd := cmdsff[0]
				if len(AllPathCmds) == 0 {
					err := GetAllPathCmds()
					if err != nil {
						//perr("ERROR AllPathCmdsGet %v", err)
					}
				}
				for _, c := range AllPathCmds {
					if strings.HasPrefix(c, cmd) || strings.HasPrefix(path.Base(c), cmd) {
						pout(c)
					}
				}
			} else if len(cmdsff) > 1 {
				fpath := cmdsff[len(cmdsff)-1]
				fpathdir := path.Dir(fpath)
				if fpathdir == "." {
					fpathdir = Pwd
				}
				if !strings.HasSuffix(fpathdir, "/") {
					fpathdir += "/"
				}
				err := GetAllFiles(fpathdir)
				if err != nil {
					//perr("ERROR AllFilesGet %v", err)
				}
				for _, f := range AllFiles {
					if strings.HasPrefix(f, fpath) || strings.HasPrefix(f, path.Join(Pwd, fpath)) {
						pout(f)
					}
				}
			}
			continue
		}

		//cmds = strings.TrimSpace(cmds)
		if cmds == "" {
			continue
		}

		cmd := strings.Split(cmds, " ")

		stdinbb = nil
		if cmd[len(cmd)-1] == "<" {
			cmd = cmd[:len(cmd)-1]
			cmds = cmds[:len(cmds)-1]
			perr("[%s] stdin:", cmds)
			stdinbb, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				perr("ERROR stdin read %v", err)
				continue
			}
		}

		Status, err = run(cmds, cmd, bytes.NewBuffer(stdinbb))
		if err != nil {
			perr("host[%s] user[%s] [%s] ERROR %v", Host, User, cmds, err)
			continue
		}
	}
}

func TermItalic(s string) string {
	if TERM != "" {
		return "\033[3m" + s + "\033[23m"
	}
	return s
}

func TermUnderline(s string) string {
	if TERM != "" {
		return "\033[4m" + s + "\033[24m"
	}
	return s
}

func TermInverse(s string) string {
	if TERM != "" {
		return "\033[7m" + s + "\033[27m"
	}
	return s
}

func TermUnderlineInverse(s string) string {
	if TERM != "" {
		return "\033[4;7m" + s + "\033[24;27m"
	}
	return s
}

func perr(msg string, args ...interface{}) {
	if strings.HasPrefix(msg, "DEBUG ") && !DEBUG {
		return
	}
	if strings.HasPrefix(msg, "VERBOSE ") && !VERBOSE {
		return
	}
	tnow := time.Now().Local()
	ts := ""
	if LogBeatTime {
		const BEAT = time.Duration(24) * time.Hour / 1000
		tnow = tnow.In(TzBiel)
		ty := tnow.Sub(time.Date(tnow.Year(), 1, 1, 0, 0, 0, 0, TzBiel))
		td := tnow.Sub(time.Date(tnow.Year(), tnow.Month(), tnow.Day(), 0, 0, 0, 0, TzBiel))
		ts = F(
			"<%03d:%d:%d>",
			tnow.Year()%1000,
			int(ty/(time.Duration(24)*time.Hour))+1,
			int(td/BEAT),
		)
	} else {
		ts = "<" + fmttime(tnow) + ">"
	}
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, ts+" "+msg+NL, args...)
	} else {
		fmt.Fprint(os.Stderr, ts+" "+msg+NL)
	}
}

func pout(msg string, args ...interface{}) {
	msgtext := msg
	if len(args) > 0 {
		msgtext = F(msg, args...)
	}
	fmt.Fprint(os.Stdout, msgtext+NL)

}

func fmttime(t time.Time) string {
	ts := F(
		"%d:%02d%02d:%02d%02d",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
	// https://pkg.go.dev/time#Time.Zone
	if _, tzoffset := t.Zone(); tzoffset == 0 {
		ts += "+"
	} else {
		ts += "-"
	}
	return ts
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
		return F("%d", i%ee)
	} else {
		f := F("0%dd", e)
		return F("%s"+SEP+"%"+f, seps(i/ee, e), i%ee)
	}
}

func logstatus() {
	fmt.Fprintf(os.Stderr, NL)
	s := F("status[%s]", Status)
	if Status != "" {
		s = TermInverse(s)
	}
	uptime := "nil"
	if BootTime > 0 {
		uptimesecs := uint64(time.Now().Unix() - BootTime)
		uptime = fmtdursec(uptimesecs)
	}
	s += F(
		" uptime<%s> bootid[%s] hostname[%s] host=%s user=%s hs -- ",
		uptime, BootId, Hostname, Host, User,
	)
	s = TermUnderline(s)
	perr(s)
}

func copynotify(dst io.Writer, src io.Reader, notify chan error) {
	_, err := io.Copy(dst, src)
	if notify != nil {
		notify <- err
	}
}

func connectssh() (err error) {
	ProxyConn, err = ProxyDialer.Dial("tcp", Host)
	if err != nil {
		perr("ERROR Dial %v", err)
		return err
	}

	SshConn, SshNewChannelCh, SshRequestCh, err := ssh.NewClientConn(ProxyConn, Host, SshClientConfig)
	if err != nil {
		perr("ERROR NewClientConn %v", err)
		return err
	}

	// https://pkg.go.dev/golang.org/x/crypto/ssh#NewClient
	SshClient = ssh.NewClient(SshConn, SshNewChannelCh, SshRequestCh)

	// https://pkg.go.dev/golang.org/x/crypto/ssh#Client.NewSession
	var session *ssh.Session

	perr("host[%s] user[%s] [%s]", Host, User, CmdBootTime)
	session, err = SshClient.NewSession()
	if err != nil {
		perr("ERROR CmdBootTime NewSession %v", err)
		return err
	}
	procstatbb, err := session.Output(CmdBootTime)
	if err != nil {
		perr("WARNING CmdBootTime Output %v", err)
	}
	if boottimem := regexp.MustCompile(`(?m)^btime ([0-9]+)$`).FindStringSubmatch(string(procstatbb)); boottimem != nil {
		BootTime, err = strconv.ParseInt(boottimem[1], 10, 64)
		if err != nil {
			perr("WARNING CmdBootTime btime ParseInt %v", err)
		}
	}
	session.Close()

	perr("host[%s] user[%s] [%s]", Host, User, CmdBootId)
	session, err = SshClient.NewSession()
	if err != nil {
		perr("ERROR CmdBootId NewSession %v", err)
		return err
	}
	bootidbb, err := session.Output(CmdBootId)
	if err != nil {
		perr("WARNING CmdBootId Output %v", err)
	}
	BootId = string(bootidbb)
	if len(BootId) > 4 {
		BootId = BootId[:4]
	}
	session.Close()

	perr("host[%s] user[%s] [%s]", Host, User, CmdHostname)
	session, err = SshClient.NewSession()
	if err != nil {
		perr("ERROR CmdHostname NewSession %v", err)
		return err
	}
	hostnamebb, err := session.Output(CmdHostname)
	if err != nil {
		perr("WARNING CmdHostname Output %v", err)
	}
	Hostname = strings.TrimSpace(string(hostnamebb))
	session.Close()

	perr("host[%s] user[%s] [%s]", Host, User, CmdPwd)
	session, err = SshClient.NewSession()
	if err != nil {
		perr("ERROR CmdPwd NewSession %v", err)
		return err
	}
	pwdbb, err := session.Output(CmdPwd)
	if err != nil {
		perr("WARNING CmdPwd Output %v", err)
	}
	Pwd = strings.TrimSpace(string(pwdbb))
	session.Close()

	return nil
}

func GetAllPathCmds() error {
	perr("host[%s] user[%s] [%s]", Host, User, CmdAllPathCmds)
	session, err := SshClient.NewSession()
	if err != nil {
		perr("ERROR CmdAllPathCmds NewSession %v", err)
		return err
	}
	pathcmdsbb, err := session.Output(CmdAllPathCmds)
	if err != nil {
		perr("WARNING CmdAllPathCmds Output %v", err)
		perr(string(pathcmdsbb))
		return err
	}
	AllPathCmds = strings.Split(string(pathcmdsbb), NL)
	perr("DEBUG AllPathCmds <%d>", len(AllPathCmds))
	session.Close()
	return nil
}

func GetAllFiles(fpathdir string) error {
	perr("host[%s] user[%s] [%s]", Host, User, F(CmdAllFiles, fpathdir))
	session, err := SshClient.NewSession()
	if err != nil {
		perr("ERROR CmdAllFiles NewSession %v", err)
		return err
	}
	allfilesbb, err := session.Output(F(CmdAllFiles, fpathdir))
	if err != nil {
		perr("WARNING CmdAllFiles Output %v", err)
		perr(string(allfilesbb))
		return err
	}
	AllFiles = strings.Split(string(allfilesbb), NL)
	perr("DEBUG AllFiles <%d>", len(AllFiles))
	session.Close()
	return nil
}

// https://github.com/golang/go/issues/21478
// https://github.com/golang/go/issues/19338
// https://pkg.go.dev/golang.org/x/crypto/ssh
func keepalive(cl *ssh.Client, conn net.Conn, done <-chan bool) (err error) {
	perr("VERBOSE keepalive start")
	t := time.NewTicker(SshKeepAliveInterval)
	defer t.Stop()
	for {
		/*
			err = conn.SetDeadline(time.Now().Add(2 * SshKeepAliveInterval))
			if err != nil {
				perr("VERBOSE keepalive failed to set deadline %v", err)
				return fmt.Errorf("failed to set deadline %w", err)
			}
		*/
		select {
		case <-t.C:
			_, _, err = cl.SendRequest("keepalive@github.com/shoce/hs", true, nil)
			if err != nil {
				perr("VERBOSE keepalive failed to send request %v", err)
				return fmt.Errorf("keepalive failed to send request %w", err)
			} else {
				perr("VERBOSE keepalive request sent and confirmed")
			}
		case <-done:
			perr("VERBOSE keepalive done")
			return nil
		}
	}
}

func runssh(cmds string, cmd []string, stdin io.Reader) (status string, err error) {
	if SshClient == nil {
		err = connectssh()
		if err != nil {
			return "", err
		}
	}

	session, err := SshClient.NewSession()
	if err != nil {
		perr("ERROR NewSession %v", err)
		perr("reconnecting...")
		err = connectssh()
		if err != nil {
			return "", err
		}
		session, err = SshClient.NewSession()
	}
	if err != nil {
		perr("ERROR NewSession %v", err)
		return "", err
	}

	/*
		for _, s := range []string{"Dir"} {
			if os.Getenv(s) == "" {
				continue
			}
			if err := session.Setenv(s, os.Getenv(s)); err != nil {
				// ( echo ; echo AcceptEnv Dir ) >>/etc/ssh/sshd_config && systemctl reload sshd
				perr("ERROR Session.Setenv [%s] %v", s, err)
				return "", err
			}
		}
	*/

	/*
		if err := session.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"); err != nil {
			// ( echo ; echo AcceptEnv PATH ) >>/etc/ssh/sshd_config && systemctl reload sshd
			perr("ERROR Session.Setenv [PATH] %v", err)
			return "", err
		}
	*/

	if stdin != nil {
		stdinpipe, err := session.StdinPipe()
		if err != nil {
			return "", fmt.Errorf("ERROR session stdin pipe %v", err)
		}

		go func() {
			_, err := io.Copy(stdinpipe, stdin)
			if err != nil {
				perr("ERROR stdin pipe Copy %v", err)
			}

			err = stdinpipe.Close()
			if err != nil {
				perr("ERROR stdin pipe Close %v", err)
			}
		}()
	}

	stdoutpipe, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe for session %v", err)
	}

	stderrpipe, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("stderr pipe for session %v", err)
	}

	copyoutnotify := make(chan error)
	go copynotify(os.Stdout, stdoutpipe, copyoutnotify)
	copyerrnotify := make(chan error)
	go copynotify(os.Stderr, stderrpipe, copyerrnotify)

	err = session.Start(cmds)

	if err != nil {
		perr("ERROR Start %v", err)
		return "", err
	}

	keepalivedonechan := make(chan bool)
	go keepalive(SshClient, ProxyConn, keepalivedonechan)

	InterruptChan = make(chan bool)
	go func() {
		interrupt := <-InterruptChan
		if !interrupt {
			return
		}
		// https://pkg.go.dev/golang.org/x/crypto/ssh
		err := session.Signal(ssh.SIGINT)
		if err != nil && err == io.EOF {
			perr("ERROR session Signal [SIGINT] EOF")
		} else if err != nil {
			perr("ERROR session Signal [SIGINT] %v", err)
		}
	}()

	err = session.Wait()

	keepalivedonechan <- true
	close(keepalivedonechan)
	keepalivedonechan = nil

	close(InterruptChan)
	InterruptChan = nil

	if err != nil {
		switch err.(type) {
		case *ssh.ExitMissingError:
			status = "missing"
		case *ssh.ExitError:
			exiterr := err.(*ssh.ExitError)
			status = F("%d", exiterr.ExitStatus())
			if sig := exiterr.Signal(); sig != "" {
				status += "-" + sig
			}
		default:
			perr("ERROR Wait %v", err)
			return "", err
		}
	}

	err = <-copyoutnotify
	if err != nil {
		perr("(%s) ERROR out copy %v", cmds, err)
	}

	err = <-copyerrnotify
	if err != nil {
		perr("(%s) ERROR err copy %v", cmds, err)
	}

	return status, nil
}

func runlocal(cmds string, cmd []string, stdin io.Reader) (status string, err error) {
	var cmdargs []string
	if len(cmd) > 1 {
		cmdargs = cmd[1:]
	}

	command := exec.Command(cmd[0], cmdargs...)

	var stdinpipe io.WriteCloser
	var stdoutpipe, stderrpipe io.ReadCloser

	if stdin != nil {
		stdinpipe, err = command.StdinPipe()
		if err != nil {
			return "", fmt.Errorf("stdin pipe for command: %v", err)
		}

		go func() {
			_, err := io.Copy(stdinpipe, stdin)
			if err != nil {
				perr("ERROR stdin pipe Copy %v", err)
			}

			err = stdinpipe.Close()
			if err != nil {
				perr("ERROR stdin pipe Close %v", err)
			}
		}()
	}

	stdoutpipe, err = command.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe for command: %v", err)
	}
	copyoutnotify := make(chan error)
	go copynotify(os.Stdout, stdoutpipe, copyoutnotify)

	stderrpipe, err = command.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("stderr pipe for command: %v", err)
	}
	copyerrnotify := make(chan error)
	go copynotify(os.Stderr, stderrpipe, copyerrnotify)

	perr("%s: ", cmds)

	err = command.Start()
	if err != nil {
		return "", fmt.Errorf("Start command: %v", err)
	}

	err = command.Wait()

	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			exiterr := err.(*exec.ExitError)
			status = F("%d", exiterr.ExitCode())
		default:
			return "", fmt.Errorf("Wait: %v", err)
		}
	}

	return status, nil
}

func run(cmds string, cmd []string, stdin io.Reader) (status string, err error) {
	if cmds == "" && len(cmd) == 0 {
		return "", errors.New("empty cmd")
	}

	perr("host[%s] user[%s] [%s]", Host, User, cmds)

	if Host == "" {
		return runlocal(cmds, cmd, stdin)
	} else {
		return runssh(cmds, cmd, stdin)
	}
}

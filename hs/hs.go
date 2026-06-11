/*
Hs

HISTORY
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
026/04	cmd<TAB> and cmd arg<TAB> "autocompletion"
026/0601	clip[]
*/

/*
GoGet 
GoFmt 
GoBuildNull 
GoBuild
GoRun -- put a '<' <readme.text #ae:>>
Kill GoRun
*/

/*

VARIABLES
`host` variable is checked before every command execution: if it is empty then run locally; otherwise run via ssh.
`user` variable stores user name if run via ssh.
'userpass' 
'userkeyfile' 
'userkey' 
`status` variable tells exit status of the last command executed.

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
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	ssh "golang.org/x/crypto/ssh"
	proxy "golang.org/x/net/proxy"
)

const (
	N = ""
	SP  = " "
	TAB = "\t"
	NL  = "\n"
	CR = "\r"
	SEP = ","

	PortDefault = "22"

	InReaderBufferSize = 100 * 1000

	CmdHostInfo = `cat /proc/sys/kernel/hostname /proc/sys/kernel/osrelease /proc/sys/kernel/arch /proc/sys/kernel/random/boot_id /proc/stat`
	CmdPwd      = `pwd`

	CmdAllPathCmds = `dd="" ; for d in ${PATH//:/ } ; do test -L "$d" && d=$(readlink -f "$d") ; test -d "$d" && dd="$dd $d" ; done ; ddd="" ; for d in $dd ; do for di in $ddd ; do test "$di" = "$d" && continue 2 ; done ; ddd="$ddd $d" ; done ; for d in $ddd ; do find "$d/" -maxdepth 1 -type f -executable -print -o -type l -exec sh -c 'test -x "{}"' \; -print | LC_ALL=C sort ; done ;`
	CmdAllFiles    = `find "%s" -maxdepth 1 -print | LC_ALL=C sort`

	CmdHistoryMax = 1111
)

type CmdHistoryRecord struct {
	Timestamp time.Time
	Cmds      string
}

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

	SshGatherHostInfo bool

	Hostname string
	BootTime int64
	BootId   string
	Pwd      string
	Kernel   string
	Arch     string

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

	CmdHistory []CmdHistoryRecord

	InterruptChan chan bool
	
	ClipEnabled bool

	TzBiel *time.Location = time.FixedZone("Biel", 60*60)

	F = fmt.Sprintf
	pout = fmt.Print
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

	Proxy = os.Getenv("proxy")
	perr(F("VERBOSE Proxy [%s]", Proxy))
	ProxyChain = strings.FieldsFunc(Proxy, func(c rune) bool { return c == ';' })
	perr(F("VERBOSE ProxyChain <%d> %v", len(ProxyChain), ProxyChain))
	ProxyDialer = proxy.Direct

	Host = os.Getenv("host")
	perr(F("VERBOSE Host [%s]", Host))
	if Host == "" {
		perr(F("ERROR host env var empty"))
		os.Exit(1)
	}

	User = os.Getenv("user")
	perr(F("VERBOSE User [%s]", User))
	if User == "" {
		perr(F("ERROR user env var empty"))
		os.Exit(1)
	}

	UserPassword = os.Getenv("userpass")
	if UserPassword != "" {
		UserAuthMethod = ssh.Password(UserPassword)
	}
	perr(F("VERBOSE UserPassword [%s]", UserPassword))

	if os.Getenv("home") == "" {
		err = os.Setenv("home", os.Getenv("HOME"))
		if err != nil {
			perr(F("WARNING Setenv home %v", err))
		}
	}

	UserKeyFile = os.ExpandEnv(os.Getenv("userkeyfile"))
	if UserKeyFile != "" {
		userkeybb, err := ioutil.ReadFile(UserKeyFile)
		if err != nil {
			perr(F("ERROR Read UserKeyFile %v", err))
			os.Exit(1)
		}
		UserKey = string(userkeybb)
	}

	if os.Getenv("userkey") != "" {
		UserKey = os.Getenv("UserKey")
	}

	if UserKey != "" {
		UserSigner, err = ssh.ParsePrivateKey([]byte(UserKey))
		if err != nil {
			perr(F("ERROR ParsePrivateKey %v", err))
			os.Exit(1)
		}
		UserAuthMethod = ssh.PublicKeys(UserSigner)
	}
	perr(F("VERBOSE UserKey [%s]", UserKey))

	if UserAuthMethod == nil {
		perr("ERROR no user auth method provided; no password and no user key")
		os.Exit(1)
	}

	SshClientConfig = &ssh.ClientConfig{
		User:    User,
		Auth:    []ssh.AuthMethod{UserAuthMethod},
		Timeout: 10 * time.Second,
		//ClientVersion: "hs", // NewClientConn: ssh: handshake failed: ssh: invalid packet length, packet too large
		//HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			perr(F("VERBOSE ssh server public key type[%s] hex[%s]", key.Type(), hex.EncodeToString(key.Marshal())))
			return nil
		},
		BannerCallback: func(msg string) error {
			msg = strings.TrimSpace(msg)
			sep := N
			if strings.Contains(msg, NL) {
				sep = NL
			}
			perr(F("ssh server banner [%s%s]", sep, msg))
			return nil
		},
	}

}

func main() {
	var err error

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		for {
			sig := <-sigchan
			switch sig {
			case syscall.SIGINT:
				fmt.Fprint(os.Stderr, NL)
				perr("interrupt signal")
				if InterruptChan != nil {
					InterruptChan <- true
				}
			case syscall.SIGHUP:
				fmt.Fprint(os.Stderr, NL)
				perr("hangup signal")
				os.Exit(2)
			}
		}
	}()

	args := os.Args[1:]

	if len(args) > 0 && args[0] != "--" {
		perr("ERROR the first argument should be `--` e.g. `hs -- id`")
		os.Exit(1)
	}

	if Host != "" && len(args) == 1 {
		SshGatherHostInfo = true
	}

	if Host == "" {

		Hostname, err = os.Hostname()
		if err != nil {
			perr(F("ERROR Hostname %v", err))
			os.Exit(1)
		}
		Hostname = strings.TrimSuffix(Hostname, ".local")
		//perr(F("Hostname [%s]", Hostname))

		u, err := user.Current()
		if err != nil {
			perr(F("WARNING user.Current %v", err))
		}
		User = u.Username

	} else {

		if len(ProxyChain) > 0 {
			for _, p := range ProxyChain {
				proxyurl, err := url.Parse(p)
				if err != nil {
					perr(F("ERROR proxy url [%s]` %v", p, err))
					os.Exit(1)
				}
				pd, err := proxy.FromURL(proxyurl, ProxyDialer)
				if err != nil {
					perr(F("ERROR proxy from url [%s] %v", p, err))
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
					perr(F("DEBUG ip [%s]", ip.String()))
					Host = net.JoinHostPort(ip.String(), PortDefault)
				} else {
					Host = net.JoinHostPort(Host, PortDefault)
				}
			} else {
				perr(F("ERROR SplitHostPort %v", err))
				os.Exit(1)
			}
		}

		perr(F("DEBUG Host [%s]", Host))
		err = connectssh()
		if err != nil {
			//perr(F("ERROR connect ssh %v", err))
		}
		if SshClient != nil {
			defer SshClient.Close()
		}
	}

	inreader := bufio.NewReaderSize(os.Stdin, InReaderBufferSize)

	if len(args) > 1 {
		cmd := args[1:]
		cmds := strings.Join(cmd, " ")

		if cmd[len(cmd)-1] == "<" {
			cmd = cmd[:len(cmd)-1]
			cmds = strings.Join(cmd, " ")
			perr(F("host=%s user=%s hs -- %s %s", Host, User, cmds, TermUnderline("stdin:")))
		}

		Status, err = run(cmds, cmd, inreader)
		if err != nil {
			perr(F("host=%s user=%s hs -- %s %s %v", Host, User, cmds, TermInverse("ERROR"), err))
			os.Exit(1)
		}
		if Status != "" {
			perr(F("host=%s user=%s hs -- %s %s", Host, User, cmds, TermInverse(F("status[%s]", Status))))
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
			perr(F("WARNING ReadString %v", err))
			continue
		}

		cmds = strings.TrimSuffix(cmds, NL)
		perr(F("DEBUG cmds [%s]", cmds))
		
		if strings.HasSuffix(cmds, TAB) {
			perr(strings.ReplaceAll(cmds, TAB, `\t`))
			// https://pkg.go.dev/strings#TrimPrefix
			cmds = strings.TrimSuffix(cmds, TAB)
			cmdsff := strings.Split(cmds, SP)
			if len(cmdsff) == 1 {
				cmd := cmdsff[0]
				if len(AllPathCmds) == 0 {
					err := GetAllPathCmds()
					if err != nil {
						//perr(F("ERROR AllPathCmdsGet %v", err))
					}
				}
				for _, c := range AllPathCmds {
					if strings.HasPrefix(c, cmd) || strings.HasPrefix(path.Base(c), cmd) {
						pout(c+NL)
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
					//perr(F("ERROR AllFilesGet %v", err))
				}
				for _, f := range AllFiles {
					// path.Join is intentional because filepath.Join would produce wrong result on Windows connecting to Unix machine
					if strings.HasPrefix(f, fpath) || strings.HasPrefix(f, path.Join(Pwd, fpath)) {
						pout(f+NL)
					}
				}
			}
			continue
		}
		
		switch cmds {
		
		case ":q":
			os.Exit(0)

		case ":p":
			tnow := time.Now()
			for _, ic := range CmdHistory[max(len(CmdHistory)-11, 0):] {
				pout("%s"+NL+TAB+"<%s> ago", ic.Cmds, fmtdur(tnow.Sub(ic.Timestamp)))
			}
			continue

		case ":pp":
			tnow := time.Now()
			for _, ic := range CmdHistory {
				pout("%s"+NL+TAB+"<%s> ago", ic.Cmds, fmtdur(tnow.Sub(ic.Timestamp)))
			}
			continue

		case ":pwd":
			pout(Pwd)
			continue
			
		case ":pb":
			if clip, err := ClipGet(); err != nil {
				perr(F("%v", err))
			} else {
				pout("["+clip+"]"+NL)
			}
			continue
			
		case ":clip":
			ClipEnabled = !ClipEnabled
			perr(F("ClibEnabled <%t>", ClipEnabled))
			continue

		//cmds = strings.TrimSpace(cmds)
		case "":
			continue
			
		}

		cmd := strings.Split(cmds, " ")

		stdinbb = nil
		if len(cmd) > 0 && cmd[len(cmd)-1] == "<" {
			cmd = cmd[:len(cmd)-1]
			cmds = cmds[:len(cmds)-1]
			perr(F("host=%s user=%s hs -- %s %s", Host, User, cmds, TermUnderline("stdin:")))
			stdinbb, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				perr(F("ERROR stdin read %v", err))
				continue
			}
		}

		Status, err = run(cmds, cmd, bytes.NewBuffer(stdinbb))
		if err != nil {
			perr(F("host=%s user=%s hs -- %s %s %v", Host, User, cmds, TermInverse("ERROR"), err))
			continue
		}
	}
}

func ClipGet() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		if pbb, err := exec.Command("pbpaste").Output(); err != nil {
			return "", fmt.Errorf("ERROR exec pbpaste %v", err)
		} else {
			return string(pbb), nil
		}
	default:
		return "", fmt.Errorf("ERROR clipboard not supported")
	}
}

func TermItalic(s string) string {
	if TERM != "" {
		return "\033[3m" + s + "\033[23m" //ae:]]
	}
	return s
}

func TermUnderline(s string) string {
	if TERM != "" {
		return "\033[4m" + s + "\033[24m" //ae:]]
	}
	return s
}

func TermInverse(s string) string {
	if TERM != "" {
		return "\033[7m" + s + "\033[27m" //ae:]]
	}
	return s
}

func TermUnderlineInverse(s string) string {
	if TERM != "" {
		return "\033[4;7m" + s + "\033[24;27m" //ae:]]
	}
	return s
}

func perr(msgtext string) {
	if strings.HasPrefix(msgtext, "DEBUG ") && !DEBUG {
		return
	}
	if strings.HasPrefix(msgtext, "VERBOSE ") && !VERBOSE {
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
	fmt.Fprint(os.Stderr, ts+" "+msgtext+NL)

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

func fmtdur(td time.Duration) string {
	tdays, tsecs := uint64(td.Seconds())/(24*3600), uint64(td.Seconds())%(24*3600)
	ts := seps(tsecs, 2) + "s"
	if tdays > 0 {
		ts = seps(tdays, 2) + "d" + SEP + ts
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
	fmt.Fprint(os.Stderr, NL)
	if ClipEnabled {
	if clip, err := ClipGet(); err == nil {
		if len(clip) > 123 {
			clip = clip[:123]+"…"
		}
		clip = strings.ReplaceAll(clip, NL, `\n`)
		clip = strings.ReplaceAll(clip, CR, `\r`)
		clip = strings.ReplaceAll(clip, TAB, `\t`)
		perr(TermUnderline("clip["+clip+"]"))
	}
	}
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
		" hostname[%s] uptime<%s> bootid[%s] kernel[%s] arch[%s] host=%s user=%s hs -- ",
		Hostname, uptime, BootId, Kernel, Arch, Host, User,
	)
	s = TermUnderline(s)
	perr(s)
}

func copynotify(dst io.Writer, src io.Reader, notify chan error) {
	_, err := io.Copy(dst, src)
	if notify != nil {
		notify <- err //ae:>
	}
}

func connectssh() (err error) {
	ProxyConn, err = ProxyDialer.Dial("tcp", Host)
	if err != nil {
		perr(F("ERROR Dial %v", err))
		return err
	}

	SshConn, SshNewChannelCh, SshRequestCh, err := ssh.NewClientConn(ProxyConn, Host, SshClientConfig)
	if err != nil {
		perr(F("ERROR NewClientConn %v", err))
		return err
	}

	// https://pkg.go.dev/golang.org/x/crypto/ssh#NewClient
	SshClient = ssh.NewClient(SshConn, SshNewChannelCh, SshRequestCh)

	if !SshGatherHostInfo {
		return nil
	}

	// https://pkg.go.dev/golang.org/x/crypto/ssh#Client.NewSession
	var session *ssh.Session

	perr(F("host=%s user=%s hs -- %s", Host, User, CmdHostInfo))
	session, err = SshClient.NewSession()
	if err != nil {
		perr(F("ERROR CmdHostInfo NewSession %v", err))
		return err
	}
	hostinfobb, err := session.Output(CmdHostInfo)
	if err != nil {
		perr(F("WARNING CmdHostInfo Output %v", err))
	}
	session.Close()
	hostinfo := strings.Split(strings.TrimSpace(string(hostinfobb)), NL)
	if len(hostinfo) > 0 {
		Hostname = hostinfo[0]
	}
	if len(hostinfo) > 1 {
		Kernel = hostinfo[1]
	}
	if len(hostinfo) > 2 {
		Arch = hostinfo[2]
	}
	if len(hostinfo) > 3 {
		BootId = hostinfo[3]
		if len(BootId) > 4 {
			BootId = BootId[:4]
		}
	}
	if len(hostinfo) > 5 {
		for _, l := range hostinfo[5:] {
			if ff := strings.Fields(l); len(ff) == 2 && ff[0] == "btime" {
				BootTime, err = strconv.ParseInt(ff[1], 10, 64)
				if err != nil {
					perr(F("WARNING CmdHostInfo btime ParseInt %v", err))
				}
			}
		}
	}

	perr(F("host=%s user=%s hs -- %s", Host, User, CmdPwd))
	session, err = SshClient.NewSession()
	if err != nil {
		perr(F("ERROR CmdPwd NewSession %v", err))
		return err
	}
	pwdbb, err := session.Output(CmdPwd)
	if err != nil {
		perr(F("WARNING CmdPwd Output %v", err))
	}
	Pwd = strings.TrimSpace(string(pwdbb))
	session.Close()

	return nil
}

func GetAllPathCmds() error {
	perr(F("host=%s user=%s hs -- %s", Host, User, CmdAllPathCmds))
	session, err := SshClient.NewSession()
	if err != nil {
		perr(F("ERROR CmdAllPathCmds NewSession %v", err))
		return err
	}
	pathcmdsbb, err := session.Output(CmdAllPathCmds)
	if err != nil {
		perr(F("WARNING CmdAllPathCmds Output %v", err))
		perr(string(pathcmdsbb))
		return err
	}
	AllPathCmds = strings.Split(string(pathcmdsbb), NL)
	perr(F("DEBUG AllPathCmds <%d>", len(AllPathCmds)))
	session.Close()
	return nil
}

func GetAllFiles(fpathdir string) error {
	perr(F("host=%s user=%s hs -- %s", Host, User, F(CmdAllFiles, fpathdir)))
	session, err := SshClient.NewSession()
	if err != nil {
		perr(F("ERROR CmdAllFiles NewSession %v", err))
		return err
	}
	allfilesbb, err := session.Output(F(CmdAllFiles, fpathdir))
	if err != nil {
		perr(F("WARNING CmdAllFiles Output %v", err))
		perr(string(allfilesbb))
		return err
	}
	AllFiles = strings.Split(string(allfilesbb), NL)
	perr(F("DEBUG AllFiles <%d>", len(AllFiles)))
	session.Close()
	return nil
}

// https://github.com/golang/go/issues/21478
// https://github.com/golang/go/issues/19338
// https://pkg.go.dev/golang.org/x/crypto/ssh
func keepalive(cl *ssh.Client, conn net.Conn, done <-chan bool) (err error) { //ae:>
	perr("VERBOSE keepalive start")
	t := time.NewTicker(SshKeepAliveInterval)
	defer t.Stop()
	for {
		/*
			err = conn.SetDeadline(time.Now().Add(2 * SshKeepAliveInterval))
			if err != nil {
				perr(F("VERBOSE keepalive failed to set deadline %v", err))
				return fmt.Errorf("failed to set deadline %w", err)
			}
		*/
		select {
		case <-t.C: //ae:>
			_, _, err = cl.SendRequest("keepalive@github.com/shoce/hs", true, nil)
			if err != nil {
				perr(F("VERBOSE keepalive failed to send request %v", err))
				return fmt.Errorf("keepalive failed to send request %w", err)
			} else {
				perr("VERBOSE keepalive request sent and confirmed")
			}
		case <-done: //ae:>
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
		perr(F("ERROR NewSession %v", err))
		perr("reconnecting...")
		err = connectssh()
		if err != nil {
			return "", err
		}
		session, err = SshClient.NewSession()
	}
	if err != nil {
		perr(F("ERROR NewSession %v", err))
		return "", err
	}

	/*
		for _, s := range []string{"Dir"} {
			if os.Getenv(s) == "" {
				continue
			}
			if err := session.Setenv(s, os.Getenv(s)); err != nil {
				//ae:<<
				// ( echo ; echo AcceptEnv Dir ) >>/etc/ssh/sshd_config && systemctl reload sshd
				perr(F("ERROR Session.Setenv [%s] %v", s, err))
				return "", err
			}
		}
	*/

	/*
		if err := session.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"); err != nil {
			//ae:<<
			// ( echo ; echo AcceptEnv PATH ) >>/etc/ssh/sshd_config && systemctl reload sshd
			perr(F("ERROR Session.Setenv [PATH] %v", err))
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
				perr(F("ERROR stdin pipe Copy %v", err))
			}

			err = stdinpipe.Close()
			if err != nil {
				perr(F("ERROR stdin pipe Close %v", err))
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
		perr(F("ERROR Start %v", err))
		return "", err
	}

	keepalivedonechan := make(chan bool)
	go keepalive(SshClient, ProxyConn, keepalivedonechan)

	InterruptChan = make(chan bool)
	go func() {
		interrupt := <-InterruptChan //ae:>
		if !interrupt {
			return
		}
		// https://pkg.go.dev/golang.org/x/crypto/ssh
		err := session.Signal(ssh.SIGINT)
		if err != nil && err == io.EOF {
			perr(F("ERROR session Signal [SIGINT] EOF"))
		} else if err != nil {
			perr(F("ERROR session Signal [SIGINT] %v", err))
		}
	}()

	err = session.Wait()

	keepalivedonechan <- true //ae:>
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
			perr(F("ERROR Wait %v", err))
			return "", err
		}
	}

	err = <-copyoutnotify
	if err != nil {
		perr(F("(%s) ERROR out copy %v", cmds, err))
	}

	err = <-copyerrnotify
	if err != nil {
		perr(F("(%s) ERROR err copy %v", cmds, err))
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
				perr(F("ERROR stdin pipe Copy %v", err))
			}

			err = stdinpipe.Close()
			if err != nil {
				perr(F("ERROR stdin pipe Close %v", err))
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

	perr(F("%s: ", cmds))

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

	CmdHistory = append(CmdHistory, CmdHistoryRecord{time.Now(), cmds})
	if len(CmdHistory) > CmdHistoryMax {
		CmdHistory = CmdHistory[len(CmdHistory)-CmdHistoryMax:]
	}
	//fmt.Fprint(os.Stderr, NL)
	perr(TermUnderline(F("host=%s user=%s hs -- %s", Host, User, cmds)))

	if Host == "" {
		return runlocal(cmds, cmd, stdin)
	} else {
		return runssh(cmds, cmd, stdin)
	}
}


/*
HISTORY
026/0726 v1
*/
/* 
GoGet
GoBuildNull
GoBuild
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	NL = "\n"
)

var (
	VERSION string
	VERBOSE bool
	DEBUG bool
	
	F = fmt.Sprintf
	pout = fmt.Print
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		pout(VERSION+NL)
		os.Exit(0)
	}
	if os.Getenv("DEBUG") != "" {
		VERBOSE = true
		DEBUG = true
		perr("DEBUG <true>")
	}
	if os.Getenv("VERBOSE") != "" {
		VERBOSE = true
		perr("VERBOSE <true>")
	}
	
	var err error
	cmdbegin := 0
	for i := range os.Args {
		if os.Args[i] == "--" {
			cmdbegin = i+1
			break
		}
	}
	if len(os.Args) < 2 || cmdbegin == 0 || cmdbegin > len(os.Args)-1 {
		perr("USAGE envset name1=value1 name2=value2 ... -- command [args]")
		os.Exit(1)
	}
	
	Vars := os.Args[1:cmdbegin-1]
	cmd := os.Args[cmdbegin]
	args := os.Args[cmdbegin+1:]
	perr(F("DEBUG Vars (%s)", AtonListString(Vars)))
	perr(F("DEBUG cmd [%s]", cmd))
	perr(F("DEBUG args (%s)", AtonListString(args)))
	
	Command := exec.Command(cmd, args...)
	Command.Stdin, Command.Stdout, Command.Stderr = os.Stdin, os.Stdout, os.Stderr
	for _, e := range os.Environ() {
		Command.Env = append(Command.Env, e)
	}
	for _, v := range Vars {
		vs := os.ExpandEnv(v)
		perr(F("VERBOSE [%s]", vs))
		Command.Env = append(Command.Env, vs)
	}
	err = Command.Run()
	os.Stderr.Sync()
	os.Stdout.Sync()
	if err != nil {
		perr(F("ERROR [%s] %v", Command, err)) 
	}
	
}

func perr(msg string) {
	if strings.HasPrefix(msg, "DEBUG ") && !DEBUG { return }
	if strings.HasPrefix(msg, "VERBOSE ") && !VERBOSE { return }
	fmt.Fprint(os.Stderr, msg+NL)
}

func AtonListString(ss []string) string {
        var aa []string
        for _, s := range ss {
                aa = append(aa, "["+s+"]")
        }
        return strings.Join(aa, " ")
}


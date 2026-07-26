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
	if len(os.Args) < 4 || cmdbegin == 0 || cmdbegin > len(os.Args)-1 {
		perr("USAGE envrange varname value1 value2 value3... -- command [args]")
		os.Exit(1)
	}
	
	VarName := os.Args[1]
	VarValues := os.Args[2:cmdbegin-1]
	cmd := os.Args[cmdbegin]
	args := os.Args[cmdbegin+1:]
	perr(F("DEBUG VarName [%s]", VarName))
	perr(F("DEBUG VarValues (%v)", VarValues))
	perr(F("DEBUG cmd [%s]", cmd))
	perr(F("DEBUG args (%v)", args))
	
	for _, v := range VarValues {
		Command := exec.Command(cmd, args...)
		Command.Stdin, Command.Stdout, Command.Stderr = os.Stdin, os.Stdout, os.Stderr
		for _, e := range os.Environ() {
			Command.Env = append(Command.Env, e)
		}
		Command.Env = append(Command.Env, VarName+"="+v)
		perr(F("VERBOSE [%s][%s] [%s]", VarName, v, Command))
		err = Command.Run()
		os.Stderr.Sync()
		os.Stdout.Sync()
		if err != nil {
			perr(F("ERROR [%s][%s] [%s] %v", VarName, v, Command, err)) 
		}
	}

}

func perr(msg string) {
	if strings.HasPrefix(msg, "DEBUG ") && !DEBUG { return }
	if strings.HasPrefix(msg, "VERBOSE ") && !VERBOSE { return }
	fmt.Fprint(os.Stderr, msg+NL)
}


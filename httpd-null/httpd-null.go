/*
GoGet 
GoFmt 
GoBuildNull 
GoBuild
GoRun
ListenAddr=:8080 GoRun
IfFileExists=$home/test ListenAddr=:8080 GoRun
*/

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	NL = "\n"
)

var (
	DEBUG bool
	
	F = fmt.Sprintf
	EF = fmt.Errorf
	pout = fmt.Print
)

func main() {
	var err error
	DEBUG = os.Getenv("DEBUG") != ""
	ListenAddr := os.Getenv("ListenAddr")
	if ListenAddr == "" {
		ListenAddr = ":80"
	}
	perr(F("listening on `%s`", ListenAddr))
	IfFileExists := os.Getenv("IfFileExists")
	if IfFileExists != "" {
		perr(F("depending on file `%s` exists", IfFileExists))
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var rbody []byte
		rbody, err = io.ReadAll(r.Body)
		if err != nil {
			perr(F("ERROR request body ReadAll %v", err))
		}
		defer r.Body.Close()

		perr(F("DEBUG proto[%s] method[%s] path[%s] body[%s]", r.Proto, r.Method, r.URL.Path, string(rbody)))
		if IfFileExists != "" {
			if _, err := os.Stat(IfFileExists); errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(ListenAddr, nil); err != nil {
		perr(F("ERROR ListenAndServe [%s] %v", ListenAddr, err))
		os.Exit(1)
	}
}

func perr(msg string) (int, error) {
	if strings.HasPrefix(msg, "DEBUG ") && !DEBUG {
		return 0, nil
	}
	return fmt.Fprint(os.Stderr, msg+NL)
}

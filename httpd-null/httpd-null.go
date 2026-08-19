/*
GoGet
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
)
const (
	ListenAddrDef = ":80"
	NL = "\n"
)
var (
	ListenAddr string
	F = fmt.Sprintf
	EF = fmt.Errorf
	pout = fmt.Print
)

func main() {
	var err error
	ListenAddr = ListenAddrDef
	if la:=os.Getenv("ListenAddr"); la!="" { ListenAddr = la }
	perr(F("ListenAddr [%s]", ListenAddr))
	IfFileExists := os.Getenv("IfFileExists")
	perr(F("IfFileExists [%s]", IfFileExists))
	if IfFileExists != "" {
		perr(F("depending on file [%s] exists", IfFileExists))
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var rbody []byte
		rbody, err = io.ReadAll(r.Body)
		if err != nil { perr(F("ERROR request body ReadAll %v", err)) }
		defer r.Body.Close()
		perr(F("proto[%s] method[%s] path[%s] body[%s]", r.Proto, r.Method, r.URL.Path, string(rbody)))
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
func perr(msg string) (int, error) { return fmt.Fprint(os.Stderr, msg+NL) }


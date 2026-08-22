package main
import (
	"bytes"
	"fmt"
	"io"
	"os"
)
const (
	nl='\n'
	NL="\n"
)
var (
	F=fmt.Sprintf
	pout=fmt.Print
)
func main() {
	var err error
	bb,err:=io.ReadAll(os.Stdin)
	if err!=nil {
		perr(F("ReadAll %v", err))
		os.Exit(1)
	}
	nlnext:=len(bb)
	for nlnext>0 {
		nlprev:=bytes.LastIndexByte(bb[:nlnext], nl)
		if nlprev<0 { nlprev=-1 }
		_,err=os.Stdout.Write(bb[nlprev+1:nlnext])
		if err!=nil {
			perr(F("Stdout Write %v", err))
			os.Exit(1)
		}
		_,err=os.Stdout.Write([]byte{nl})
		if err!=nil {
			perr(F("Stdout Write %v", err))
			os.Exit(1)
		}
		nlnext=nlprev
	}
	
}

func perr(msg string) { fmt.Fprint(os.Stderr, msg+NL) }



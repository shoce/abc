/*

GoGet GoFmt GoBuildNull GoBuild
GoRun
ListenAddr=:8080 GoRun
IfFileExists=$home/test ListenAddr=:8080 GoRun

*/

package main

import (
	"errors"
	"log"
	"net/http"
	"os"
)

func main() {
	LogRequests := os.Getenv("LogRequests")
	ListenAddr := os.Getenv("ListenAddr")
	if ListenAddr == "" {
		ListenAddr = ":80"
	}
	log.Printf("Listening on `%s`.", ListenAddr)
	IfFileExists := os.Getenv("IfFileExists")
	if IfFileExists != "" {
		log.Printf("Depending on file `%s` exists.", IfFileExists)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if LogRequests != "" {
			log.Printf("%#v\n", r)
		}
		if IfFileExists != "" {
			if _, err := os.Stat(IfFileExists); errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	log.Fatal(http.ListenAndServe(ListenAddr, nil))
}

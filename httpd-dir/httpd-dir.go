/*
GoGet GoFmt GoBuildNull
GoBuild GoRelease GoRun
*/

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func http_log(h1 http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf(
			"%s %s `%s` User-Agent=`%s`",
			req.RemoteAddr, req.Method, req.URL, req.Header.Get("User-Agent"),
		)
		h1.ServeHTTP(w, req)
	})
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Printf("Usage: %s [dir] [[addr]:port]\n", os.Args[0])
		return
	}

	var dir string
	var addrport string

	dir = "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	addrport = "0.0.0.0:80"
	if len(os.Args) > 2 {
		addrport = os.Args[2]
	}

	log.Printf("Serving dir `%s` on http://%s/\n", dir, addrport)

	http.Handle("/", http_log(http.FileServer(http.Dir(dir))))

	log.Fatal(http.ListenAndServe(addrport, nil))
}

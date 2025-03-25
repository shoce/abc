// history:
// 2015-1117 v1
// 2015-1120 http+html version

// go run timezones.go
// go fmt timezones.go
// go build -o timezones timezones.go
// ./timezones

package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

const listenAddr = "0:5873"
const timeFmt = "Mon 01/02 15:04"

var locationsNames = []string{
	"America/Los_Angeles",
	"America/Denver",
	"America/New_York",
	"UTC",
	"Europe/London",
	"Europe/Moscow",
	"Australia/Sydney",
}
var locations []*time.Location
var locationsOffsets []string

func httpTimezones(w http.ResponseWriter, req *http.Request) {
	var timezones [][]string
	timezones = append(timezones, []string{"Timezone", "Offset", "Now"})
	t1, _ := time.Parse("15:04", "9:00")
	t2, _ := time.Parse("15:04", "18:00")
	now := time.Now()
	for _, l := range locations {
		timezones = append(timezones, []string{
			l.String(),
			now.In(l).Format("-0700"),
			t1.In(l).Format(timeFmt),
			t2.In(l).Format(timeFmt),
			now.In(l).Format(timeFmt),
			})
	}


	timezonesTmpl, err := template.ParseFiles("timezones.html")
	if err != nil {
		log.Print("template.html: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	timezonesTmpl.Execute(w, timezones)
}

func main() {
	var err error

	for _, locName := range locationsNames {
		l, err := time.LoadLocation(locName)
		if err != nil {
			log.Fatal(locName, err)
		}
		locations = append(locations, l)
	}

	http.HandleFunc("/", httpTimezones)
	log.Print("Starting to listen http://", listenAddr)
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

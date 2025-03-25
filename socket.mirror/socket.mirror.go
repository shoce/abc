package main

import (
	"io"
	"log"
	"net"
)

func listen(l net.Listener) {
	var err error
	for {
		var c net.Conn
		c, err = l.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		log.Print("accepted")
		go func(c net.Conn) {
			var err error
			var n int64
			n, err = io.Copy(c, c)
			if err != nil {
				log.Print(err)
			}
			log.Print("copied ", n)
			c.Close()
		}(c)
	}
}

func main() {
	var err error

	var l1 net.Listener
	l1, err = net.Listen("unix", "/re")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("started listening on `/re`")
	defer l1.Close()
	go listen(l1)

	var l2 net.Listener
	l2, err = net.Listen("tcp4", "0.0.0.0:2")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("started listening on `0.0.0.0:2`")
	defer l2.Close()
	go listen(l2)

	select {}
}

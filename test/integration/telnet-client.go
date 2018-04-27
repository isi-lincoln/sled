package main

import (
	"github.com/ziutek/telnet"
	"log"
)

func main() {
	t, err := telnet.Dial("tcp", "localhost:4007")
	if err != nil {
		log.Fatalf("%v", err)
	}

	buf := make([]byte, 512)

	// set the mac address for the interface
	buf = []byte("ip link set eth1 address 00:00:00:00:00:01\r\n")
	n, err := t.Write(buf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("%d: %s", n, buf[:n])

	// show the link information
	buf = []byte("ip link show\r\n")
	n, err = t.Write(buf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("%d: %s", n, buf[:n])
	buf = []byte("")

	// check the output of ip link show
	buf, err = t.ReadUntil("%")
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("%s", buf)
}

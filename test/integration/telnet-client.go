package main

import (
	"fmt"
	"github.com/ziutek/telnet"
	"log"
	"regexp"
	"strings"
)

func main() {
	// TODO: programatically find the port
	t, err := telnet.Dial("tcp", "localhost:4007")
	if err != nil {
		log.Fatalf("%v", err)
	}

	iface := "eth1"
	macAddr := "00:00:00:00:00:01"

	buf := make([]byte, 512)
	lastBuf := make([]byte, 512)

	// set the mac address for the interface for bolt-db
	buf = []byte(fmt.Sprintf("ip link set %s address %s\r\n", iface, macAddr))
	_, err = t.Write(buf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// read out bytes until '%' to clear output for this command
	_, err = t.ReadBytes('%')
	if err != nil {
		log.Fatalf("%v", err)
	}

	// show the link information
	buf = []byte("ip link show\r\n")
	_, err = t.Write(buf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// read out bytes until '%' to clear output for this command
	_, err = t.ReadBytes('%')
	if err != nil {
		log.Fatalf("%v", err)
	}

	// clear buffer before we start writing to it with info
	buf = []byte("")

	// check the output of ip link show
	buf, err = t.ReadBytes('%')
	// while our slice is not empty
	for !emptySlice(buf) {
		if err != nil {
			log.Fatalf("%v", err)
		}
		//log.Printf("%s", buf)
		lastBuf = buf
		buf, err = t.ReadBytes('%')
	}

	mac := getLinkMAC(iface, lastBuf)
	log.Printf("%s", mac)

}

// return true is slice is 'empty' == ' %' for sled
// necessary to not try and poll reading
func emptySlice(n []byte) bool {
	if n[0] == ' ' && n[1] == '%' {
		return true
	}
	for i := 2; i < len(n); i++ {
		if string(n[i]) != "" {
			return false
		}
	}
	return true
}

// given ip link output, get the mac address for an interface
func getLinkMAC(iface string, buf []byte) string {
	// find the index that has the interface
	indexBegin := strings.Index(string(buf), iface)
	// find the first new line, start here
	indexFirst := strings.Index(string(buf[indexBegin:]), "\n")
	// find the next new line, this is the new region we care about
	indexLast := strings.Index(string(buf[indexBegin+indexFirst+1:]), "\n")
	macLine := string(buf[indexBegin+indexFirst+1 : indexBegin+indexFirst+indexLast])
	// get the mac address
	Re := regexp.MustCompile("([0-9a-f][0-9a-f]:){5}[0-9a-f][0-9a-f]")
	return Re.FindString(macLine)
}

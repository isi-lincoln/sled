package main

import (
	"fmt"
	"github.com/ziutek/telnet"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

func main() {
	iface := "eth1"
	macAddr := "00:00:00:00:00:01"
	// set the mac address
	SetClientMAC(iface, macAddr)
	// test that the mac address was set correctly
	success, mac := CheckClientMAC(iface, macAddr)
	log.Printf("%v %v", success, mac)
}

func SetClientMAC(iface, macAddr string) {
	host, port := findClientPort()
	t, err := telnet.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatalf("%v", err)
	}

	buf := make([]byte, 512)

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

	// need to close the telnet connection... or else!
	t.Close()
}

func CheckClientMAC(iface, macAddr string) (bool, string) {
	host, port := findClientPort()
	t, err := telnet.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatalf("%v", err)
	}

	buf := make([]byte, 512)
	// show the link information
	buf = []byte("ip link show\r\n")
	_, err = t.Write(buf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// read out bytes until '%' to clear output for this command
	buf, err = t.ReadBytes('%')
	if err != nil {
		log.Fatalf("%v", err)
	}

	// close the breach (telnet)
	t.Close()

	// actually test what we expect should be the mac address
	mac := getLinkMAC(iface, buf)
	if mac == macAddr {
		return true, mac
	} else {
		return false, mac
	}

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

// may fail given multiple serial types
func findClientPort() (host, port string) {
	content, err := ioutil.ReadFile(".rvn/dom_sled-basic_client.xml")
	if err != nil {
		log.Fatal(err)
	}
	// TODO: do a for each compile, check <protocol type="telnet"></protocol>
	Re := regexp.MustCompile("(?s)<serial type=\"tcp\">.+</serial>")
	subString := Re.FindString(string(content))
	//log.Printf("%s", subString)
	Re = regexp.MustCompile("host=\"(?P<host>[a-z]*)\"")
	host = Re.FindStringSubmatch(subString)[1]
	Re = regexp.MustCompile("service=\"(?P<service>[0-9]*)\"")
	port = Re.FindStringSubmatch(subString)[1]
	return
}

package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/ziutek/telnet"
	"io/ioutil"
	"regexp"
	"strings"
)

/*
 * This should be run after running:
 * sudo -E rvn build
 * sudo -E rvn deploy
 * sudo -E rvn pingwait server
 * sudo -E rvn configure
 * now, client and server are ready, you can run this function
 * to unit test setting the client MAC address
 */

func main() {
	iface := "eth1"
	macAddr := "00:00:00:00:00:01"
	ipAddr := "10.0.0.2"

	log.SetLevel(log.DebugLevel)

	// set the mac address
	SetClientMAC(iface, macAddr)

	// test that the mac address was set correctly
	success, mac := CheckClientMAC(iface, macAddr)
	log.Printf("%v %v", success, mac)

	// set the client interface UP
	SetClientIfaceUP(iface)

	// test that the client interface is up
	success, link := CheckClientIfaceUP(iface)
	log.Printf("%v %v", success, link)

	// set the client IP address
	SetClientIP(iface, ipAddr)

	//test that the client IP address was set correctly
	success, ip := CheckClientIP(iface, ipAddr)
	log.Printf("%v %v", success, ip)

	//sledRet := RunSledc("10.0.0.1")
	//log.Printf("%s", sledRet)
}

// ----------- RUN SLED ------------ //

func RunSledc(server string) string {
	host, port := findClientPort()
	t, err := telnet.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatalf("%v", err)
	}
	buf := make([]byte, 512)
	// show the link information
	buf = []byte(fmt.Sprintf("sledc -server %s\r\n", server))
	_, err = t.Write(buf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	buf, err = t.ReadBytes('%')
	if err != nil {
		log.Fatalf("%v", err)
	}
	return string(buf)
}

// ----------- SET CLIENT SETTINGS ------------ //

// set iface up - unit testable
func SetClientIfaceUP(iface string) {
	host, port := findClientPort()
	t, err := telnet.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatalf("%v", err)
	}

	buf := make([]byte, 512)

	// set the mac address for the interface for bolt-db
	buf = []byte(fmt.Sprintf("ip link set %s up\r\n", iface))
	_, err = t.Write(buf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// read out bytes until '%' to clear output for this command
	buf, err = t.ReadBytes('%')
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Debugln("SetClientIfaceUP: " + string(buf))
	t.Close()
}

func SetClientIP(iface, ipAddr string) {
	host, port := findClientPort()
	t, err := telnet.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatalf("%v", err)
	}

	buf := make([]byte, 512)

	// set the mac address for the interface for bolt-db
	buf = []byte(fmt.Sprintf("ip addr add %s/24 dev %s\r\n", ipAddr, iface))
	_, err = t.Write(buf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// read out bytes until '%' to clear output for this command
	buf, err = t.ReadBytes('%')
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Debugln("SetClientIP: " + string(buf))

	t.Close()
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
	buf, err = t.ReadBytes('%')
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Debugln("SetClientMAC: " + string(buf))
	// need to close the telnet connection... or else!
	t.Close()
}

// ----------- VERIFY CLIENT SETTINGS ------------ //

func CheckClientMAC(iface, macAddr string) (bool, string) {
	buf := make([]byte, 512)
	buf = getClientLinkInfo()

	// actually test what we expect should be the mac address
	mac := getLinkMAC(iface, buf)
	if mac == macAddr {
		return true, mac
	} else {
		return false, mac
	}
}

func CheckClientIfaceUP(iface string) (bool, string) {
	buf := make([]byte, 512)
	buf = getClientLinkInfo()

	// actually test what we expect should be the mac address
	line := getLinkIface(iface, buf)
	if line == "UP" {
		return true, line
	} else {
		return false, line
	}
}

func CheckClientIP(iface, ipAddr string) (bool, string) {
	buf := make([]byte, 512)
	buf = getClientAddrInfo()

	// actually test what we expect should be the mac address
	ip := getLinkIP(iface, buf)
	if ip == ipAddr {
		return true, ip
	} else {
		return false, ip
	}
}

// ----------- HELPER FUNCTIONS -------------- //

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

// given ip link output, get the mac address for an interface
func getLinkIface(iface string, buf []byte) string {
	// find the index that has the interface
	indexBegin := strings.Index(string(buf), iface)
	// find the first new line, start here
	indexFirst := strings.Index(string(buf[indexBegin:]), "\n")
	// find the next new line, this is the new region we care about
	statusLine := string(buf[indexBegin+1 : indexBegin+indexFirst])
	// get the mac address
	Re := regexp.MustCompile("(UP|DOWN)")
	return Re.FindString(statusLine)
}

// given ip link output, get the mac address for an interface
func getLinkIP(iface string, buf []byte) string {
	// find the index that has the interface
	indexBegin := strings.Index(string(buf), iface)
	// find the first new line, skip it, find next
	indexSkip := strings.Index(string(buf[indexBegin:]), "\n")
	indexFirst := strings.Index(string(buf[indexBegin+indexSkip+1:]), "\n")
	// all the +1 are to move beyond the newlines that are found.
	indexFirst = indexBegin + indexSkip + indexFirst + 1
	// find the next new line, this is the new region we care about
	indexLast := strings.Index(string(buf[indexFirst+1:]), "\n")
	ipLine := string(buf[indexFirst+1 : indexFirst+indexLast])
	// find the ip address with the inet, partly to avoid finding broadcast
	Re := regexp.MustCompile("inet (?P<ip>[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})")
	// instance 0 is the string itself, instance 1 is the ip address
	return FindStringSubmatch(ipLine)[1]
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

func getClientLinkInfo() []byte {
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
	return buf
}

func getClientAddrInfo() []byte {
	host, port := findClientPort()
	t, err := telnet.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatalf("%v", err)
	}

	buf := make([]byte, 512)
	// show the link information
	buf = []byte("ip addr\r\n")
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
	return buf
}

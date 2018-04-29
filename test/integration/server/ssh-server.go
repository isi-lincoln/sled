package server

import (
	"fmt"
	"github.com/ceftb/sled/test/integration/shared"
	"github.com/rcgoodfellow/raven/rvn"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"regexp"
	"strings"
)

/*
* because this imports raven, requires root to run
 */

func main() {
	iface := shared.ServerIface
	macAddr := shared.ClientMAC
	ipAddr := shared.ServerIP

	SetServerIface(iface)

	success, link := CheckServerIface(iface)
	log.Printf("%v %v", success, link)

	SetServerIP(iface, ipAddr)

	success, ip := CheckServerIP(iface, ipAddr)
	log.Printf("%v %v", success, ip)

	success, boltEntry := SetAndCheckBoltDB(macAddr)
	log.Printf("%v %v", success, boltEntry)
}

// ----------- SET SERVER SETTINGS ------------ //

func SetServerIface(iface string) {
	serverIP := getRavenIP("server")
	out, err := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", "/var/rvn/ssh/rvn", fmt.Sprintf("rvn@%s", serverIP), fmt.Sprintf("sudo ip link set %s up", iface)).Output()
	if err != nil {
		log.Fatalf("%v : %s", err, string(out))
	}
}

// ip addr add 10.0.0.1/24 dev eth1
func SetServerIP(iface, ip string) {
	serverIP := getRavenIP("server")
	out, err := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", "/var/rvn/ssh/rvn", fmt.Sprintf("rvn@%s", serverIP), fmt.Sprintf("sudo ip addr add %s/24 dev %s", ip, iface)).Output()
	if err != nil {
		log.Fatalf("%v : %s", err, string(out))
	}
}

func SetAndCheckBoltDB(macAddr string) (bool, string) {
	serverIP := getRavenIP("server")
	out, err := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", "/var/rvn/ssh/rvn", fmt.Sprintf("rvn@%s", serverIP), fmt.Sprintf("sudo %s", shared.BoltDBPath)).Output()
	if err != nil {
		log.Fatalf("%v : %s", err, string(out))
	}
	Re := regexp.MustCompile("([0-9a-f][0-9a-f]:){5}[0-9a-f][0-9a-f]")
	// instance 0 is the string itself, instance 1 is the ip address
	mac := Re.FindString(string(out))
	if mac == macAddr {
		return true, mac
	} else {
		return false, mac
	}
}

// ----------- CHECK SERVER FUNCTIONS -------------- //

func CheckServerIface(iface string) (bool, string) {
	serverIP := getRavenIP("server")
	out, err := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", "/var/rvn/ssh/rvn", fmt.Sprintf("rvn@%s", serverIP), fmt.Sprintf("sudo ip addr show %s", iface)).Output()
	if err != nil {
		log.Fatalf("%v : %s", err, string(out))
	}

	link := getServerIface(iface, out)
	if link == "UP" {
		return true, link
	} else {
		return false, link
	}

}

func CheckServerIP(iface, ip string) (bool, string) {
	serverIP := getRavenIP("server")
	out, err := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", "/var/rvn/ssh/rvn", fmt.Sprintf("rvn@%s", serverIP), fmt.Sprintf("sudo ip addr show %s", iface)).Output()
	if err != nil {
		log.Fatalf("%v : %s", err, string(out))
	}

	ipAddr := getServerAddr(iface, out)
	if ip == ipAddr {
		return true, ipAddr
	} else {
		return false, ipAddr
	}

}

// ----------- HELPER FUNCTIONS -------------- //

func getRavenIP(node string) string {
	topo, err := rvn.LoadTopo()
	if err != nil {
		log.Fatal(err)
	}
	ds, err := rvn.DomainStatus(topo.Name, node)
	if err != nil {
		log.Fatal(err)
	}
	return ds.IP
}

func getServerAddr(iface string, buf []byte) string {
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
	return Re.FindStringSubmatch(ipLine)[1]

}

func getServerIface(iface string, buf []byte) string {
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

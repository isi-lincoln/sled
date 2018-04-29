package integration

import (
	"errors"
	"fmt"
	"github.com/ceftb/sled/test/integration/client"
	"github.com/ceftb/sled/test/integration/server"
	"github.com/ceftb/sled/test/integration/shared"
	"github.com/rcgoodfellow/raven/rvn"
	log "github.com/sirupsen/logrus"
	"github.com/sparrc/go-ping" // pingwait
	"os"
	"testing"
	"time"
)

func setupServer() error {
	macAddr := shared.ClientMAC
	iface := shared.ServerIface
	ipAddr := shared.ServerIP

	server.SetServerIface(iface)

	success, link := server.CheckServerIface(iface)
	if !success {
		return errors.New(fmt.Sprintf("Server interface %s - wanted: UP, got: %v", iface, link))
	}
	server.SetServerIP(iface, ipAddr)

	success, ip := server.CheckServerIP(iface, ipAddr)
	if !success {
		return errors.New(fmt.Sprintf("Server IP not set! wanted: %v, got: %v", ipAddr, ip))
	}

	_, err := os.Stat(shared.LocalBoltPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Bolt DB executable missing at: %s", shared.LocalBoltPath))
	}

	success, boltEntry := server.SetAndCheckBoltDB(macAddr)
	if !success {
		return errors.New(fmt.Sprintf("Server BOLTDB not set! wanted: %v, got: %v", macAddr, boltEntry))
	}
	return nil
}

func setupClient() error {
	macAddr := shared.ClientMAC
	iface := shared.ClientIface
	ipAddr := shared.ClientIP

	client.SetClientMAC(iface, macAddr)
	success, mac := client.CheckClientMAC(iface, macAddr)
	if !success {
		return errors.New(fmt.Sprintf("Client MAC not set! wanted: %v, got: %v", macAddr, mac))
	}
	client.SetClientIfaceUP(iface)

	// test that the client interface is up
	success, link := client.CheckClientIfaceUP(iface)
	if !success {
		return errors.New(fmt.Sprintf("Client interface %s - wanted: UP, got: %v", iface, link))
	}

	// set the client IP address
	client.SetClientIP(iface, ipAddr)

	//test that the client IP address was set correctly
	success, ip := client.CheckClientIP(iface, ipAddr)
	if !success {
		return errors.New(fmt.Sprintf("Client IP not set! wanted: %v, got: %v", ip, ipAddr))
	}
	return nil
}

func TestRvnBootSimple(t *testing.T) {
	log.Infof("Testing: Starting Raven configuration sledc - sledd")
	// bring up and do all the raven initalization
	err := startRaven()
	if err != nil {
		t.Errorf("Unable to start Raven: %v", err)
	}

	// setup the client and server to communicate
	err = setupClient()
	if err != nil {
		t.Errorf("%v", err)
	}
	err = setupServer()
	if err != nil {
		t.Errorf("%v", err)
	}

	client.RunSledc(shared.ServerIP)
	// run sledc on client

	// tear down all the raven configuration
	err = stopRaven()
	if err != nil {
		t.Errorf("Unable to stop Raven: %v", err)
	}
}

// ~~~~~~~~~~~~ RVN CLI ~~~~~~~~~~~~~~

// modified ping wait to wait until the hosts are up, once they are
// reachable, we can configure the hosts.
func doPingwait() {

	ipmap := make(map[string]string)
	var nodes map[string]rvn.DomStatus

	// first wait until everything we need to ping has an IP
	success := false
	for !success {

		success = true

		status := rvn.Status()
		if status == nil {
			log.Fatal("could not query libvirt status")
		}

		nodes = status["nodes"].(map[string]rvn.DomStatus)

		for name, element := range nodes {
			ipmap[name] = element.IP
		}

	}

	// now try to ping everything
	success = false

	for !success {

		success = true
		for x := range nodes {
			success = success && doPing(ipmap[x])
		}

	}

}

func doPing(host string) bool {

	p, err := ping.NewPinger(host)
	if err != nil {
		log.Fatal(err)
	}
	p.Count = 2
	p.Timeout = time.Millisecond * 500
	p.Interval = time.Millisecond * 50
	pings := 0
	p.OnRecv = func(pkt *ping.Packet) {
		pings++
	}
	p.Run()

	return pings == 2

}

func startRaven() error {
	// lets start by creating rvn info
	err := rvn.RunModel()
	if err != nil {
		return err
	}
	// do libvirt things, atm does not return error
	rvn.Create()

	// now we can start to use our nodes, bring raven online
	strErr := rvn.Launch()
	if len(strErr) != 0 {
		for _, e := range strErr {
			log.Println(e)
		}
		// TODO: another hack, just return the first of many errors
		return errors.New(strErr[0])
	}

	// wait until the nodes are ready to be setup
	doPingwait()

	// configure the nodes -- set up mounts, again no error
	rvn.Configure(true)

	// wait until the nodes are ready to be setup
	doPingwait()

	return nil
}

func stopRaven() error {
	err := rvn.Shutdown()
	if err != nil {
		for _, e := range err {
			log.Printf("%v", e)
		}
		// FIXME: short hack, just return first of the errors
		return err[0]
	}

	// remove all info/details on the raven test configuration
	rvn.Destroy()
	return nil
}

func getNodeIP(node string) string {
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

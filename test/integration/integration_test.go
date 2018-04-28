package integration

import (
	"errors"
	"github.com/rcgoodfellow/raven/rvn"
	log "github.com/sirupsen/logrus"
	"github.com/sparrc/go-ping" // pingwait
	"testing"
	"time"
)

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
	return nil

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

func TestRvnBootSimple(t *testing.T) {
	log.Infof("Testing: Starting Raven configuration sledc - sledd")
	// bring up and do all the raven initalization
	err := startRaven()
	if err != nil {
		t.Errorf("Unable to start Raven: %v", err)
	}
	// do test here ~~~~~~`

	macAddr := "00:00:00:00:00:01"
	SetClientMAC("eth1", macAddr)
	success, mac := CheckClientMAC("eth1", macAddr)
	if !success {
		t.Errorf("Client MAC not set! wanted: %v, got: %v", macAddr, mac)
	}

	// get the server IP
	serverIP := getNodeIP("server")

	// do hacky ssh exec - would like an ssh pipe
	// this is only for the server IP, which is necessary to run bolt update on server
	// TODO: test this
	cmd := exec.Command(
		"ssh -o StrictHostKeyChecking=no -i /var/rvn/ssh/rvn rvn@%s /tmp/code/test/integration/bolt-update", serverIP)

	// do telnet to connection to client
	// TODO: implement this

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

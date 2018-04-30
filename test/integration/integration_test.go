package integration

import (
	"errors"
	"fmt"
	"github.com/ceftb/sled/test/integration/client"
	"github.com/ceftb/sled/test/integration/server"
	"github.com/ceftb/sled/test/integration/shared"
	"github.com/rcgoodfellow/raven/rvn"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"testing"
	//"time"
)

func setupServer() error {
	macAddr := shared.ClientMAC
	iface := shared.ServerIface
	ipAddr := shared.ServerIP

	log.Infof("Server: Setting interface UP")
	err := server.SetServerIface(iface)

	success, link := server.CheckServerIface(iface)
	if !success || err != nil {
		return errors.New(fmt.Sprintf("Server interface %s - wanted: UP, got: %v", iface, link))
	}

	log.Infof("Server: Setting IP Address")
	err = server.SetServerIP(iface, ipAddr)

	success, ip := server.CheckServerIP(iface, ipAddr)
	if !success || err != nil {
		return errors.New(fmt.Sprintf("Server IP not set! wanted: %v, got: %v", ipAddr, ip))
	}

	log.Infof("Server: Creating BoltDB Entry")
	_, err = os.Stat(shared.LocalBoltPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Bolt DB executable missing at: %s", shared.LocalBoltPath))
	}

	err, success, boltEntry := server.SetAndCheckBoltDB(macAddr)
	if !success || err != nil {
		return errors.New(fmt.Sprintf("Server BOLTDB not set! wanted: %v, got: %v", macAddr, boltEntry))
	}
	return nil
}

func setupClient() error {
	macAddr := shared.ClientMAC
	iface := shared.ClientIface
	ipAddr := shared.ClientIP

	log.Infof("Client: Setting MAC Address")
	client.SetClientMAC(iface, macAddr)
	success, mac := client.CheckClientMAC(iface, macAddr)
	if !success {
		return errors.New(fmt.Sprintf("Client MAC not set! wanted: %v, got: %v", macAddr, mac))
	}

	log.Infof("Client: Setting interface UP")
	client.SetClientIfaceUP(iface)
	// test that the client interface is up
	success, link := client.CheckClientIfaceUP(iface)
	if !success {
		return errors.New(fmt.Sprintf("Client interface %s - wanted: UP, got: %v", iface, link))
	}

	log.Infof("Client: Setting IP Address")
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
		t.Fatalf("Unable to start Raven: %v", err)
	}

	// setup the client and server to communicate
	err = setupClient()
	if err != nil {
		stopRaven()
		t.Fatalf("%v", err)
	}
	err = setupServer()
	if err != nil {
		stopRaven()
		t.Fatalf("%v", err)
	}

	log.Infof("Server: Starting Sledd")
	// run sledc on client
	err = server.RunSledd()
	if err != nil {
		stopRaven()
		t.Fatalf("%v", err)
	}

	log.Infof("Client: Running Sledc")
	// run sledc on client
	sledRet := client.RunSledc(shared.ServerIP)
	log.Infof("%s", sledRet)

	//time.Sleep(time.Minute * 2)

	// tear down all the raven configuration
	err = stopRaven()
	if err != nil {
		t.Fatalf("Unable to stop Raven: %v", err)
	}
}

// ~~~~~~~~~~~~ HELPER FUNCTIONS ~~~~~~~~~~~~~~

func startRaven() error {
	log.Infof("Building Raven Topology")
	out, err := exec.Command("sudo", "-E", "rvn", "build").Output()
	if err != nil {
		return errors.New(fmt.Sprintf("%v : %s", err, string(out)))
	}
	log.Infof("Deploying Raven Topology")
	out, err = exec.Command("sudo", "-E", "rvn", "deploy").Output()
	if err != nil {
		return errors.New(fmt.Sprintf("%v : %s", err, string(out)))
	}
	log.Infof("Waiting on Raven Topology")
	out, err = exec.Command("sudo", "-E", "rvn", "pingwait", "server").Output()
	if err != nil {
		return errors.New(fmt.Sprintf("%v : %s", err, string(out)))
	}
	log.Infof("Configuring Raven Topology")
	out, err = exec.Command("sudo", "-E", "rvn", "configure").Output()
	if err != nil {
		return errors.New(fmt.Sprintf("%v : %s", err, string(out)))
	}

	return nil
}

func stopRaven() error {
	log.Infof("Tearing down Raven Topology")
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

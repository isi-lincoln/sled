package main

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"math"
	"net"

	"github.com/ceftb/sled"
	sledc "github.com/ceftb/sled/sledc/pkg"
)

var server = flag.String("server", "sled", "sled server to connect to")
var ifx = flag.String("interface", "eth0", "the interface to use for client id")

func main() {

	flag.Parse()

	conn, sledd := initClient()
	defer conn.Close()

	ifxs, err := net.Interfaces()
	if err != nil {
		log.Fatalf("error getting interface info - %v", err)
	}
	if len(ifxs) < 1 {
		log.Fatalf("no interfaces!")
	}
	mac := ""
	for _, x := range ifxs {
		if x.Name == *ifx {
			mac = x.HardwareAddr.String()
		}
	}
	if mac == "" {
		log.Fatalf("interface %s not found", *ifx)
	}

	resp, err := sledd.Command(context.TODO(), &sled.CommandRequest{mac})
	if err != nil {
		log.Fatalf("error getting sledd command - %v", err)
	}

	if resp.Wipe != nil {
		sledc.WipeBlock(resp.Wipe.Device)
	}
	if resp.Write != nil {
		sledc.Write(resp.Write.Device, resp.Write.Image, resp.Write.Kernel, resp.Write.Initrd)
	}
	if resp.Kexec != nil {
		sledc.Kexec(resp.Kexec.Kernel, resp.Kexec.Append, resp.Kexec.Initrd)
	}
	if resp.Wipe == nil && resp.Write == nil && resp.Kexec == nil {
		log.Warn("received empty command from server")
	}
}

// 8 GB max image size$
const maxMsgSize = math.MaxUint32

func initClient() (*grpc.ClientConn, sled.SledClient) {
	conn, err := grpc.Dial(
		*server+":6000",
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("could not connect to sled server - %v", err)
	}

	return conn, sled.NewSledClient(conn)
}

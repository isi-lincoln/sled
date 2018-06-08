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

	// the Command function now returns which functions the client should request
	resp, err := sledd.Command(context.TODO(), &sled.CommandRequest{mac})
	if err != nil {
		log.Fatalf("error getting sledd command - %v", err)
	}

	if resp.Wipe != nil {
		wipe, err := sledd.Wipe(context.TODO(), &sled.WipeRequest{mac})
		if err != nil {
			log.Fatalf("error getting wipe command - %v", err)
		}
		if wipe != nil {
			sledc.WipeBlock(wipe.Device)
		}
	}
	// write is the message that would hold the fileystem, kernel, initrd
	// protobuf is not efficient for large file transfers because of serialization
	// so for write we will use raw sockets to minimize memory footprint
	if resp.Write != nil {
		write, err := sledd.Write(context.TODO(), &sled.WriteRequest{mac})
		if err != nil {
			log.Fatalf("error getting write command - %v", err)
		}
		var images []string
		if write.Image != "" {
			images.append(write.Image)
		}
		if write.Kernel != "" {
			images.append(write.Kernel)
		}
		if write.Initrd != "" {
			images.append(write.Initrd)
		}
		if len(images) > 0 {
			err = sledc.WriteCommunicator(*server, images)
			if err != nil {
				log.Fatalf("error communicating with sledd - %v", err)
			}
		} else {
			log.Infof("Client recieved empty write response - %v", write)
		}
	}
	if resp.Kexec != nil {
		kexec, err := sledd.Kexec(context.TODO(), &sled.KexecRequest{mac})
		if err != nil {
			log.Fatalf("error getting kexec command - %v", err)
		}
		if wipe != nil {
			sledc.Kexec(kexec.Kernel, kexec.Append, kexec.Initrd)
		}
	}
	if resp.Wipe == "" && resp.Write == "" && resp.Kexec == "" {
		log.Warn("received empty command from server")
	}
}

// FIXME: Add certificates for authenticated communitcation client <---> server
func initClient() (*grpc.ClientConn, sled.SledClient) {
	conn, err := grpc.Dial(
		*server+":6000",
		grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to sled server - %v", err)
	}

	return conn, sled.NewSledClient(conn)
}

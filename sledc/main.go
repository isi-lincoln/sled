package main

import (
	"context"
	"flag"
	"github.com/ceftb/sled"
	sledc "github.com/ceftb/sled/sledc/pkg"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
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
		if resp.Wipe.Device != "" {
			sledc.WipeBlock(resp.Wipe.Device)
		} else {
			log.Errorf("Wipe: Empty Device String given.")
		}
	}
	// write is the message that would hold the fileystem, kernel, initrd
	// protobuf is not efficient for large file transfers because of serialization
	// so for write we will use raw sockets to minimize memory footprint
	if resp.Write != "" {
		write, err := sledd.Write(context.TODO(), &sled.WriteRequest{mac})
		if err != nil {
			log.Fatalf("error getting write command - %v", err)
		}
		var images []string
		if write.Image != "" {
			images = append(images, write.Image)
		}
		if write.Kernel != "" {
			images = append(images, write.Kernel)
		}
		if write.Initrd != "" {
			images = append(images, write.Initrd)
		}
		if len(images) > 0 {
			err = sledc.WriteCommunicator(*server, mac, images)
			if err != nil {
				log.Errorf("error communicating with sledd - %v", err)
			}
		} else {
			log.Infof("Client recieved empty write response - %v", write)
		}
	}
	if resp.Kexec != nil {
		sledc.Kexec(resp.Kexec.Kernel, resp.Kexec.Append, resp.Kexec.Initrd)
	}
	if resp.Wipe == nil && resp.Write == "" && resp.Kexec == nil {
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

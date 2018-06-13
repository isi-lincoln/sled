package main

import (
	"context"
	"github.com/isi-lincoln/sled"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {

	conn, sledd := initClient()
	defer conn.Close()
	mac := "localhost"

	// the Command function now returns which functions the client should request
	resp, err := sledd.Command(context.TODO(), &sled.CommandRequest{mac})
	if err != nil {
		log.Fatalf("error getting sledd command - %v", err)
	}

	log.Infof("got: %v", resp)

	if resp.Wipe != nil {
		if resp.Wipe.Device != "" {
			log.Infof("Fake wipe: %s", resp.Wipe.Device)
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
		log.Infof("Write Images: %s", images)
	}
	if resp.Kexec != nil {
		log.Infof("kernel: %s, cmd: %s, initrd: %s", resp.Kexec.Kernel, resp.Kexec.Append, resp.Kexec.Initrd)
	}
	if resp.Wipe == nil && resp.Write == "" && resp.Kexec == nil {
		log.Warn("received empty command from server")
	}
}

// FIXME: Add certificates for authenticated communitcation client <---> server
func initClient() (*grpc.ClientConn, sled.SledClient) {
	conn, err := grpc.Dial(
		"localhost:6000",
		grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to sled server - %v", err)
	}

	return conn, sled.NewSledClient(conn)
}

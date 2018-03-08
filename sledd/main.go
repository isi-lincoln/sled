package main

import (
	"context"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/ceftb/sled"
)

type Sledd struct{}

func (s *Sledd) Command(
	ctx context.Context, e *sled.CommandRequest,
) (*sled.CommandResponse, error) {

	return nil, fmt.Errorf("not implemented")
}

func main() {
	fmt.Println("sled-server")

	grpcServer := grpc.NewServer()
	sled.RegisterSledServer(grpcServer, &Sledd{})

	l, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatalf("failed to listen: %#v", err)
	}

	log.Info("Listening on tcp://0.0.0.0:6000")
	grpcServer.Serve(l)

}

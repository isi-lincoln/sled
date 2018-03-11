package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	bolt "github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/ceftb/sled"
)

type Sledd struct{}

func (s *Sledd) Command(
	ctx context.Context, e *sled.CommandRequest,
) (*sled.CommandSet, error) {

	log.Printf("command %#v", e)

	cs := &sled.CommandSet{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("clients"))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(e.Mac))
		if v == nil {
			return nil
		} else {
			var cs sled.CommandSet
			err := json.Unmarshal(v, cs)
			if err != nil {
				log.Errorf("command: malformed command set @ %s", e.Mac)
				return nil
			}
			return nil
		}
	})

	if cs.Write != nil {
		filename := fmt.Sprintf("/var/img/%s", cs.Write.Name)
		_, err := os.Stat(filename)
		if err != nil {
			log.Errorf("command: non-existant write image %s", cs.Write.Name)
			cs.Write = nil
		}
		cs.Write.Image, err = ioutil.ReadFile(filename)
		if err != nil {
			log.Errorf("command: error reading image %v", err)
		}
	}

	return cs, nil

}

func (s *Sledd) Update(
	ctx context.Context, e *sled.UpdateRequest,
) (*sled.UpdateResponse, error) {

	log.Printf("update %#v", e)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("clients"))
		if b == nil {
			return fmt.Errorf("update: no client bucket")
		}
		js, err := json.Marshal(e.CommandSet)
		if err != nil {
			return fmt.Errorf("update: failed to serialize command set %v", err)
		}
		err = b.Put([]byte(e.Mac), js)
		if err != nil {
			return fmt.Errorf("update: failed to put command set %v", err)
		}
		return nil
	})

	if err != nil {
		log.Printf("update: failed - %v", err)
	}

	return &sled.UpdateResponse{
		Success: err != nil,
		Message: err.Error(),
	}, nil
}

var db *bolt.DB

func main() {
	fmt.Println("sled-server")

	grpcServer := grpc.NewServer()
	sled.RegisterSledServer(grpcServer, &Sledd{})

	l, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatalf("failed to listen: %#v", err)
	}

	db, err = bolt.Open("/var/sled.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("clients"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	log.Info("Listening on tcp://0.0.0.0:6000")
	grpcServer.Serve(l)

}

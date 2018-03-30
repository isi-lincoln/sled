package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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
			err := json.Unmarshal(v, cs)
			if err != nil {
				log.Errorf("command: malformed command set @ %s", e.Mac)
				log.Error(err)
				return nil
			}
			return nil
		}
	})

	// for image, kernel, initramfs, load each of them into memory based
	// on the path stored in the bolt db
	if cs.Write != nil {
		imageName := fmt.Sprintf("/var/img/%s", cs.Write.ImageName)
		_, err := os.Stat(imageName)
		if err != nil {
			log.Errorf("command: non-existant write image %s", cs.Write.ImageName)
			cs.Write = nil
		}
		cs.Write.Image, err = ioutil.ReadFile(imageName)
		if err != nil {
			log.Errorf("command: error reading image %v", err)
		}

		kernelName := fmt.Sprintf("/var/img/%s", cs.Write.KernelName)
		_, err = os.Stat(kernelName)
		if err != nil {
			log.Errorf("command: non-existant kernel %s", cs.Write.KernelName)
			cs.Write = nil
		}
		cs.Write.Kernel, err = ioutil.ReadFile(kernelName)
		if err != nil {
			log.Errorf("command: error reading kernel %v", err)
		}

		initrdName := fmt.Sprintf("/var/img/%s", cs.Write.InitrdName)
		_, err = os.Stat(initrdName)
		if err != nil {
			log.Errorf("command: non-existant initramfs %s", cs.Write.InitrdName)
			cs.Write = nil
		}
		cs.Write.Initrd, err = ioutil.ReadFile(initrdName)
		if err != nil {
			log.Errorf("command: error reading initramfs %v", err)
		}

	}

	return cs, nil

}

func (s *Sledd) Update(
	ctx context.Context, e *sled.UpdateRequest,
) (*sled.UpdateResponse, error) {

	log.Printf("update %#v", e)

	err := db.Update(func(tx *bolt.Tx) error {

		// grab a reference to the clients bucket
		b := tx.Bucket([]byte("clients"))
		if b == nil {
			return fmt.Errorf("update: no client bucket")
		}

		// get the current value associated with the specified mac
		var current *sled.CommandSet
		v := b.Get([]byte(e.Mac))
		if v != nil {
			err := json.Unmarshal(v, current)
			if err != nil {
				return fmt.Errorf("command: malformed command set @ %s", e.Mac)
			}
		}

		// merge the update into the current command set
		updated := csMerge(current, e.CommandSet)

		// psersist the update
		js, err := json.Marshal(updated)
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

func csMerge(current, update *sled.CommandSet) *sled.CommandSet {
	result := &sled.CommandSet{}
	*result = *update
	if current == nil {
		return result
	}
	if update.Wipe != nil {
		result.Wipe = &sled.Wipe{}
		*result.Wipe = *update.Wipe
	}
	if update.Write != nil {
		result.Write = &sled.Write{}
		*result.Write = *update.Write
	}
	if update.Kexec != nil {
		result.Kexec = &sled.Kexec{}
		*result.Kexec = *update.Kexec
	}
	return result
}

var db *bolt.DB

// 8 GB max image size
const maxMsgSize = math.MaxUint32

func main() {
	fmt.Println("sled-server")

	// overwrite grpc Msg sizes to allow larger images to be sent
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize))
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
		_, err := tx.CreateBucketIfNotExists([]byte("clients"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	log.Info("Listening on tcp://0.0.0.0:6000")
	grpcServer.Serve(l)
}

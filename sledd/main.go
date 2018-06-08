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

func (s *Sledd) Wipe(
	ctx context.Context, e *sled.WipeRequest,
) (*sled.Wipe, error) {

	log.Printf("wipe %#v", e)
	cs := boltLookup(e.Mac)
	return cs.Wipe, nil
}

func (s *Sledd) Kexec(
	ctx context.Context, e *sled.KexecRequest,
) (*sled.Kexec, error) {

	log.Printf("kexec %#v", e)
	cs := boltLookup(e.Mac)
	return cs.Kexec, nil
}

func (s *Sledd) Write(
	ctx context.Context, e *sled.WriteRequest,
) (*sled.Write, error) {

	log.Printf("kexec %#v", e)
	cs := boltLookup(e.Mac)
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

	return cs.Write, nil
}

func (s *Sledd) Command(
	ctx context.Context, e *sled.CommandRequest,
) (*sled.CommandSet, error) {
	log.Printf("command %#v", e)
	cs := boltLookup(e.Mac)
	if cs.Write == nil {
		cs.Write = ""
	} else {
		cs.Write = "1"
	}
	if cs.Wipe == nil {
		cs.Wipe = ""
	} else {
		cs.Wipe = "1"
	}
	if cs.Kexec == nil {
		cs.Kexec = ""
	} else {
		cs.Kexec = "1"
	}
	// for image, kernel, initramfs, load each of them into memory based
	// on the path stored in the bolt db
	/*

	 */

	return cs, nil

}

func (s *Sledd) Wwrite(
	ctx context.Context, e *sled.CommandRequest,
) (*sled.CommandSet, error) {
	writeCmd := boltLookup(e.Mac)

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

func main() {
	fmt.Println("Starting sled-server.")

	// overwrite grpc Msg sizes to allow larger images to be sent
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize))
	sled.RegisterSledServer(grpcServer, &Sledd{})

	protobufServer, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatalf("protobuf failed to listen: %#v", err)
	}

	writeServer, err := net.Listen("tcp", "0.0.0.0:3000")
	if err != nil {
		log.Fatalf("write server failed to listen: %#v", err)
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

	// spin this off as a goroutine
	log.Info("Listening on tcp://0.0.0.0:6000")
	go grpcServer.Serve(protobufServer)

	defer writeServer.Close()
	log.Info("Listening on tcp://0.0.0.0:3000")
	for {
		// Listen for an incoming connection.
		conn, err := writeServer.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go sendBytes(conn)
	}

}

/*             Helper Functions                         */

func sendBytes(conn net.Conn) {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

}

func boltLookup(mac string) *sled.CommandSet {
	cs := &sled.CommandSet{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("clients"))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(mac))
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
	return cs
}

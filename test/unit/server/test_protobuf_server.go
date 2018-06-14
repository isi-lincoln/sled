package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"

	bolt "github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/ceftb/sled"
)

var db *bolt.DB
var clientPrefix string = "."
var localBoltServer string = "./bolt.db"
var localBoltPerms os.FileMode = 0600

type Sledd struct{}

/* PROTOBUF MESSAGES */

func (s *Sledd) Wipe(
	ctx context.Context, e *sled.WipeRequest,
) (*sled.WipeResponse, error) {
	log.Infof("wipe %#v", e)
	var cs *sled.CommandSet
	cs = boltLookup(e.Mac)
	wr := &sled.WipeResponse{}
	wr.Wipe = cs.Wipe
	return wr, nil
}

func (s *Sledd) Kexec(
	ctx context.Context, e *sled.KexecRequest,
) (*sled.KexecResponse, error) {
	log.Infof("kexec %#v", e)
	var cs *sled.CommandSet
	cs = boltLookup(e.Mac)
	kr := &sled.KexecResponse{}
	kr.Kexec = cs.Kexec
	return kr, nil
}

func (s *Sledd) Write(
	ctx context.Context, e *sled.WriteRequest,
) (*sled.WriteResponse, error) {
	log.Infof("write %#v", e)
	var cs *sled.CommandSet
	cs = boltLookup(e.Mac)
	wr := &sled.WriteResponse{}
	if cs.Write != nil {
		imageName := fmt.Sprintf("%s/%s", clientPrefix, cs.Write.ImageName)
		_, err := os.Stat(imageName)
		if err != nil {
			log.Errorf("command: non-existant write image %s", cs.Write.ImageName)
			return nil, err
		}
		wr.Image = imageName
		kernelName := fmt.Sprintf("%s/%s", clientPrefix, cs.Write.KernelName)
		_, err = os.Stat(kernelName)
		if err != nil {
			log.Errorf("command: non-existant kernel %s", cs.Write.KernelName)
			return nil, err
		}
		wr.Kernel = kernelName
		initrdName := fmt.Sprintf("%s/%s", clientPrefix, cs.Write.InitrdName)
		_, err = os.Stat(initrdName)
		if err != nil {
			log.Errorf("command: non-existant initramfs %s", cs.Write.InitrdName)
			return nil, err
		}
		wr.Initrd = initrdName
	}
	return wr, nil
}

func (s *Sledd) Command(
	ctx context.Context, e *sled.CommandRequest,
) (*sled.PartialCommandSet, error) {
	log.Infof("command %#v", e)
	var cs *sled.CommandSet
	cs = boltLookup(e.Mac)
	pr := &sled.PartialCommandSet{}
	if cs.Write == nil {
		pr.Write = ""
	} else {
		pr.Write = "1"
	}
	pr.Kexec = cs.Kexec
	pr.Wipe = cs.Wipe
	return pr, nil
}

func (s *Sledd) Update(
	ctx context.Context, e *sled.UpdateRequest,
) (*sled.UpdateResponse, error) {
	var xyxx *sled.UpdateResponse
	return xyxx, nil
}

/* MAIN FUNCTION */

func main() {
	fmt.Println("Starting sled-server.")
	grpcServer := grpc.NewServer()
	sled.RegisterSledServer(grpcServer, &Sledd{})
	protobufServer, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatalf("protobuf failed to listen: %#v", err)
	}
	db, err = bolt.Open(localBoltServer, localBoltPerms, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("clients"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	db.Close()

	configureBolt()
	// spin this off as a goroutine
	log.Info("Listening on tcp://0.0.0.0:6000")
	grpcServer.Serve(protobufServer)
}

/* AUXILIARY FUNCTION(S) */

func boltLookup(mac string) *sled.CommandSet {
	log.Infof("Boltdb lookup on: %s", mac)
	db, err := bolt.Open(localBoltServer, localBoltPerms, nil)
	if err != nil {
		log.Fatal(err)
	}
	// close connection when we exit
	defer db.Close()

	cs := &sled.CommandSet{}
	err = db.View(func(tx *bolt.Tx) error {
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
				log.Errorf("command: malformed command set @ %s", mac)
				log.Error(err)
				return nil
			}
			return nil
		}
	})
	if err != nil {
		log.Error(err)
	}
	log.Infof("found: %v", cs)
	return cs
}

// test.integration.bolt.manual_update.go
func configureBolt() {
	log.Info("Configuring Bolt db")
	// open the server's sled database
	db, err := bolt.Open(localBoltServer, localBoltPerms, nil)
	if err != nil {
		log.Fatal(err)
	}
	// close connection when we exit
	defer db.Close()

	log.Info("before upadte")
	// Update is a bolt db write, only one writer at a time.
	err = db.Update(func(tx *bolt.Tx) error {
		log.Info("in upadte")
		// kill bucket (for testing)
		_ = tx.DeleteBucket([]byte("clients"))
		// bucket is like a traditional database with name clients
		bucket, err := tx.CreateBucket([]byte("clients"))
		if err != nil {
			log.Fatal(err)
		}

		// Test using ubuntu image, with fedora kernel, initramfs
		imgName := "image.txt"
		kerName := "kernel.txt"
		initName := "initrd.txt"
		// create a bogus wipe sled request
		// device is the block device name (/sys/block) not (/dev), e.g. sda
		// unlike the actual requests, the images will not be stored in the bolt db
		// which is done here to shortcut some of the process
		sledCmd := sled.CommandSet{
			&sled.Wipe{Device: "sda"},
			&sled.Write{
				ImageName:  imgName,
				Device:     "sda",
				Image:      []byte(""),
				KernelName: kerName,
				Kernel:     []byte(""),
				InitrdName: initName,
				Initrd:     []byte(""),
			},
			&sled.Kexec{
				Append: "console=ttyS1 root=/dev/sda1 rootfstype=ext4 rw",
				Kernel: "/tmp/kernel",
				Initrd: "/tmp/initrd",
			},
		}

		log.Infof("message in a bottle: ", sledCmd)

		jsonWipe, err := json.Marshal(sledCmd)
		if err != nil {
			log.Fatal("encode error:", err)
		}

		// because this is a localhost test, dont use macaddr, use "localhost" string
		clientMAC := "localhost"
		err = bucket.Put([]byte(clientMAC), jsonWipe)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// 'tis a silly thing to print db when one saves a 1gb image to it.
	// only print out the keys to verify we've added it correctly.
	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("clients"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			log.Infof("key=%s, value=%s", k, v)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

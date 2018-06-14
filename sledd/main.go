package main

import (
	"context"
	"encoding/json"
	"fmt"
	bolt "github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"net"
	"os"
	"strings" // split

	"github.com/isi-lincoln/sled"
)

type Sledd struct{}

var clientPrefix string = "/var/img"
var bufferSize int = 4096
var sledBoltDB string = "/var/sled.db"

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
		wr.Device = cs.Write.Device

		imageName := fmt.Sprintf("%s/%s", clientPrefix, cs.Write.ImageName)
		_, err := os.Stat(imageName)
		if err != nil {
			log.Errorf("command: non-existant write image %s", cs.Write.ImageName)
			return nil, err
		}
		wr.Image = cs.Write.ImageName

		kernelName := fmt.Sprintf("%s/%s", clientPrefix, cs.Write.KernelName)
		_, err = os.Stat(kernelName)
		if err != nil {
			log.Errorf("command: non-existant kernel %s", cs.Write.KernelName)
			return nil, err
		}
		wr.Kernel = cs.Write.KernelName

		initrdName := fmt.Sprintf("%s/%s", clientPrefix, cs.Write.InitrdName)
		_, err = os.Stat(initrdName)
		if err != nil {
			log.Errorf("command: non-existant initramfs %s", cs.Write.InitrdName)
			return nil, err
		}
		wr.Initrd = cs.Write.InitrdName
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

	log.Infof("update %#v", e)
	var db *bolt.DB
	db, err := bolt.Open(sledBoltDB, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
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
				log.Error("command: malformed command set @ %s", e.Mac)
				return err
			}
		}
		// merge the update into the current command set
		updated := csMerge(current, e.CommandSet)
		// psersist the update
		js, err := json.Marshal(updated)
		if err != nil {
			log.Error("update: failed to serialize command set %v", err)
			return err
		}
		err = b.Put([]byte(e.Mac), js)
		if err != nil {
			log.Error("update: failed to put command set %v", err)
			return err
		}
		return nil
	})
	db.Close()

	if err != nil {
		log.Infof("update: failed - %v", err)
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

func main() {
	log.Infof("Starting sled-server.")

	// overwrite grpc Msg sizes to allow larger images to be sent
	grpcServer := grpc.NewServer()
	sled.RegisterSledServer(grpcServer, &Sledd{})

	protobufServer, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatalf("protobuf failed to listen: %#v", err)
	}

	writeServer, err := net.Listen("tcp", "0.0.0.0:3000")
	if err != nil {
		log.Fatalf("write server failed to listen: %#v", err)
	}
	defer writeServer.Close()

	db, err := bolt.Open(sledBoltDB, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("clients"))
		if err != nil {
			log.Errorf("create bucket: %s", err)
			return err
		}
		return nil
	})
	db.Close()

	// spin this off as a goroutine
	log.Info("Listening on tcp://0.0.0.0:6000")
	go grpcServer.Serve(protobufServer)

	log.Info("Listening on tcp://0.0.0.0:3000")
	for {
		// Listen for an incoming connection.
		conn, err := writeServer.Accept()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go LowMemWrite(conn)
	}

}

/*             Helper Functions                         */
func LowMemWrite(conn net.Conn) error {
	buf := make([]byte, bufferSize)
	_, err := conn.Read(buf)
	if err != nil {
		log.Errorf("Unable to read from socket err: %v", err)
		return err
	}
	log.Infof("Received: %s", string(buf))
	msgRec := string(buf)
	// need to send the first message request the item with the associated mac
	parsedMsg := strings.Split(msgRec, ",")
	// just for sanity, remove whitespaces, add the prefix location

	req := strings.TrimSpace(parsedMsg[0])
	// conn.Read 0 pads the message, 0s need to be removed
	// 17 is 12 + semicolons, \00\00 padding removed
	macAddr := string([]byte(parsedMsg[1])[:17])
	writeCmd := boltLookup(macAddr)
	log.Infof("(%s), (%s), (%s)", req, macAddr, writeCmd)
	log.Infof("(%x), (%x), (%x)", writeCmd.Write.ImageName, writeCmd.Write.KernelName, writeCmd.Write.InitrdName)
	log.Warnf("%x", req)
	// if key in our map, then send the contents
	if req == writeCmd.Write.ImageName || req == writeCmd.Write.KernelName || req == writeCmd.Write.InitrdName {
		fmtReq := fmt.Sprintf("%s/%s", clientPrefix, strings.TrimSpace(parsedMsg[0]))
		// open file to read contents from
		fs, err := os.Open(fmtReq)
		if err != nil {
			log.Errorf("Unable to open %s, err: %v", fmtReq, err)
			return err
		}
		for {
			// read the file and send out socket while not EOF
			lenb, err := fs.Read(buf)
			if err != nil {
				// when we hit an EOF, break execution
				if err == io.EOF {
					conn.Close()
					log.Debugf("Sent file.")
					return nil
				}
				log.Errorf("Unable to read %s, err: %v", fmtReq, err)
				return err
			}
			log.Debugf("writing: %d", lenb)
			conn.Write(buf[:lenb])
		}
	}
	log.Warnf("Write image not found: %s", req)
	return nil
}

func boltLookup(mac string) *sled.CommandSet {
	cs := &sled.CommandSet{}
	var db *bolt.DB
	db, err := bolt.Open(sledBoltDB, 0600, nil)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("bolt: lookup %s", mac)
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
		log.Errorf("boltDBLookup: %v", err)
	}
	return cs
}

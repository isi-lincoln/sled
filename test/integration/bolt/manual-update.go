package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/ceftb/sled"
	"io/ioutil"
	"log"
)

/*
* This code is meant to verify the Wipe, Write, and Kexec state machine
* implemented via grpc calls from sledd to sledc.  This code is run on sledd.
 */

func main() {
	// open the server's sled database
	db, err := bolt.Open("/var/sled.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	// close connection when we exit
	defer db.Close()

	// Update is a bolt db write, only one writer at a time.
	err = db.Update(func(tx *bolt.Tx) error {
		// kill bucket (for testing)
		err = tx.DeleteBucket([]byte("clients"))
		// bucket is like a traditional database with name clients
		bucket, err := tx.CreateBucket([]byte("clients"))
		if err != nil {
			log.Fatal(err)
		}

		// Test using ubuntu image, with fedora kernel, initramfs
		imgBytes, err := ioutil.ReadFile("/var/img/ubuntu-1804-nsystemd.img")
		kerBytes, err := ioutil.ReadFile("/var/img/vmlinuz-fedora-test")
		initBytes, err := ioutil.ReadFile("/var/img/initramfs-fedora-test")
		// create a bogus wipe sled request
		// device is the block device name (/sys/block) not (/dev), e.g. sda
		// unlike the actual requests, the images will not be stored in the bolt db
		// which is done here to shortcut some of the process
		sledCmd := sled.CommandSet{
			&sled.Wipe{Device: "sda"},
			&sled.Write{
				ImageName:  "ubuntu-1804-nsystemd.img",
				Device:     "sda",
				Image:      imgBytes,
				KernelName: "vmlinuz-fedora-test",
				Kernel:     kerBytes,
				InitrdName: "initramfs-fedora-test",
				Initrd:     initBytes,
			},
			&sled.Kexec{
				Append: "console=ttyS1 root=/dev/sda1 rootfstype=ext4 rw",
				Kernel: "/tmp/kernel",
				Initrd: "/tmp/initrd",
			},
		}

		// FIXME: memory issue when encoding, wether json or gob
		// FIXME: From 1 GB to 3.2GB after sledCmd
		// FIXME: up to 11GB during the Marshing...
		jsonWipe, err := json.Marshal(sledCmd)
		if err != nil {
			log.Fatal("encode error:", err)
		}

		// put in a key-value for our mac address
		// this is eth1 mac address, the value needs to be a sled.CommandSet
		clientMAC := "00:00:00:00:00:01"
		err = bucket.Put([]byte(clientMAC), jsonWipe)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

	// 'tis a silly thing to print db when one saves a 1gb image to it.
	// only print out the keys to verify we've added it correctly.
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("clients"))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			fmt.Printf("key=%s\n", k)
		}
		return nil
	})
}

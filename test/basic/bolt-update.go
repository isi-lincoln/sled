package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/ceftb/sled"
	"io/ioutil"
	"log"
)

/*
* This code is meant to verify the Wipe, Write, and Kexec state machine
* implemented via grpc calls from sledd to sledc
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

		// Test using alpine image, kernel, initramfs
		//imgBytes, err := ioutil.ReadFile("/var/img/alpine.img")
		//kerBytes, err := ioutil.ReadFile("/var/img/alpine/hardened/vmlinuz-hardened")
		//initBytes, err := ioutil.ReadFile("/var/img/alpine/hardened/initramfs-hardened")

		// Test using ubuntu image, with fedora kernel, initramfs
		imgBytes, err := ioutil.ReadFile("/var/img/mini-ubuntu.img")
		kerBytes, err := ioutil.ReadFile("/var/img/fedora/vmlinuz-4.15.10-fedora")
		initBytes, err := ioutil.ReadFile("/var/img/fedora/initramfs-4.15.10-fedora")

		// create a bogus wipe sled request
		// device is the block device name (/sys/block) not (/dev), e.g. sda
		// unlike the actual requests, the images will not be stored in the bolt db
		// which is done here to shortcut some of the process
		sledCmd := sled.CommandSet{
			&sled.Wipe{
				Device: "sda",
			},
			&sled.Write{
				//ImageName: "alpine.img",
				ImageName: "mini-ubuntu.img",
				Device:    "sda",
				Image:     imgBytes,
				//KernelName: "alpine/hardened/vmlinuz-hardened",
				KernelName: "fedora/vmlinuz-4.15.10-fedora",
				Kernel:     kerBytes,
				//InitrdName: "alpine/hardened/initramfs-hardened",
				InitrdName: "fedora/initramfs-4.15.10-fedora",
				Initrd:     initBytes,
			},
			&sled.Kexec{
				Append: "console=ttyS1 root=/dev/sda1 rootfstype=ext4",
				Kernel: "/tmp/kernel",
				Initrd: "/tmp/initrd",
			},
		}

		// now we need to marshall it to write out
		jsonWipe, err := json.Marshal(sledCmd)
		if err != nil {
			log.Fatal(err)
		}

		// put in a key-value for our mac address
		// this is eth0 mac address, the value needs to be a sled.CommandSet
		// NOTE: this has to change every time, no way to set mac via ip link in u-root
		err = bucket.Put([]byte("52:54:00:32:eb:6a"), []byte(jsonWipe))
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
			log.Printf("key=%s\n", k)
		}
		return nil
	})
}
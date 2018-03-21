package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/ceftb/sled"
	"io/ioutil"
	"log"
)

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

		// alpine image in bytes! in memory!
		imgBytes, err := ioutil.ReadFile("/var/img/alpine.img")

		// create a bogus wipe sled request
		// device is the block device name (/sys/block) not (/dev), e.g. sda
		sledCmd := sled.CommandSet{
			&sled.Wipe{
				Device: "sda",
			},
			//&sled.Write {},
			// nil,
			&sled.Write{
				Name:   "alpine",
				Device: "sda",
				Image:  imgBytes,
			},
			//&sled.Kexec {},
			nil,
		}

		// now we need to marshall it to write out
		jsonWipe, err := json.Marshal(sledCmd)
		if err != nil {
			log.Fatal(err)
		}

		// put in a key-value for our mac address
		// this is eth0 mac address, the value needs to be a sled.CommandSet
		// NOTE: this has to change every time, no way to set mac via ip link in u-root
		err = bucket.Put([]byte("52:54:00:c6:af:75"), []byte(jsonWipe))

		// add a few other for shit and giggle
		err = bucket.Put([]byte("52:54:00:b1:64:a1"), []byte("42"))
		return nil
	})

	// 'tis a silly thing to print db when one saves a 1gb image to it.
	/*
	   db.View(func(tx *bolt.Tx) error {
	       // Assume bucket exists and has keys
	       b := tx.Bucket([]byte("clients"))

	       c := b.Cursor()

	       for k, v := c.First(); k != nil; k, v = c.Next() {
	           log.Printf("key=%s, value=%s\n", k, v)
	       }

	       return nil
	   })
	*/
}

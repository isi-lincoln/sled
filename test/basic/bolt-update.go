package main

import (
    "log"
    "encoding/json"
    "github.com/boltdb/bolt"
    "github.com/ceftb/sled"
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
        // bucket is like a traditional database with name clients
        bucket, err := tx.CreateBucketIfNotExists([]byte("clients"))
        if err != nil {
            log.Fatal(err)
        }

        // create a bogus wipe sled request
        // device is the block device, e.g. /dev/sda
        sledCmd := sled.CommandSet{
            &sled.Wipe {
                Device: "/dev/sda",
            },
            &sled.Write {},
            &sled.Kexec {},
        }

        // now we need to marshall it to write out
        jsonWipe, err := json.Marshal(sledCmd)
        if err != nil {
            log.Fatal(err)
        }

        // put in a key-value for our mac address
        // this is eth0 mac address, the value needs to be a sled.CommandSet
        err = bucket.Put([]byte("52:54:00:b4:c5:0d"), []byte(jsonWipe))

        // add a few other for shit and giggle
        err = bucket.Put([]byte("52:54:00:b1:64:a1"), []byte("42"))
        return nil
    })

    db.View(func(tx *bolt.Tx) error {
        // Assume bucket exists and has keys
        b := tx.Bucket([]byte("clients"))

        c := b.Cursor()

        for k, v := c.First(); k != nil; k, v = c.Next() {
            log.Printf("key=%s, value=%s\n", k, v)
        }

        return nil
    })
}

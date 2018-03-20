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
        // kill bucket (for testing)
        err = tx.DeleteBucket([]byte("clients"))
        // bucket is like a traditional database with name clients
        bucket, err := tx.CreateBucket([]byte("clients"))
        if err != nil {
            log.Fatal(err)
        }

        // create a bogus wipe sled request
        // device is the block device name (/sys/block) not (/dev), e.g. sda
        sledCmd := sled.CommandSet{
            &sled.Wipe {
                Device: "sda",
            },
            //&sled.Write {},
            nil,
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
        // FIXME: change everytime based on generated mac
        err = bucket.Put([]byte("52:54:00:ff:0c:b1"), []byte(jsonWipe))

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

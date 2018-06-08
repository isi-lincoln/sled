package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
)

var server string = "localhost"
var port string = "3000"

func main() {
	images := []string{"/tmp/kernel", "/tmp/initrd", "/tmp/image"}
	buf := make([]byte, 16)
	for _, v := range images {
		// create connection
		conn, _ := net.Dial("tcp", server+":"+port)
		log.Infof("connected to server")
		dev, err := os.OpenFile(v, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("write: error opening device %v", err)
		}
		// send message asking for iamge
		conn.Write([]byte(v))
		log.Infof("wrote %s to server", string(v))
		for {
			lenb, err := conn.Read(buf)
			log.Infof("reading: %s", string(buf[:lenb]))
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorf("Unable to read from server, err: %v", err)
			}
			// write image to disk
			dev.Write(buf[:lenb])
		}
		log.Infof("finished %s", string(v))
		dev.Close()
		conn.Close()
	}
}

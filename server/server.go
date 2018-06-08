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
	var amap map[string]string
	amap = make(map[string]string)
	amap["/tmp/kernel"] = "./kernel.txt"
	amap["/tmp/initrd"] = "./initrd.txt"
	amap["/tmp/image"] = "./image.txt"

	l, err := net.Listen("tcp", server+":"+port)
	if err != nil {
		log.Errorf("Error listening: %v", err)
	}
	defer l.Close()

	log.Infof("Listening on " + server + ":" + port)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Errorf("Error accepting: ", err.Error())
		}
		// thread the connection handling
		log.Infof("Received new communication")
		go handleRequest(amap, conn)
	}
}

func handleRequest(amap map[string]string, conn net.Conn) {
	buf := make([]byte, 16)
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Errorf("Unable to read from socket err: %v", err)
	}
	log.Infof("Received: %s", string(buf))
	// if key in our map, then send the contents
	if val, ok := amap[string(buf[:reqLen])]; ok {
		// open file to read contents from
		fs, err := os.Open(val)
		if err != nil {
			log.Errorf("Unable to open %s, err: %v", val, err)
		}
		for {
			// read the file and send out socket while not EOF
			lenb, err := fs.Read(buf)
			if err != nil {
				// when we hit an EOF, break execution
				if err == io.EOF {
					conn.Close()
					break
				}
				log.Errorf("Unable to read %s, err: %v", val, err)
			}
			log.Infof("writing: %s", string(buf[:lenb]))
			conn.Write(buf[:lenb])
		}
		log.Infof("Sent file.")
	}
}

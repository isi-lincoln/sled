package unit

import (
	"crypto/rand"
	sledc "github.com/ceftb/sled/sledc/pkg"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var DEVICE string = "test-device"
var SIZE int = 1024 * 1024

func TestSledcWipe(t *testing.T) {
	log.Infof("Testing: Sledc Wipe")

	// must be a /dev/test that we are wiping
	createTestDevice(DEVICE, SIZE)

	fistat, err := os.Stat(DEVICE)
	if err != nil {
		t.Errorf("device %s does not exist", DEVICE)
	}
	// get the size
	size := fistat.Size()

	// now we will test that wipe worked by reading all the bytes
	sledc.Wipe(DEVICE, size)

	buf := make([]byte, SIZE)
	// validate that what we wrote and what we read are correct
	readTestDevice(DEVICE, buf)
}

func TestSledcWrite(t *testing.T) {
	log.Infof("Testing: Sledc Write")
	// this should overwrite what we had there before - blank slate
	createTestDevice(DEVICE, SIZE)
	fistat, err := os.Stat(DEVICE)
	if err != nil {
		t.Errorf("device %s does not exist", DEVICE)
	} else {
	}
	// get the size
	size := fistat.Size()

	randBytes := readRand(size)
	log.Infof("Sledc Write: Writing Garbage to device")
	writeGarbage(DEVICE, randBytes)
	readTestDevice(DEVICE, randBytes)
	log.Infof("Sledc Write: contents of device and buffer are equal")

}

// helpers ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
func readRand(numBytes int64) []byte {
	randBytes := make([]byte, numBytes)
	_, err := rand.Read(randBytes)
	if err != nil {
		log.Fatalf("could not read from rand %v", err)
	}
	return randBytes
}

func writeGarbage(device string, buf []byte) {
	dev, err := os.OpenFile(device, os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}
	_, err = dev.Write(buf)
	if err != nil {
		log.Fatalf("write: error writing to device %v", err)
	}
}

// create a fake "device" - flat file
func createTestDevice(device string, size int) {
	dev, err := os.OpenFile(device, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Error creating test device %v", err)
	}
	buf := make([]byte, size)
	dev.Write(buf)
}

func readTestDevice(device string, testBuf []byte) {
	size := len(testBuf)
	buf := make([]byte, size)

	dev, err := os.OpenFile(device, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("Read: error opening device %v", err)
	}

	_, err = dev.Read(buf)
	if err != nil {
		log.Fatalf("Read: error reading device %v", err)
	}

	if len(buf) != len(testBuf) {
		log.Fatalf("device buffer does not equal write buffer size")
	}
	for i := range buf {
		if buf[i] != testBuf[i] {
			log.Fatalf("device buffer is not equal to write buffer")
		}
	}
}

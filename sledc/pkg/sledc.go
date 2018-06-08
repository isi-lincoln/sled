package sledc

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings" // remove newline from file
)

// Wipe the specified device clean with zeros.
// NOTE: now forcing "/dev" to be included in device name
func WipeBlock(device string) {
	log.Infof("wiping device %s", device)

	if !blockDeviceExists(device) {
		log.Fatalf("block device %s does not exist", device)
	}

	size := GetBlockDeviceSize(device)
	// TODO: add return code
	Wipe(device, size)

}

func Wipe(device string, size int64) {
	//wipe 1 kB at a time
	buf := make([]byte, 1024)

	dev, err := os.OpenFile(device,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}

	// explicitly set N to 0
	var N int64 = 0
	// golang while loop, set each 1k block to 0
	for N < size {
		n, err := dev.Write(buf)
		N += int64(n)
		if n < 1024 {
			break
		}
		if err != nil {
			log.Fatalf("error zeroing disk: %v", err)
		}
	}
	log.Infof("%d bytes zero'd in %s", size, device)

	if N < size {
		log.Warningf("only zeroed %d of %d bytes on disk", N, size)
	} else {
		log.Infof("device wiped")
	}
}

// Write the binary image to the specified device.
func Write(device string, image, kernel, initrd []byte) {
	// write the image to the block device
	WriteBlockImage(device, image)
	// write kernel, use flag "kernel" to name location
	WriteOther(kernel, "kernel")
	// write initramfs, use name "initrd"
	WriteOther(initrd, "initrd")
}

// kexec the image with the specified args
func Kexec(kernel, append, initrd string) {
	log.Infof("kexec - %s %s %s", kernel, append, initrd)

	// kexec -l -cmdline args -i initramfs kernel following u-root/golang-flag parsing
	out, err := exec.Command("kexec", "-l", "-cmdline", append,
		"-i", initrd, kernel).CombinedOutput()
	if err != nil {
		log.Fatalf("kexec load failed - %v : %s", err, out)
	}

	out, err = exec.Command("kexec", "-e").CombinedOutput()
	if err != nil {
		log.Fatalf("kexec execute failed - %v : %s", err, out)
	}
	log.Fatal("kexec did not exec ....")

}

// helpers ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func blockDeviceExists(device string) bool {
	devs := GetBlockDevices()
	for _, x := range devs {
		if device == x {
			return true
		}
	}
	return false
}

func GetBlockDevices() []string {

	file, err := os.Open("/sys/block")
	if err != nil {
		log.Fatalf("error opening /sys/block - %v", err)
	}
	defer file.Close()

	devs, err := file.Readdirnames(0)
	if err != nil {
		log.Fatalf("error reading /sys/block - %v", err)
	}

	return devs
}

func GetBlockDeviceSize(device string) int64 {
	content, err := ioutil.ReadFile(fmt.Sprintf("/sys/block/%s/size", device))
	if err != nil {
		log.Fatalf("error opening /sys/block/%s/size - %v", err)
	}
	// get rid of the nasty newline if it exists
	contentStr := strings.TrimSuffix(string(content), "\n")
	size, err := strconv.ParseInt(contentStr, 10, 0)
	if err != nil {
		log.Fatalf("error parsing /sys/block/%s/size = %v - %s", err, content)
	}
	// size is in disk sectors, multiply by 512 to get bytes
	size = size * 512
	return size
}

// writers ~~~~~~~~~~~~~~~~~~~~~~~~~

func WriteBlockImage(device string, image []byte) {
	if !blockDeviceExists(device) {
		log.Fatalf("block device %s does not exist", device)
	}
	// GetBlockDeviceSize is in bytes
	size := GetBlockDeviceSize(device)
	if int64(len(image)) > size {
		log.Fatalf("image is larger than target device - %d > %d", len(image), size)
	}

	WriteImage(device, image)
}

func WriteImage(device string, image []byte) {
	dev, err := os.OpenFile(device, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}
	defer dev.Close()

	n, err := dev.Write(image)
	if err != nil {
		log.Fatalf("write: error writing image - %v", err)
	}
	if n < len(image) {
		log.Fatalf("write: failed to write full image %d of %d bytes", n, len(image))
	} else {
		log.Infof("wrote %d bytes to %s", n, device)
	}
	log.Infof("writing image to device %s", device)
}

func WriteOther(kori []byte, flag string) {
	log.Infof("copying %s to /tmp/%s", flag, flag)
	// write kernel to tmp, shouldnt be need to have more than one kernel
	dev, err := os.OpenFile(fmt.Sprintf("/tmp/%s", flag),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}
	n, err := dev.Write(kori)
	if err != nil {
		log.Fatalf("write: error writing image - %v", err)
	}
	if n < len(kori) {
		log.Fatalf("write: failed to write %s %d of %d bytes", flag, n, len(kori))
	}
	dev.Close()
}

// images consist of the location on client to write them to
func WriteCommunicator(server string, images []string) error {
	buf := make([]byte, 4096)
	for _, v := range images {
		// create connection
		conn, _ := net.Dial("tcp", server+":3000")
		log.Infof("connected to server")
		dev, err := os.OpenFile(v, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("write: error opening device %v", err)
		}
		// send message asking for iamge
		conn.Write([]byte(v))
		log.Debugf("wrote %s to server", string(v))
		for {
			lenb, err := conn.Read(buf)
			//log.Debugf("reading: %s", string(buf[:lenb]))
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorf("Unable to read from server, err: %v", err)
			}
			// write image to disk
			dev.Write(buf[:lenb])
		}
		log.Debugf("finished %s", string(v))
		dev.Close()
		conn.Close()
	}
}

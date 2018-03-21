package main

import (
	"context"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io/ioutil"
	"math"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings" // remove newline from file

	"github.com/ceftb/sled"
)

var server = flag.String("server", "sled", "sled server to connect to")
var ifx = flag.String("interface", "eth0", "the interface to use for client id")

func main() {

	flag.Parse()

	conn, sledd := initClient()
	defer conn.Close()

	ifxs, err := net.Interfaces()
	if err != nil {
		log.Fatalf("error getting interface info - %v", err)
	}
	if len(ifxs) < 1 {
		log.Fatalf("no interfaces!")
	}
	mac := ""
	for _, x := range ifxs {
		if x.Name == *ifx {
			mac = x.HardwareAddr.String()
		}
	}
	if mac == "" {
		log.Fatalf("interface %s not found", *ifx)
	}

	resp, err := sledd.Command(context.TODO(), &sled.CommandRequest{mac})
	if err != nil {
		log.Fatalf("error getting sledd command - %v", err)
	}

	if resp.Wipe != nil {
		wipe(resp.Wipe.Device)
	}
	if resp.Write != nil {
		write(resp.Write.Device, resp.Write.Image, resp.Write.Kernel, resp.Write.Initrd)
	}
	if resp.Kexec != nil {
		kexec(resp.Kexec.Kernel, resp.Kexec.Append, resp.Kexec.Initrd)
	}
	if resp.Wipe == nil && resp.Write == nil && resp.Kexec == nil {
		log.Warn("received empty command from server")
	}
}

// Wipe the specified device clean with zeros.
func wipe(device string) {
	log.Infof("wiping device %s", device)

	if !blockDeviceExists(device) {
		log.Fatalf("block device %s does not exist", device)
	}

	//wipe 1 kB at a time
	buf := make([]byte, 1024)

	size := getBlockDeviceSize(device)

	dev, err := os.OpenFile(fmt.Sprintf("/dev/%s", device),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}

	// explicitly set N to 0
	var N int64 = 0
	// golang while loop
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

	if N < size {
		log.Warningf("only zeroed %d of %d bytes on disk", N, size)
	} else {
		log.Println("device wiped")
	}
}

// Write the binary image to the specified device.
func write(device string, image, kernel, initrd []byte) {
	log.Infof("copying kernel to memory")

	kdev, err := os.OpenFile(fmt.Sprintf("/tmp/kernel"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}
	n, err := kdev.Write(kernel)
	if err != nil {
		log.Fatalf("write: error writing image - %v", err)
	}
	if n < len(kernel) {
		log.Fatalf("write: failed to write kernel %d of %d bytes", n, len(kernel))
	}
	kdev.Close()

	log.Infof("copying initrd to memory")
	idev, err := os.OpenFile(fmt.Sprintf("/tmp/initrd"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}
	n, err = idev.Write(initrd)
	if err != nil {
		log.Fatalf("write: error writing image - %v", err)
	}
	if n < len(initrd) {
		log.Fatalf("write: failed to write kernel %d of %d bytes", n, len(initrd))
	}
	idev.Close()

	log.Infof("writing image to device %s", device)

	if !blockDeviceExists(device) {
		log.Fatalf("block device %s does not exist", device)
	}
	// getBlockDeviceSize is in bytes
	size := getBlockDeviceSize(device)
	if int64(len(image)) > size {
		log.Fatalf("image is larger than target device - %d > %d", len(image), size)
	}

	dev, err := os.OpenFile(fmt.Sprintf("/dev/%s", device),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}
	defer dev.Close()

	n, err = dev.Write(image)
	if err != nil {
		log.Fatalf("write: error writing image - %v", err)
	}
	if n < len(image) {
		log.Fatalf("write: failed to write full image %d of %d bytes", n, len(image))
	}
}

// kexec the image with the specified args
func kexec(kernel, append, initrd string) {
	log.Infof("kexec - %s %s %s", kernel, append, initrd)

	out, err := exec.Command("kexec", "-l", kernel, append, initrd).CombinedOutput()
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
	devs := getBlockDevices()
	for _, x := range devs {
		if device == x {
			return true
		}
	}
	return false
}

func getBlockDevices() []string {

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

func getBlockDeviceSize(device string) int64 {
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

// 8 GB max image size$
const maxMsgSize = math.MaxUint32

func initClient() (*grpc.ClientConn, sled.SledClient) {
	conn, err := grpc.Dial(
		*server+":6000",
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("could not connect to sled server - %v", err)
	}

	return conn, sled.NewSledClient(conn)
}

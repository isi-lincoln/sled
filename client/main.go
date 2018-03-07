package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/ceftb/sled"
)

func main() {
	fmt.Println("sled-client")

	conn, sledd := initClient()
	defer conn.Close()

	resp, err := sledd.Command(context.TODO(), &sled.CommandRequest{})
	if err != nil {
		log.Fatalf("error getting sledd command - %v", err)
	}

	if resp.Wipe != nil {
		wipe(resp.Wipe.Device)
	}
	if resp.Write != nil {
		write(resp.Write.Image, resp.Write.Device)
	}
	if resp.Kexec != nil {
		kexec(resp.Kexec.Kernel, resp.Kexec.Append, resp.Kexec.Initrd)
	}
}

// Wipe the specified device clean with zeros.
func wipe(device string) {
	if !blockDeviceExists(device) {
		log.Fatalf("block device %s does not exist", device)
	}
	size := getBlockDeviceSize(device)
	cmd := exec.Command(
		"dd",
		"if=/dev/null",
		fmt.Sprintf("of=/dev/%s", device),
		"bs=1",
		fmt.Sprintf("count=%d", size),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("wipe: could not execute dd - %v, %v", err, out)
	}
}

// Write the binary image to the specified device.
func write(image []byte, device string) {
	if !blockDeviceExists(device) {
		log.Fatalf("block device %s does not exist", device)
	}
	size := getBlockDeviceSize(device)
	if int64(len(image)) > size {
		log.Fatalf("image is larger than target device - %d > %d", len(image), size)
	}

	dev, err := os.Open(fmt.Sprintf("/dev/%s", device))
	if err != nil {
		log.Fatalf("write: error opening device %v", err)
	}
	defer dev.Close()

	n, err := dev.Write(image)
	if err != nil {
		log.Fatalf("write: error writing image - %v", err)
	}
	if n < len(image) {
		log.Fatalf("write: failed to write full image %d or %d bytes", n, len(image))
	}
}

// kexec the image with the specified args
func kexec(kernel, append, initrd string) {
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
	size, err := strconv.ParseInt(string(content), 10, 0)
	if err != nil {
		log.Fatalf("error parsing /sys/block/%s/size = %v - %s", err, content)
	}
	return size
}

func initClient() (*grpc.ClientConn, sled.SledClient) {
	conn, err := grpc.Dial("sled:6000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to sled server - %v", err)
	}

	return conn, sled.NewSledClient(conn)
}

package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {

	imageDir := "./images"
	baseURL := "https://mirror.deterlab.net/rvn/tests/"

	log.Infof("testing for image directory")
	_, err := os.Stat(imageDir)
	if err != nil {
		log.Infof("image directory not found, downloading..")
		err := os.Mkdir(imageDir, 0755)
		if err != nil {
			log.Fatalf("unable to create image directory")
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = os.Chdir(imageDir)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// hardcoded for now, parse unknown json later, or load rvn
	// client.kernel
	cKernel := "4.14.32-kernel"
	// client.initrd
	cInitrd := "4.14.32-initramfs"
	// client --- image ~~ not in model.js
	cImage := "ubuntu-1604"
	// now we need the kernel we will load with sledc
	cNewKernel := "vmlinuz-fedora-test"
	// now we need the initramfs we will load with sledc
	cNewInitrd := "initramfs-fedora-test"
	// server.image
	sImage := "fedora-27"
	// custom netboot image (smaller)
	netboot := "netboot"

	_, err = os.Stat(cKernel)
	if err != nil {
		wget(baseURL + cKernel)
	}
	_, err = os.Stat(cInitrd)
	if err != nil {
		wget(baseURL + cInitrd)
	}
	_, err = os.Stat(cImage)
	if err != nil {
		wget(baseURL + cImage)
	}
	_, err = os.Stat(cNewKernel)
	if err != nil {
		wget(baseURL + cNewKernel)
	}
	_, err = os.Stat(cNewInitrd)
	if err != nil {
		wget(baseURL + cNewInitrd)
	}

	_, err = os.Stat(sImage)
	if err != nil {
		wget(baseURL + sImage)
	}
	_, err = os.Stat(netboot)
	if err != nil {
		cmd := exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", netboot), "bs=1024", "count=4194304")
		log.Infof("Creating small netboot image.")
		err := cmd.Run()
		if err != nil {
			log.Warnf("Unable to create small netboot image: %v", err)
		}
	}
	err = os.Chdir(pwd)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func wget(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("unload to get url image: ", url)
	}
	defer resp.Body.Close()
	fileName := filepath.Base(url)
	output, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("unable to create file: ", fileName)
	}
	_, err = io.Copy(output, resp.Body)
	if err != nil {
		log.Fatalf("unable to copy file contents")
	}
	log.Infof("%s: Downloaded", fileName)
}

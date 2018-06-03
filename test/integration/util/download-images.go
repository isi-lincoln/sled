package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
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
	cKernel := "4.15.11-kernel"
	// client.initrd
	cInitrd := "initramfs"
	// client --- image ~~ not in model.js
	cImage := "ubuntu-1604"
	// server.image
	sImage := "fedora-27"
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
	_, err = os.Stat(sImage)
	if err != nil {
		wget(baseURL + sImage)
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

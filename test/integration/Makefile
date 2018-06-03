all: binary images

binary:
	cd bolt/ && go build manual-update.go

images:
	go run util/download-images.go

install:
	sudo cp images/netboot /var/rvn/img/
	sudo cp images/4.14.32-kernel /var/rvn/kernel
	sudo cp images/4.14.32-initramfs /var/rvn/initrd
	sudo cp images/fedora-27 /var/rvn/img

clean:
	rm -rf ./bolt/manual-update

spotless: clean
	rm -rf ./images/

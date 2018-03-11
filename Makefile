all: build/sledc build/sledd build/initramfs.cpio

build/sledc: sledc/main.go sled.pb.go | build
	go build -o $@ $<

build/sledd: sledd/main.go sled.pb.go | build
	go build -o $@ $<

sled.pb.go: sled.proto
	protoc -I . ./sled.proto --go_out=plugins=grpc:.

build/initramfs.cpio: $(GOPATH)/bin/u-root | build
	u-root -format=cpio -build=bb -o $@ \
		github.com/u-root/u-root/cmds/{ps,ls,ip,io,dhclient,wget,tcz,cat,pwd,builtin,boot,dd,dmesg,ed,find,grep,kexec,kill,modprobe,lsmod,mount,mv,ping,umount,uname,vboot,which,shutdown} \
		github.com/elves/elvish \
		github.com/ceftb/sled/sledc
	./update-cpio.sh

$(GOPATH)/bin/u-root:
	go install github.com/u-root/u-root

clean:
	rm -rf build

build:
	mkdir build

install: build/initramfs.cpio
	sudo cp build/initramfs.cpio /var/rvn/initrd/sled-0.1.0:x86_64

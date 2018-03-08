all: build/sledc build/sledd build/initramfs.cpio

build/sledc: sledc/main.go sled.pb.go | build
	go build -o $@ $<

build/sledd: sledd/main.go sled.pb.go | build
	go build -o $@ $<

sled.pb.go: sled.proto
	protoc -I . ./sled.proto --go_out=plugins=grpc:.

build/initramfs.cpio: $(GOPATH)/bin/u-root | build
	u-root -format=cpio -build=bb -o $@ \
		github.com/u-root/u-root/cmds/{ls,ip,dhclient,wget,tcz,cat} \
		github.com/ceftb/sled/sledc

$(GOPATH)/bin/u-root:
	go install github.com/u-root/u-root

clean:
	rm -rf build

build:
	mkdir build

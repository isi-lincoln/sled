
all: sledc sledd

sledc: client/main.go sled.pb.go
	go build -o $@ $<

sledd: server/main.go sled.pb.go
	go build -o $@ $<

sled.pb.go: sled.proto
	protoc -I . ./sled.proto --go_out=plugins=grpc:.


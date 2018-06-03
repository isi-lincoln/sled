all: binary images

binary:
	cd bolt/ && go build manual-update.go

images:
	go run download-images.go

clean:
	rm -rf ./bolt/manual-update

spotless:
	rm -rf ./images/

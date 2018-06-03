all: binary images

binary:
	cd bolt/ && go build manual-update.go

images:
	go run util/download-images.go

clean:
	rm -rf ./bolt/manual-update

spotless: clean
	rm -rf ./images/

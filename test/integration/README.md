## Integration testing

This suite is meant to test the end-to-end components for sled.

Using a very basic single client, server model.  The server and client
are configured using ssh and telnet respectively.  The server's bolt db
is updated with the specified image for the client.  The client begins
running sledc to grab the image from the server.  The image is specific
for testing, and must include the `rvn` user and ssh keys, as it relies
on pingwait for work.


### Testing

Run `make` to build the binary to update the server's bolt db entries.
Make will also download all the necessary images to use into the images/
directory.  The server image (default: fedora-27) will need to be manually
copied to `/var/rvn/img/`.  The client image by default is netboot. For
testing, I would recommend creating a blank 2GB image (via `dd`) to
reduce the time it takes to wipe the client drive.  That netboot image will
also live in `/var/rvn/img`.

From this `integration` directory, make sure to be root and run `go test`.
Root is necessary in order to run rvn.  `model.js` and
`bolt/manual-update.go` use the images/ directory for the kernel,
intramfs, and client image to load (ubuntu 1604).

The default timeout is 10 minutes, make sure that if a custom timeout is
used that it does not go below 5 minutes, as this is the timeout used by
pingwait, to enforce that the client drive was wiped, image downloaded,
and properly kexec'd.

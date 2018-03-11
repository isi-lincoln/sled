# sled - System Loader for Ephemeral Devices

sled is a system software loader designed to run in the [u-root](https://u-root.tk) initramfs. sled is designed to support loading systems onto devices that are ephemeral, for example shared devices that get reloaded with new systems for every user that uses them. The sled software consists of a client, server and API.

The sled client supports the following functionalities.

- disk wiping
- remote system image loading
- image image writing
- kexec-ing into images

Every time the client boots it asks the server what to do, so large numbers of devices can be manged from a single logical place.

## sledd

sledd issues commands to clients. Commands are issued to clients in groups. A group of commands can consist of any one of the following.

- wipe(device string)
- write(image, device string)
- kexec(kernel, append, initrd string)

sledd keeps track of clients by mac address. The internal bolt db has a single bucket called 'clients'. That bucket maps client mac addresses to command sets. sled also has a collection of images it can serve to clients. These images are kept in `/var/img`. Each `name` entry in the `write` is a reference to a filename located at `/var/img`.

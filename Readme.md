# sled - System Loader for Ephemeral Devices

sled is a system software loader designed to run in the [u-root](https://u-root.tk) initramfs. sled is designed to support loading systems onto devices that are ephemeral, for example shared devices that get reloaded with new systems for every user that uses them. The sled software consists of a client, server and API.

The sled client supports the following functionalities.

- disk wiping
- remote system image loading
- image image writing
- kexec-ing into images

Every time the client boots it asks the server what to do, so large numbers of devices can be manged from a single logical place.

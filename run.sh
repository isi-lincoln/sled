#!/bin/bash

qemu-system-x86_64 \
  -kernel ~/code/linux/vanilla/linux-stable/arch/x86/boot/bzImage \
  -initrd ~/code/ceftb/sled/build/initramfs.cpio \
  -append console=ttyS0 \
  -nographic \
  -netdev tap,id=net0,ifname=tap74,script=no \
  -device e1000,netdev=net0,id=net0,mac=52:54:00:a9:e1:27

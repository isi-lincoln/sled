#!/bin/sh

cd build

#xargs realpath --relative-base=/tmp/mods --no-symlinks |\
find -P /tmp/mods/lib -type f -or -type d |\
  xargs readlink -f | grep -oP "lib.*" |\
  cpio -H newc -o -D /tmp/mods --append -O initramfs.cpio


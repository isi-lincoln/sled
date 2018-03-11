#!/bin/sh

cd build

find -P /tmp/mods/lib -type f -or -type d |\
  xargs realpath --relative-to=/tmp/mods --no-symlinks |\
  cpio -H newc -o -D /tmp/mods --append -O initramfs.cpio


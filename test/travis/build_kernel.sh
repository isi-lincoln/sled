#!/usr/bin/env bash

# make sure we have pre-reqs to build kernel
sudo apt-get install -y libncurses5-dev gcc make git exuberant-ctags bc libssl-dev libelf-dev

LINUX_VERSION=4.14.32

# wget kernel, b/c it is quick
wget https://cdn.kernel.org/pub/linux/kernel/v4.x/linux-$LINUX_VERSION.tar.xz

tar xf linux-$LINUX_VERSION.tar.xz

# root directory
SLED=$(pwd | grep -Po '.*/sled')
# current directory
CWD=$(pwd)

# copy our config to linux stable
cp $SLED/kconfig/.config linux-$LINUX_VERSION/

cd linux-$LINUX_VERSION

# answer yes to all the new jazz
yes "" | make oldconfig

# build the kernel with our config
make -j `nproc` modules bzImage

# create the modules to build the initramfs, /tmp/mods default for scipts
INSTALL_MOD_PATH=/tmp/mods make modules_install

cd $CWD

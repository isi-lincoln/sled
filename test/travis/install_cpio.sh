#!/bin/bash
wget https://ftp.gnu.org/gnu/cpio/cpio-2.12.tar.gz
tar -xzf cpio-2.12.tar.gz
cd cpio-2.12
./configure
make
sudo make install

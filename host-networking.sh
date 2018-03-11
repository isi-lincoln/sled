#!/bin/bash

set -e

sudo ip link add br74 type bridge
sudo ip tuntap add dev tap74 mode tap user $(whoami)
sudo ip tuntap add dev tap77 mode tap user $(whoami)
sudo ip link set tap74 master br74
sudo ip link set tap77 master br74
sudo ip link set dev br74 up
sudo ip link set dev tap74 up

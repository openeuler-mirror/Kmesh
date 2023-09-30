#! /bin/sh

set -x

mkdir -p /opt/cni/bin
mkdir -p /etc/cni/net.d
\cp network.conf /etc/cni/net.d/

# Todo : git clone https://github.com/containernetworking/plugins and exec ./build_linux.sh to install ipam,loopback cni



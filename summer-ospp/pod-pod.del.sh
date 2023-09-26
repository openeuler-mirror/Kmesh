#! /bin/sh

set -x

del() {
    ip netns delete $1
}

del container1
del container2

ip link delete br0
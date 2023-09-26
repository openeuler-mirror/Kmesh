#! /bin/sh

set -x

ip route delete 10.1.2.0/24

container1="iperf1"
container2="iperf2"
container3="iperf3"



docker container rm -f $container1
docker container rm -f $container2
docker container rm -f $container3



ip netns delete container1
ip netns delete container2
ip netns delete container3
ip netns delete container5
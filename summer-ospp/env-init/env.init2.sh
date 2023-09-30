#! /bin/sh

set -x

ip route add 10.1.1.0/24 via 192.168.43.154 dev ens33 proto kernel

container1="iperf1"
container2="iperf2"

docker run -d --name nettool1  --network none my/nettool:v1
docker run -d --name nettool2  --network none my/nettool:v2


mkdir -p /var/run/netns/
pid1=$(docker inspect -f '{{.State.Pid}}' ${container1})
pid2=$(docker inspect -f '{{.State.Pid}}' ${container2})

ln -sfT /proc/$pid1/ns/net /var/run/netns/container1
ln -sfT /proc/$pid2/ns/net /var/run/netns/container2
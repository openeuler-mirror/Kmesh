# ! /bin/sh

set -x

ip route add 10.1.2.0/24 via 192.168.43.158 dev ens33 proto kernel


ip netns add container5


container1="iperf1"
container2="iperf2"
container3="iperf3"

docker run -d --name iperf1   --network none networkstatic/iperf3 -s
docker run -d --name iperf2   --network none networkstatic/iperf3 -s
docker run -d --name iperf3   --network none networkstatic/iperf3 -s


mkdir -p /var/run/netns/
pid1=$(docker inspect -f '{{.State.Pid}}' ${container1})
pid2=$(docker inspect -f '{{.State.Pid}}' ${container2})
pid3=$(docker inspect -f '{{.State.Pid}}' ${container3})

ln -sfT /proc/$pid1/ns/net /var/run/netns/container1
ln -sfT /proc/$pid2/ns/net /var/run/netns/container2
ln -sfT /proc/$pid3/ns/net /var/run/netns/container3
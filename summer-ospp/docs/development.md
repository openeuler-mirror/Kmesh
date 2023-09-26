
## run the program

0、sh pod-pod.sh
1、cd userspace
2、go generate
3、go run ../main.go
4、sudo mount -t bpf /sys/fs/bpf  /sys/fs/bpf
5、go build -o /opt/cni/bin/ebpf-based-cni main.go

## tc ops

[tc-bpf behavior](https://gist.github.com/anfredette/732eeb0fe519c8928d6d9c190728f7b5)

`tc qdisc add dev veth11 clsact`

`tc qdisc del dev veth11 clsact`

`tc filter add dev veth11 ingress bpf da obj ospp_bpfel.o sec tc`

`tc filter show dev veth11 ingress`

`tc filter del  dev veth11 ingress pref 49151`

`mount -t bpf /sys/fs/bpf  /sys/fs/bpf`

## arp ops

`ip netns exec container1 sh`
`ip neigh add 172.16.0.2 lladdr 0a:4a:51:57:45:d6 dev eth0`
`arp -d ipaddr`
`arp -n`

## cni development guide

https://github.com/containernetworking/plugins
https://github.com/containernetworking/cni/blob/main/SPEC.md
https://github.com/containerd/go-cni

mynet ipaddr manager location is `/var/lib/cni/networks/mynet`
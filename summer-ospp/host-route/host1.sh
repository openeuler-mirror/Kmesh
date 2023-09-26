#! /bin/sh
set -x
ip route add 10.1.1.0/24 via 10.1.1.1 dev ospp_host proto kernel src 10.1.1.1
ip route add 10.1.1.1 dev ospp_host proto kernel scope link
ip route add 10.1.2.0/24 via 192.168.43.158 dev ens33 proto kernel
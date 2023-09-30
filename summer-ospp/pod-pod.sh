#! /bin/sh

set -x

    BRIDGE_IP="10.1.0.1"
    IP1="10.1.0.2"
    IP2="10.1.0.3"
	

    CreateContainer(){

        # /*主机端 */
        ip netns add $1

        ip link add $2 type veth peer name $3

        ip link set dev $2 up

        ip link set $3 netns $1 name eth0

        # /* 容器端  */
        # // 给eth0分配ip地址
        if [ "$1" = "container1" ]; then
            ip netns exec $1 ip addr add $IP1/24 dev eth0
        else
            ip netns exec $1 ip addr add $IP2/24 dev eth0 
        fi   
        ip netns exec $1 ip link set lo up

        ip netns exec $1 ip link set eth0 up

        # // 创建默认路由 
        ip netns exec $1 ip route add default via $BRIDGE_IP dev eth0
    }

    CreateBridge(){
        # /* 非必要 ： 
        ip link add name br0 type bridge
        ip link set dev veth11 master br0
        ip link set dev veth21 master br0
        ip addr add $BRIDGE_IP/24 dev br0
        ip link set dev br0 up
        ip link set dev veth11 up
        ip link set dev veth21 up
        # */
    }


    CreateContainer container1 veth11 veth12 
    CreateContainer container2 veth21 veth22
    # CreateBridge
    
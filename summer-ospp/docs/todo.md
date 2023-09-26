## 9.2
1、pod1 --> node1  tc ingress 直接 TC_ACT_OK
2、node2 --> cilium_host
3、cilium_host --> pod  tc egress 根据目的ip查找ep，替换源mac和目的mac为pod的veth对
4、hardcode cilium_host cilium_net, 和pod内route info 和arp info
5、手动添加每个主机的路由
```
10.0.0.0/24 via 10.0.0.145 dev cilium_host proto kernel src 10.0.0.145 
10.0.0.145 dev cilium_host proto kernel scope link 
10.0.1.0/24 via 192.168.43.153 dev ens33 proto kernel 
```

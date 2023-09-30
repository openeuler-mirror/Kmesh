#!/bin/bash

# 循环生成10000个自定义链
for ((i=1; i<=10000; i++))
do
  # chain_name="MY_CUSTOM_CHAIN_$i"

  # 创建自定义链
  # iptables -t nat -N "$chain_name"

  # 引用自定义链到KUBE-SERVICES链
#   iptables -t nat -A KUBE-SERVICES -j "$chain_name"
  iptables -t nat -I PREROUTING 1 -p tcp -i eth_$i --dport 8061 -j ACCEPT
  
  # 在自定义链中添加规则（这里示例中添加一条DNAT规则）
  # iptables -t nat -A "$chain_name" -p tcp  --dport 8061 -i eth0 -j DNAT --to-destination 192.168.189.100:3456
done 
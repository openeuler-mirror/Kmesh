#!/bin/bash

# 循环删除10000个自定义链
for ((i=1; i<=1000; i++))
do
  chain_name="MY_CUSTOM_CHAIN_$i"

  # 删除自定义链
  iptables -t nat -D KUBE-SERVICES -p tcp  --dport 80 -j "$chain_name"
  iptables -t nat -F "$chain_name"
  iptables -t nat -X "$chain_name"
done

// SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause)
/* Copyright (c) 2022 Hengqi Chen */
#include "vmlinux.h"
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include "common.h"

#define IP_ADDRESS(x) (unsigned int)(10 + (1 << 8) + (64 << 16) + (x << 24))
#define TC_ACT_OK 0
#define TC_ACT_SHOT 2
#define ETH_P_IP 0x0800 /* Internet Protocol packet	*/
#define ETH_ALEN 6
#define NULL ((void *)0)
#define TCP_CSUM_OFF offsetof(struct tcphdr, check)
#define IS_PSEUDO 0x10
#define IP_CSUM_OFF offsetof(struct iphdr, check)
#define IP_DST_OFF offsetof(struct iphdr, daddr)
#define BPF_F_PSEUDO_HDR                (1ULL << 4)
#define ETH_HLEN 14
#define VIP_MOD  1

// The container IP address of the local node is stored
struct
{
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 255);
	__type(key, __u32);
	__type(value, struct endpoint_info);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
} local_pod_map SEC(".maps");

// lb_v4 svc ---> ip of ep
struct
{
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 255);
	__type(key, struct lb4_key);
	__type(value, __u32);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
} svc_map SEC(".maps");

struct
{
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 255);
	__type(key, struct simple_ct_key);
	__type(value, __u32);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
} simple_ct SEC(".maps");

SEC("tc")
int tc_ingress(struct __sk_buff *skb)
{
	void *data_end = (void *)(__u64)skb->data_end;
	void *data = (void *)(__u64)skb->data;
	struct ethhdr *l2;
	struct iphdr *l3;
	struct tcphdr *l4;
	int l3_off;
	int l4_off;
	struct endpoint_key key;
	struct endpoint_info *ep;
	u32 slot;
	struct lb4_key lb4key;
	u32 *backend_addr;
	__be32 sum;

	if (skb->protocol != bpf_htons(ETH_P_IP))
		return TC_ACT_OK;

	l2 = data;
	if ((void *)(l2 + 1) > data_end)
		return TC_ACT_OK;

	l3 = (struct iphdr *)(l2 + 1);
	if ((void *)(l3 + 1) > data_end)
		return TC_ACT_OK;
	l4 = (struct tcphdr *)((void *)l3 + ipv4_hdrlen(l3));
	if ((void *)(l4 + 1) > data_end)
		return TC_ACT_OK;
	if (l3->daddr == IP_ADDRESS(64))
	{
		// is vip addr
		__u32 old_ip;
		__u32 new_ip;
		bpf_printk("access via vip");
		slot = l4->source % VIP_MOD;
		bpf_printk("slot is %d",slot);
		old_ip = l3->daddr;
		lb4key.address = l3->daddr;
		lb4key.backend_slot = slot;
		backend_addr = bpf_map_lookup_elem(&svc_map,&lb4key);
		if (backend_addr == NULL)
			return TC_ACT_SHOT;
		new_ip = *backend_addr;
		struct simple_ct_key ct_key = {
			.dst_ip = l3->saddr,
			.src_ip = *backend_addr
		};
		__u32 v = 1;
		if (bpf_map_update_elem(&simple_ct,&ct_key,&v,BPF_ANY) < 0) {
			bpf_printk("update simple ct error");
		}
		int flags = BPF_F_PSEUDO_HDR;
		int ret;
		l3_off = ETH_HLEN;
		l4_off = l3_off + l3->ihl * 4;
		ret = bpf_skb_store_bytes(skb, l3_off + IP_DST_OFF, &new_ip, 4, 0);
		if (ret < 0) {
			bpf_printk("bpf_skb_store_bytes() failed: %d", ret);
		}
		__be32 sum;
		sum = bpf_csum_diff(&old_ip,4,&new_ip,4,0);
		ret = bpf_l3_csum_replace(skb, l3_off + IP_CSUM_OFF, 0, sum, 0);
		if (ret < 0) {
			bpf_printk("bpf_l3_csum_replace failed");
		}
		ret = bpf_l4_csum_replace(skb,l4_off + TCP_CSUM_OFF,0,sum,flags);
		if (ret < 0) {
			bpf_printk("bpf_l4_csum_replace failed");
		}
		bpf_printk("backend_addr is %x",new_ip);
		key.ip4 = *backend_addr;
	}
	else
	{
		// is ep addr
		bpf_printk("access via ep ip");
		key.ip4 = l3->daddr;
	}
	// key.ip4 = l3->daddr;
	ep = bpf_map_lookup_elem(&local_pod_map, &key);
	if (ep)
	{
		__u64 node_mac = ep->nodeMac;
		__u64 lxc_mac = ep->mac;
		bpf_skb_store_bytes(skb, 0, (__u8 *)&lxc_mac, ETH_ALEN, 0);
		bpf_skb_store_bytes(skb, ETH_ALEN, (__u8 *)&node_mac, ETH_ALEN, 0);
		return bpf_redirect(ep->ifindex, 0);
	}
	else
	{
		bpf_printk("ep not found in local pod map at host facing veth tc bpf");
		data = (void *)(__u64)skb->data;
		data_end = (void *)(__u64)skb->data_end;
		l2 = data;
		if ((void *)(l2 + 1) > data_end)
			return TC_ACT_OK;
		return bpf_redirect_neigh(2, NULL, 0, 0);
	}
	return TC_ACT_OK;
}

SEC("tc")
int tc_egress(struct __sk_buff *skb) {

	bpf_printk("enter veth egress");
	void *data_end = (void *)(__u64)skb->data_end;
	void *data = (void *)(__u64)skb->data;
	struct ethhdr *l2;
	struct iphdr *l3;
	l2 = data;
	if ((void *)(l2 + 1) > data_end)
		return TC_ACT_OK;

	l3 = (struct iphdr *)(l2 + 1);
	if ((void *)(l3 + 1) > data_end)
		return TC_ACT_OK;
	struct simple_ct_key ct_key = {
		.src_ip = l3->saddr,
		.dst_ip = l3->daddr
	};
	__u32 *v;
	v = bpf_map_lookup_elem(&simple_ct,&ct_key);
	if (v!=NULL){
		bpf_printk("found a ct entry");
	}
	if (v != NULL) {
		//Todo: reply skb to process tcp csum
		l3->saddr = IP_ADDRESS(64);
		l3->check = iph_csum(l3);
		bpf_printk("snat");
	}
	bpf_printk("veth egress 2 point saddr is %x",l3->saddr);
	return TC_ACT_OK;
}

SEC("tc_host")
int tc_ingress_host(struct __sk_buff *skb)
{

	void *data_end = (void *)(__u64)skb->data_end;
	void *data = (void *)(__u64)skb->data;
	struct ethhdr *l2;
	struct iphdr *l3;
	struct endpoint_key key;
	struct endpoint_info *ep;

	if (skb->protocol != bpf_htons(ETH_P_IP))
		return TC_ACT_OK;

	l2 = data;
	if ((void *)(l2 + 1) > data_end)
		return TC_ACT_OK;

	l3 = (struct iphdr *)(l2 + 1);
	if ((void *)(l3 + 1) > data_end)
		return TC_ACT_OK;
	key.ip4 = l3->daddr;
	if (l3->daddr == 0x9e2ba8c0 || l3->daddr == 0x9a2ba8c0)
		return TC_ACT_OK;
	ep = bpf_map_lookup_elem(&local_pod_map, &key);
	if (ep)
	{
		/* code */
		bpf_printk("ep found in local pod map at host tc bpf ");
		__u64 node_mac = ep->nodeMac;
		__u64 lxc_mac = ep->mac;
		bpf_skb_store_bytes(skb, 0, (__u8 *)&lxc_mac, ETH_ALEN, 0);
		bpf_skb_store_bytes(skb, ETH_ALEN, (__u8 *)&node_mac, ETH_ALEN, 0);
		return bpf_redirect(ep->ifindex, 1);
	}
	else 
	{
		bpf_printk("ep not found in local pod map at host tc bpf");
	}
	return TC_ACT_OK;
}

char __license[] SEC("license") = "GPL";

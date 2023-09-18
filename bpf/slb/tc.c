/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
 * MeshAccelerating is licensed under the Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *	 http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Author: Bitcoffee
 * Create: 2023-05-22
 */
#include <linux/pkt_cls.h>
#include "slb_common.h"
#include "map.h"
#include "bpf_log.h"
#include "header_info.h"

#define TCP_CSUM_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, check))
#define TCP_SPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, source))

#define IP_CSUM_OFF (ETH_HLEN + offsetof(struct iphdr, check))
#define IP_SRC_OFF (ETH_HLEN + offsetof(struct iphdr, saddr))

SEC("tc")
int tc_xdp_rev_nat(struct __sk_buff *skb)
{
	
	int ret;
	struct header_info_t header_info = {0};
	tuple_t tuple = {0};
	tuple_t tuple_rev = {0};

	ret = parser_header_info(skb, &header_info);
	if (ret != PROCESS) {
		BPF_LOG(INFO, KMESH, "tc skb parse failed");
		return ret;
	}
	
	parse_v4_tuple(header_info.iph, header_info.tcph, &tuple);

	//parse_v4_tuple(header_info->iph, header_info->tcph, &tuple);
	// flag =1 :snat records
	// tuple.flags = TUPLE_FLAGS_EGRESS;

	/* dnat
	 * key:     src_ipv4	1.2.3.4		src_port	1234	client
	 *          dst_ipv4	7.6.122.2	dst_port	30080	nodport listen
	 * value:   nat_saddr 	1.2.3.4 	nat_sport	1234	client
	 *          nat_daddr	192.168.0.2	nat_dport	8080	endpoint
	 *          nat_smac	A1:A2:A3:A4:A5:A6				client mac
	 *          nat_dmac	01:02:03:04:05:06				endpoint mac
	 *          ifindex		7								egress adapt
	 * rev-dnat
	 * key:     src_ipv4	192.168.0.2	src_port	8080	endpoint
	 *          dst_ipv4	1.2.3.4 	dst_port	1234	client
	 * value:   nat_saddr	7.6.122.2	nat_sport	30080	nodeport listen
	 *          nat_daddr	1.2.3.4		nat_dport	1234	client
	 *          nat_smac	01:02:03:04:05:06				nodport mac
	 *          nat_dmac	A1:A2:A3:A4:A5:A6				client macok
	 *          index		7								ingress adapt
	 */
	
	struct ct_value_t *rev_ct_value = map_lookup_ct(&tuple);
	if (!rev_ct_value)
		return TC_ACT_OK;

	BPF_LOG(DEBUG, KMESH, "tc_xdp_rev_nat:origin src %x:%x, protocol=%x",
			tuple.src_ipv4, tuple.src_port, tuple.protocol);
	BPF_LOG(DEBUG, KMESH, "tc_xdp_rev_nat:dst %x:%x", tuple.dst_ipv4, tuple.dst_port);

	__u32 new_src_ip = rev_ct_value->nat_info.nat_saddr.v4_saddr;
	__u16 new_src_port = (__be16)rev_ct_value->nat_info.nat_sport;

	if (is_fin_package(&header_info))
		rev_ct_value->fin_seq = bpf_ntohl(header_info.tcph->seq);
	else if (rev_ct_value->fin_seq && bpf_ntohl(header_info.tcph->seq) > rev_ct_value->fin_seq) {
		tuple_rev = tuple;
		tuple_rev.src_ipv4 = rev_ct_value->nat_info.nat_daddr.v4_daddr;
		tuple_rev.src_port = rev_ct_value->nat_info.nat_dport;
		tuple_rev.dst_ipv4 = rev_ct_value->nat_info.nat_saddr.v4_saddr;
		tuple_rev.dst_port = rev_ct_value->nat_info.nat_sport;
		map_delete_ct(&tuple);
		map_delete_ct(&tuple_rev);
	}

	__u32 old_src_ip = header_info.iph->saddr;
	__u16 old_src_port = (__be16)header_info.tcph->source;

	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_src_ip, new_src_ip, BPF_F_PSEUDO_HDR | sizeof(new_src_ip));
	bpf_l3_csum_replace(skb, IP_CSUM_OFF, old_src_ip, new_src_ip, sizeof(new_src_ip));
	bpf_skb_store_bytes(skb, IP_SRC_OFF, &new_src_ip, sizeof(new_src_ip), 0);

	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_src_port, new_src_port, sizeof(new_src_port));
	bpf_skb_store_bytes(skb, TCP_SPORT_OFF, &new_src_port, sizeof(new_src_port), 0);
	BPF_LOG(DEBUG, KMESH, "tc_xdp_rev_nat: new src info: %x:%x.\n", new_src_ip, new_src_port);
	return TC_ACT_OK;
}


char _license[] SEC("license") = "GPL";

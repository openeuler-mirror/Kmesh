/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
 * MeshAccelerating is licensed under the Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Author: bitcoffee
 * Create: 2023-05-12
 */
#ifndef _HEADER_INFO_H_
#define _HEADER_INFO_H_

#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/tcp.h>
#include <linux/pkt_cls.h>
#include <linux/bpf.h>
#include <stdbool.h>
#include "csum.h"

#ifdef XDP_CTX
#define PASS XDP_PASS	// 2
#define DROP XDP_DROP   // 1
#define PROCESS -1
#elif TC_CTX
#define PASS TC_ACT_OK  // 0
#define DROP TC_ACT_SHOT  // 2
#define PROCESS -1
#endif

struct header_info_t {
	struct ethhdr  *ethh;
	struct iphdr   *iph;
	struct ipv6hdr *ip6h;
	struct tcphdr  *tcph;
	void *data;
	void *data_end;
};

#define CHECK_ETH_NULL_OR_OVER_BOUND(p, bound) \
({ \
	bool ret = false; \
	if ((!p) || \
		((void *)(p) + sizeof(struct ethhdr) > (bound))) \
		ret = true; \
	ret; \
})

#define CHECK_IPV4_NULL_OR_OVER_BOUND(p, bound) \
({ \
	bool ret = false; \
	if ((!p) || \
	((void *)(p) + sizeof(struct iphdr) > (bound)) || \
		((void *)(p) + (size_t)((p)->ihl * 4) > (bound))) \
		ret = true; \
	ret; \
})

#define CHECK_TCP_NULL_OR_OVER_BOUND(p, bound) \
({ \
	bool ret = false; \
	if ((!p) || \
		((void *)(p) + sizeof(struct tcphdr) > (bound)) || \
		((void *)(p) + (size_t)((p)->doff * 4) > (bound))) \
		ret = true; \
	ret; \
})

static inline int parser_tcp_info(struct header_info_t *header, size_t l3_off)
{
	int ret = PROCESS;
	size_t l4_off;

	l4_off = l3_off + sizeof(struct iphdr);
	header->tcph = (struct tcphdr *)(header->data + l4_off);
	if (CHECK_TCP_NULL_OR_OVER_BOUND(header->tcph, header->data_end)) {
		BPF_LOG(INFO, KMESH, "get a invalid package\n");
		ret = DROP;
	}
	return ret;
}

static inline bool is_syn_package(struct header_info_t *header_info)
{
	if (header_info->tcph->syn && !header_info->tcph->ack)
		return true;
	return false;
}

static inline bool is_fin_package(struct header_info_t *header_info)
{
	if (header_info->tcph->fin)
		return true;
	return false;
}

/*
 * need header->data/data_end not null and header->ethh pass range check
 */
static inline int parser_ipv4_info(struct header_info_t *header)
{
	int ret;
	size_t l3_off = sizeof(struct ethhdr);

	header->iph = (struct iphdr *)(header->data + l3_off);
	if (CHECK_IPV4_NULL_OR_OVER_BOUND(header->iph, header->data_end)) {
		BPF_LOG(INFO, KMESH, "parser header info load get a invalid package, incomplete ipv4 header\n");
		return DROP;
	}

	switch(header->iph->protocol) {
		case IPPROTO_TCP:
			// FIXME: current 5.10 kernel verifier not support l4_off calc by ihl * 4
			// l4_off = l3_off + (size_t)(header->iph->ihl * 4);
			ret = parser_tcp_info(header, l3_off);
			break;
		default:
			BPF_LOG(INFO, KMESH, "don't support %u\n", header->iph->protocol);
			ret = PASS;
	}
	return ret;
}

#ifdef XDP_CTX
static inline int parser_header_info(struct xdp_md *ctx, struct header_info_t *header)
#elif TC_CTX
static inline int parser_header_info(struct __sk_buff *ctx, struct header_info_t *header)
#else 
#error "parser current use in xdp or tc program, need -DXDP_CTX or -DTC_CTX"
#endif
{
	int ret;
	header->data = (void*)(unsigned long)ctx->data;
	header->data_end = (void*)(unsigned long)ctx->data_end;
	header->ethh = (struct ethhdr *)header->data;
	
	if (CHECK_ETH_NULL_OR_OVER_BOUND(header->ethh, header->data_end)) {
		BPF_LOG(INFO, KMESH, "parser header get a invalid package, eth incomplete!\n");
		return DROP;
	}
	
	switch (header->ethh->h_proto) {
		case bpf_htons(ETH_P_IP):
			ret = parser_ipv4_info(header);
			break;
		// todo v6
		default:
			BPF_LOG(INFO, KMESH, "don't support %u\n", header->ethh->h_proto);
			ret = PASS;
	}
	return ret;
}

static inline void dnat_header_v4(struct header_info_t *header, __u32 new_daddr, __u32 new_dport)
{
	header->iph->daddr = new_daddr;
	header->tcph->dest = new_dport;
}

static inline void snat_header_v4(struct header_info_t *header, __u32 new_saddr, __u32 new_sport)
{
	header->iph->saddr = new_saddr;
	header->tcph->source = new_sport;
}

static inline void replace_header_v4_l3(struct header_info_t *header, __wsum diff)
{
	csum_replace_by_diff(&header->iph->check, diff);
}

static inline void replace_header_v4_l4(struct header_info_t *header, __wsum diff_ip, __wsum diff_port)
{
	csum_replace_by_diff(&header->tcph->check, diff_ip);
	csum_replace_by_diff(&header->tcph->check, diff_port);
}
#endif /* _HEADER_INFO_H_ */
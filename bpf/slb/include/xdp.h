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
#ifndef _XDP_H_
#define _XDP_H_

#include "map.h"
#include "bpf_log.h"
#include "tuple.h"
#include "header_info.h"
#include "conntrack.h"
#include "csum.h"
#include <bpf/bpf_endian.h>
#include <linux/bpf.h>

static inline void xdp_eth_nat(__u8 origin_smac[], __u8 origin_dmac[], __u8 smac[], __u8 dmac[])
{
	const int MAC_LENGTH = 6;
	bpf_memcpy(origin_smac, smac, MAC_LENGTH);
	bpf_memcpy(origin_dmac, dmac, MAC_LENGTH);
}

static inline void add_dnat_ct(struct header_info_t *header_info, struct ct_value_t *ct_value,
	tuple_t *tuple, tuple_t *tuple_rev)
{
	int ret;
	// add ct
	ct_value->nat_info.protocol = header_info->iph->protocol;
	ct_value->nat_info.nat_saddr.v4_saddr = header_info->iph->saddr;
	ct_value->nat_info.nat_daddr.v4_daddr = header_info->iph->daddr;
	ct_value->nat_info.nat_sport = header_info->tcph->source;
	ct_value->nat_info.nat_dport = header_info->tcph->dest;
	ct_value->nat_info.nat_type = NAT_TYPE_DNAT;
	xdp_eth_nat(ct_value->nat_info.nat_mac_info.nat_smac, ct_value->nat_info.nat_mac_info.nat_dmac,
		header_info->ethh->h_source, header_info->ethh->h_dest);
	ret = map_update_ct(tuple, ct_value);
	if (ret) {
		BPF_LOG(ERR, KMESH, "add connect track failed!, ret:%d\n", ret);
		return;
	}
	// add rev_ct
	tuple_rev->protocol = tuple->protocol;
	tuple_rev->src_ipv4 = header_info->iph->daddr;
	tuple_rev->dst_ipv4 = header_info->iph->saddr;
	tuple_rev->src_port = header_info->tcph->dest;
	tuple_rev->dst_port = header_info->tcph->source;
	ct_value->nat_info.nat_saddr.v4_saddr = tuple->dst_ipv4;
	ct_value->nat_info.nat_daddr.v4_daddr = header_info->iph->saddr;
	ct_value->nat_info.nat_sport = tuple->dst_port;
	ct_value->nat_info.nat_dport = header_info->tcph->source;
	ct_value->nat_info.nat_type = NAT_TYPE_DNAT;
	xdp_eth_nat(ct_value->nat_info.nat_mac_info.nat_smac, ct_value->nat_info.nat_mac_info.nat_dmac,
		header_info->ethh->h_dest, header_info->ethh->h_source);
	ret = map_update_ct(tuple_rev, ct_value);
	if (ret) {
		BPF_LOG(ERR, KMESH, "add connect track failed!, ret:%d\n", ret);
		(void)map_delete_ct(tuple);
	}
}

static inline void add_fullnat_ct(struct header_info_t *header_info, struct ct_value_t *ct_value,
	tuple_t *tuple, tuple_t *tuple_rev, struct bpf_fib_lookup *fib_params)
{
	// add rev ct
	int ret;
	tuple_rev->protocol = tuple->protocol;
	tuple_rev->src_ipv4 = header_info->iph->daddr;
	tuple_rev->dst_ipv4 = header_info->iph->saddr;
	tuple_rev->src_port = header_info->tcph->dest;
	tuple_rev->dst_port = header_info->tcph->source;

	ct_value->nat_info.protocol = tuple->protocol;
	ct_value->nat_info.nat_saddr.v4_saddr = tuple->dst_ipv4;
	ct_value->nat_info.nat_daddr.v4_daddr = tuple->src_ipv4;
	ct_value->nat_info.nat_sport = tuple->dst_port;
	ct_value->nat_info.nat_dport = tuple->src_port;
	ct_value->nat_info.nat_type = NAT_TYPE_FULL_NAT;
	// rev ct mac need rev
	xdp_eth_nat(ct_value->nat_info.nat_mac_info.nat_smac, ct_value->nat_info.nat_mac_info.nat_dmac,
		header_info->ethh->h_dest, header_info->ethh->h_source);
	
	ret = map_update_ct(tuple_rev, ct_value);
	if (ret) 
		BPF_LOG(ERR, KMESH, "add connect track failed!, ret:%d\n", ret);

	bpf_memcpy(header_info->ethh->h_source, fib_params->smac, sizeof(fib_params->smac));
	bpf_memcpy(header_info->ethh->h_dest, fib_params->dmac, sizeof(fib_params->dmac));

	// add ct
	ct_value->nat_info.protocol = header_info->iph->protocol;
	ct_value->nat_info.nat_saddr.v4_saddr = header_info->iph->saddr;
	ct_value->nat_info.nat_daddr.v4_daddr = header_info->iph->daddr;
	ct_value->nat_info.nat_sport = header_info->tcph->source;
	ct_value->nat_info.nat_dport = header_info->tcph->dest;
	ct_value->nat_info.nat_type = NAT_TYPE_FULL_NAT;
	xdp_eth_nat(ct_value->nat_info.nat_mac_info.nat_smac, ct_value->nat_info.nat_mac_info.nat_dmac,
		fib_params->smac, fib_params->dmac);
	ct_value->nat_info.nat_mac_info.nat_ifindex = fib_params->ifindex;
	
	ret = map_update_ct(tuple, ct_value);
	if (ret) {
		BPF_LOG(ERR, KMESH, "add connect track failed!, ret:%d\n", ret);
		(void)map_delete_ct(tuple);
	}
}

static inline __u32 get_local_ipv4(tuple_t *tuple)
{
	return tuple->dst_ipv4;
}

static inline bool is_port_used(tuple_t *tuple)
{
	if (map_lookup_usedport(tuple))
		return true;
	return false;
}

static inline __u32 get_local_port(tuple_t *tuple)
{
	__u32 current_time;
	__le32 usable_port = bpf_get_prandom_u32() % 60000 + 2000;
#pragma unroll
	for (int i = 0; i < 32; i++, usable_port++) {
		tuple->dst_port = bpf_htons(usable_port);
		if (is_port_used(tuple)) {
			continue;
		}
		__u32 current_time = bpf_ktime_get_ns() / 1000000000;
		if (!map_update_usedport(tuple, &current_time))
			goto success;
	}
	return 0;
success:
	return bpf_htons(usable_port);
}

static inline bool is_local()
{
	return true;
}

static inline int xdp_process_nat(struct xdp_md *xdp_ctx,
		struct header_info_t *header_info,
		tuple_t *tuple,
		struct endpoint_entry_t *endpoint)
{
	
	int ret;
	__u32 local_ipv4, local_port;

	struct ct_value_t ct_value = {0};
	struct bpf_fib_lookup fib_params = {0};
	tuple_t tuple_rev = {0};

	__u32 sum_diff_l3;
	__u32 sum_diff_l4;
	__u32 origin_dport = header_info->tcph->dest;
	__u32 origin_sport = header_info->tcph->source;

	sum_diff_l3 = bpf_csum_diff(&header_info->iph->daddr, sizeof(header_info->iph->daddr),
		&endpoint->ipv4, sizeof(endpoint->ipv4), 0);
	sum_diff_l4 = bpf_csum_diff(&origin_dport, sizeof(origin_dport),
		&endpoint->port, sizeof(endpoint->port), 0);
	replace_header_v4_l3(header_info, sum_diff_l3);
	replace_header_v4_l4(header_info, sum_diff_l3, sum_diff_l4);
	dnat_header_v4(header_info, endpoint->ipv4, endpoint->port);

	ct_value.nat_info.nat_mac_info.nat_ifindex = xdp_ctx->ingress_ifindex;

	if (is_local()) {
		add_dnat_ct(header_info, &ct_value, tuple, &tuple_rev);
		return XDP_PASS;
	}

	local_ipv4 = get_local_ipv4(tuple);

	tuple_rev.protocol = tuple->protocol;
	tuple_rev.src_ipv4 = endpoint->ipv4;
	tuple_rev.src_port = endpoint->port;
	tuple_rev.dst_ipv4 = tuple->dst_ipv4;
	local_port = get_local_port(&tuple_rev);

	sum_diff_l3 = bpf_csum_diff(&header_info->iph->saddr, sizeof(header_info->iph->saddr),
		&local_ipv4, sizeof(local_ipv4), 0);
	sum_diff_l4 = bpf_csum_diff(&origin_sport, sizeof(origin_sport),
		&local_port, sizeof(local_port), 0);
	replace_header_v4_l3(header_info, sum_diff_l3);
	replace_header_v4_l4(header_info, sum_diff_l3, sum_diff_l4);
	snat_header_v4(header_info, local_ipv4, local_port);

	fib_params.family = AF_INET;
	fib_params.tos = header_info->iph->tos;
	fib_params.l4_protocol = header_info->iph->protocol;
	fib_params.tot_len = bpf_ntohs(header_info->iph->tot_len);
	fib_params.ifindex = xdp_ctx->ingress_ifindex;
	fib_params.ipv4_dst = endpoint->ipv4;
	fib_params.ipv4_src = local_ipv4;	

	ret = bpf_fib_lookup(xdp_ctx, &fib_params, sizeof(fib_params), 0);
	if (ret) {
		BPF_LOG(ERR, KMESH, "xdp lookup fib failed! ret:%d\n", ret);
		return XDP_ABORTED;
	}

	add_fullnat_ct(header_info, &ct_value, tuple, &tuple_rev, &fib_params);

	if (fib_params.ifindex == xdp_ctx->ingress_ifindex)
		return XDP_TX;

	ret = bpf_redirect(fib_params.ifindex, 0);
	BPF_LOG(DEBUG, KMESH, "bpf redirect ret %u\n", ret);
	return ret;    

}

#endif /* _XDP_H_ */

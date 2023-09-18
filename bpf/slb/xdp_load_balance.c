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

#include "slb_common.h"
#include "l4_manager.h"
#include "xdp.h"
#include <bpf/bpf_endian.h>

static inline int xdp_redirect_endpoint(struct xdp_md *xdp_ctx,
	struct header_info_t *header_info, tuple_t *tuple)
{
	int ret;
	// ipv4
	ret = l4_manager(xdp_ctx, header_info, tuple);
	// todo ipv6
	if (ret < 0) {
		if (ret == -ENOENT)
			ret = XDP_PASS;
		else
			ret = XDP_ABORTED;;
	}
	return ret;
}

static inline int process_ipv4_package(struct xdp_md *xdp_ctx, struct header_info_t *header_info,
	tuple_t *tuple, struct ct_value_t *ct)
{
	__u32 sum_diff_l3, sum_diff_l4;
	__u32 origin_sport = header_info->tcph->source;
	__u32 origin_dport = header_info->tcph->dest;

	sum_diff_l3 = bpf_csum_diff(&header_info->iph->daddr, sizeof(header_info->iph->daddr),
		&ct->nat_info.nat_daddr.v4_daddr, sizeof(ct->nat_info.nat_daddr.v4_daddr), 0);
	sum_diff_l4 = bpf_csum_diff(&origin_dport, sizeof(origin_dport),
		&ct->nat_info.nat_dport, sizeof(ct->nat_info.nat_dport), 0);
	replace_header_v4_l3(header_info, sum_diff_l3);
	replace_header_v4_l4(header_info, sum_diff_l3, sum_diff_l4);
	dnat_header_v4(header_info, ct->nat_info.nat_daddr.v4_daddr, ct->nat_info.nat_dport);

	if (ct->nat_info.nat_type == NAT_TYPE_DNAT)
		return XDP_PASS;
		
	sum_diff_l3 = bpf_csum_diff(&header_info->iph->saddr, sizeof(header_info->iph->saddr),
		&ct->nat_info.nat_saddr.v4_saddr, sizeof(ct->nat_info.nat_saddr.v4_saddr), 0);
	sum_diff_l4 = bpf_csum_diff(&origin_sport, sizeof(origin_sport),
		&ct->nat_info.nat_sport, sizeof(ct->nat_info.nat_sport), 0);
	replace_header_v4_l3(header_info, sum_diff_l3);
	replace_header_v4_l4(header_info, sum_diff_l3, sum_diff_l4);
	snat_header_v4(header_info, ct->nat_info.nat_saddr.v4_saddr, ct->nat_info.nat_sport);

	xdp_eth_nat(header_info->ethh->h_source, header_info->ethh->h_dest,
		ct->nat_info.nat_mac_info.nat_smac,
		ct->nat_info.nat_mac_info.nat_dmac);

	if (ct->nat_info.nat_mac_info.nat_ifindex == xdp_ctx->ingress_ifindex)
		return XDP_TX;
	
	return bpf_redirect(ct->nat_info.nat_mac_info.nat_ifindex, 0);
}

static inline int xdp_redirect(struct xdp_md *xdp_ctx, struct header_info_t *header_info)
{
	struct ct_value_t *rev_value;
	tuple_t tuple = {0};
	tuple_t tuple_rev = {0};
	int ret;

	parse_v4_tuple(header_info->iph, header_info->tcph, &tuple);

	if (is_syn_package(header_info))
		return xdp_redirect_endpoint(xdp_ctx, header_info, &tuple);

	rev_value = map_lookup_ct(&tuple);
	if (!rev_value)
		return XDP_PASS;
	/* fullnat
	 * key:     src_ipv4	1.2.3.4		src_port	1234	client
	 *          dst_ipv4	7.6.122.2	dst_port	30080	nodeport listen
	 * value:   nat_saddr 	7.6.122.2	nat_sport	55555	nodeport
	 *          nat_daddr	192.168.0.2	nat_dport	8080	endpoint
	 *          nat_smac	01:02:03:04:05:06				nodeport mac
	 *          nat_dmac	11:12:13:14:15:16				endpoint mac
	 *          ifindex		7								egress adapt
	 * rev-fullnat
	 * key:     src_ipv4	192.168.0.2	src_port	8080	endpoint
	 *          dst_ipv4	7.6.122.2	dst_port	55555	nodeport
	 * value:   nat_saddr	7.6.122.2	nat_sport	30080	nodeport listen
	 *          nat_daddr	1.2.3.4		nat_dport	1234	client
	 *          nat_smac	01:02:03:04:05:06				nodport mac
	 *          nat_dmac	A1:A2:A3:A4:A5:A6				client mac
	 *          index		7								ingress adapt
	 */
	// ipv4
	BPF_LOG(DEBUG, KMESH, "found a ct record");
	ret = process_ipv4_package(xdp_ctx, header_info, &tuple, rev_value);
	if (is_fin_package(header_info))
		rev_value->fin_seq = bpf_ntohl(header_info->tcph->seq);
	else if (rev_value->fin_seq && bpf_ntohl(header_info->tcph->seq) > rev_value->fin_seq) {
		tuple_rev = tuple;
		tuple_rev.src_ipv4 = rev_value->nat_info.nat_daddr.v4_daddr;
		tuple_rev.src_port = rev_value->nat_info.nat_dport;
		tuple_rev.dst_ipv4 = rev_value->nat_info.nat_saddr.v4_saddr;
		tuple_rev.dst_port = rev_value->nat_info.nat_sport;
		map_delete_ct(&tuple);
		map_delete_usedport(&tuple);
		map_delete_ct(&tuple_rev);
		map_delete_usedport(&tuple_rev);
	}
	return ret;
	// todo: ipv6
}

/* Balance xdp bpf prog */
SEC("xdp_balance")
int xdp_load_balance(struct xdp_md *ctx)
{
	struct header_info_t header_info = {0};
	// TODO ipv6
	int ret = parser_header_info(ctx, &header_info);
	if (ret != PROCESS) {
		BPF_LOG(INFO, KMESH, "xdp parse failed");
		return ret;
	}

	return xdp_redirect(ctx, &header_info);	
}

char _license[] SEC("license") = "GPL";
int _version SEC("version") = 1;
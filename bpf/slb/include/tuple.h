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
 * Author: bitcoffee
 * Create: 2023-05-12
 */
#ifndef _TUPLE_H_
#define _TUPLE_H_

typedef struct {
	__u32 protocol;
	__u32 src_ipv4;
	__u32 src_ipv6[4];
	__u32 src_port;
	__u32 dst_ipv4;
	__u32 dst_ipv6[4];
	__u32 dst_port;
	/* marked nat or rev-nat */
	__u32 flags;
} tuple_t;

#define INIT_TUPLE_IPV4(src, dst, tuple) \
	(tuple).src_ipv4 = (src)->ipv4; \
	(tuple).src_port = (src)->port; \
	(tuple).protocol = (src)->protocol; \
	(tuple).dst_ipv4 = (dst)->ipv4; \
	(tuple).dst_port = (dst)->port

static inline void parse_v4_tuple(struct iphdr* iph, struct tcphdr* tcph, tuple_t* tuple)
{
	tuple->dst_ipv4 = iph->daddr;
	tuple->dst_port = tcph->dest;
	tuple->protocol = iph->protocol;
	tuple->src_ipv4 = iph->saddr;
	tuple->src_port = tcph->source;
}

static inline void parse_v4_sock_tuple(struct bpf_sock_addr *ctx, tuple_t *tuple)
{
	tuple->dst_ipv4 = ctx->user_ip4;
	tuple->dst_port = ctx->user_port;
	tuple->protocol = ctx->protocol;
}

#endif /* _TUPLE_H_ */

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
 * Create: 2023-07-29
 */

#include "slb_common.h"
#include "map.h"

#define SYS_REJECT	0
#define SYS_PROCEED	1

SEC("cgroup/post_bind4")
int check_port(struct bpf_sock *ctx) {

    char fmt[] = "BPF prog1 is called";
    bpf_trace_printk(fmt, sizeof(fmt));
    tuple_t tuple = {0};

    tuple.protocol = ctx->protocol;
    tuple.src_ipv4 = ctx->dst_ip4;
    tuple.src_port = bpf_htons((__u32)ctx->dst_port);
    tuple.dst_ipv4 = ctx->src_ip4;
    tuple.dst_port = bpf_htons(ctx->src_port);

    char msg[] = "protocol:%u";
    char msg2[] = "src: ip:%u, port: %u";
    char msg3[] = "dst: ip:%u, port: %u";
    bpf_trace_printk(msg, sizeof(msg), ctx->protocol);
    bpf_trace_printk(msg3, sizeof(msg3), ctx->src_ip4, bpf_htons(ctx->src_port));
    bpf_trace_printk(msg2, sizeof(msg2), ctx->dst_ip4, bpf_htons(ctx->dst_port));

    if (map_lookup_usedport(&tuple)) {
        char fmt2[] = "return reject!";
        bpf_trace_printk(fmt2, sizeof(fmt2));
        return SYS_REJECT;
    }
    char fmt2[] = "return proceed!";
    bpf_trace_printk(fmt2, sizeof(fmt2));
    return SYS_PROCEED;
}

char _license[] SEC("license") = "GPL";
int _version SEC("version") = 1;
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
 * Author: superCharge
 * Create: 2023-04-22
 */
#include "slb_common.h"
#include "l4_manager.h"

SEC("cgroup/connect4")
int sock_connect4(struct bpf_sock_addr *ctx)
{
	tuple_t tuple = {0};
	parse_v4_sock_tuple(ctx, &tuple);

	(void)l4_manager(ctx, &tuple);
	return CGROUP_SOCK_OK;
}

/* todo udp
SEC("cgroup/sendmsg4")
int sock_sendmsg4(struct bpf_sock_addr *ctx)
{
	BPF_LOG(DEBUG, KMESH, "udp info sock_sendmsg4, userip=%u\n", ctx->user_ip4);
	sock4_traffic_control(ctx);
	return CGROUP_SOCK_OK;
}

SEC("cgroup/recvmsg4")
int sock_recvmsg4(struct bpf_sock_addr *ctx)
{
	BPF_LOG(DEBUG, KMESH, "udp info sock_recvmsg4, userip=%u\n", ctx->user_ip4);
	sock4_recvmsg_control(ctx);
	return CGROUP_SOCK_OK;
}
*/

char _license[] SEC("license") = "GPL";
int _version SEC("version") = 1;

/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
 * MeshAccelerating is licensed under the Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *	 http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Author: nlgwcy
 * Create: 2022-02-14
 */
#include <sys/socket.h>
#include "bpf_log.h"
#include "listener.h"
#include "listener/listener.pb-c.h"

#if KMESH_ENABLE_IPV4
#if KMESH_ENABLE_HTTP

static int sockops_traffic_control(struct bpf_sock_ops *skops, struct bpf_mem_ptr *msg)
{
	/* 1 lookup listener */
	DECLARE_VAR_ADDRESS(skops, addr);
	Listener__Listener *listener = map_lookup_listener(&addr);

	if (!listener) {
		addr.ipv4 = 0;
		listener = map_lookup_listener(&addr);
		if (!listener) {
			/* no match vip/nodeport listener */
			return 0;
		}
	}

	BPF_LOG(DEBUG, SOCKOPS, "sockops_traffic_control listener=\"%s\", addr=[%u:%u]\n",
		(char *)kmesh_get_ptr_val(listener->name), skops->remote_ip4, skops->remote_port);

	(void)bpf_parse_header_msg(msg);
	return l7_listener_manager(skops, listener, msg);
}

SEC("sockops")
int sockops_prog(struct bpf_sock_ops *skops)
{
#define BPF_CONSTRUCT_PTR(low_32, high_32) \
	(unsigned long long)(((unsigned long long)(high_32) << 32) + (low_32))

	struct bpf_mem_ptr *msg = NULL;
	
	if (skops->family != AF_INET)
		return 0;
	
	switch (skops->op) {
		case BPF_SOCK_OPS_TCP_DEFER_CONNECT_CB:
			msg = (struct bpf_mem_ptr *)BPF_CONSTRUCT_PTR(skops->args[0], skops->args[1]);
			(void)sockops_traffic_control(skops, msg);
	}
	return 0;
}

#endif
#endif
char _license[] SEC("license") = "GPL";
int _version SEC("version") = 1;

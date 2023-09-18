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
 * Create: 2023-05-04
 */
#include "slb_common.h"
#include "map.h"
#include "bpf_log.h"
#include "loadbalance.h"
#ifdef XDP_CTX
#include "xdp.h"
#endif

static inline void set_ctx_address(struct bpf_sock_addr *ctx, struct endpoint_entry_t *endpoint)
{
	ctx->user_ip4  = endpoint->ipv4;
	ctx->user_port = endpoint->port;
}

#ifdef CGROUP_SOCK_CTX
static inline int l4_manager(struct bpf_sock_addr *ctx, tuple_t *tuple)
#elif XDP_CTX
static inline int l4_manager(struct xdp_md *ctx,
	struct header_info_t *header_info, tuple_t *tuple)
#endif
{
	int ret;
	struct service_entry_t *service;
	struct service_key_t service_key = {0};
	struct endpoint_entry_t *endpoint;
	service_key.protocol = tuple->protocol;
	service_key.port = tuple->dst_port;
	service_key.ipv4 = tuple->dst_ipv4;
	/* todo ipv6
		bpf_memcpy(key.ipv6, tuple->ipv6, 4);
	*/
	service = map_lookup_service(&service_key);
	if (!service) {
		/* todo ipv6 */
		BPF_LOG(DEBUG, KMESH, "find no listener, dst: address %u, port %u",
				service_key.ipv4, service_key.port);
		return -ENOENT;
	}

	// loadbalance
	endpoint = handle_loadbalance(service);
	if (!endpoint) {
		BPF_LOG(INFO, KMESH, "can not found reachable endpoint");
		return -ENOENT;
	}
	BPF_LOG(DEBUG, KMESH, "endpoint info ip:%u, port:%u, protocol:%u\n",
		endpoint->ipv4, endpoint->port, endpoint->protocol);

#ifdef CGROUP_SOCK_CTX
	set_ctx_address(ctx, endpoint);
	return 0;
#elif XDP_CTX
	return xdp_process_nat(ctx, header_info, tuple, endpoint);
#else
#error "l4 manager miss handle function"
#endif
	return 0;
}
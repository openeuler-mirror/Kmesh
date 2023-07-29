/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

 * Author: nlgwcy
 * Create: 2022-02-15
 */

#include "bpf_log.h"
#include "route_config.h"
#include "tail_call.h"

static inline char *select_weight_cluster(Route__RouteAction *route_act) {
	void *ptr = NULL;
	Route__WeightedCluster *weightedCluster = NULL;
	Route__ClusterWeight *route_cluster_weight = NULL;
	int32_t select_value;
	void *cluster_name = NULL;

	weightedCluster = kmesh_get_ptr_val((route_act->weighted_clusters));
	if (!weightedCluster) {
		return NULL;
	}
	ptr = kmesh_get_ptr_val(weightedCluster->clusters);
	if (!ptr) {
		return NULL;
	}
	select_value = (int)(bpf_get_prandom_u32() % 100);
	for (int i = 0; i < KMESH_PER_WEIGHT_CLUSTER_NUM; i ++) {
		if (i >= weightedCluster->n_clusters) {
			break;
		}
		route_cluster_weight = (Route__ClusterWeight *) kmesh_get_ptr_val(
				(void *) *((__u64 *) ptr + i));
		if (!route_cluster_weight) {
			return NULL;
		}
		select_value = select_value - (int)route_cluster_weight->weight;
		if (select_value <= 0) {
			cluster_name = kmesh_get_ptr_val(route_cluster_weight->name);
			BPF_LOG(DEBUG, ROUTER_CONFIG, "select cluster, name:weight %s:%d\n",
					cluster_name, route_cluster_weight->weight);
			return cluster_name;
		}
	}
	return NULL;
}

static inline char *route_get_cluster(const Route__Route *route)
{
	Route__RouteAction *route_act = NULL;
	route_act = kmesh_get_ptr_val(_(route->route));
	if (!route_act) {
		BPF_LOG(ERR, ROUTER_CONFIG, "failed to get route action ptr\n");
		return NULL;
	}

	if (route_act->cluster_specifier_case == ROUTE__ROUTE_ACTION__CLUSTER_SPECIFIER_WEIGHTED_CLUSTERS) {
		return select_weight_cluster(route_act);
	}

	return kmesh_get_ptr_val(_(route_act->cluster));
}

SEC_TAIL(KMESH_SOCKOPS_CALLS, KMESH_TAIL_CALL_ROUTER_CONFIG)
int route_config_manager(ctx_buff_t *ctx)
{
	int ret;
	char *cluster = NULL;
	ctx_key_t ctx_key = {0};
	ctx_val_t *ctx_val = NULL;
	ctx_val_t ctx_val_1 = {0};
	Route__RouteConfiguration *route_config = NULL;
	Route__VirtualHost *virt_host = NULL;
	Route__Route *route = NULL;

	DECLARE_VAR_ADDRESS(ctx, addr);
	ctx_key.address = addr;
	ctx_key.tail_call_index = KMESH_TAIL_CALL_ROUTER_CONFIG + bpf_get_current_task();
	ctx_val = kmesh_tail_lookup_ctx(&ctx_key);
	if (!ctx_val)
		return convert_sockops_ret(-1);

	route_config = map_lookup_route_config(ctx_val->data);
	kmesh_tail_delete_ctx(&ctx_key);
	if (!route_config) {
		BPF_LOG(WARN, ROUTER_CONFIG, "failed to lookup route config, route_name=\"%s\"\n", ctx_val->data);
		return convert_sockops_ret(-1);
	}

	virt_host = virtual_host_match(route_config, &addr, ctx);
	if (!virt_host) {
		BPF_LOG(ERR, ROUTER_CONFIG, "failed to match virtual host, addr=%u\n", addr.ipv4);
		return convert_sockops_ret(-1);
	}

	route = virtual_host_route_match(virt_host, &addr, ctx, (struct bpf_mem_ptr *)ctx_val->msg);
	if (!route) {
		BPF_LOG(ERR, ROUTER_CONFIG, "failed to match route action, addr=%u\n", addr.ipv4);
		return convert_sockops_ret(-1);
	}

	cluster = route_get_cluster(route);
	if (!cluster) {
		BPF_LOG(ERR, ROUTER_CONFIG, "failed to get cluster\n");
		return convert_sockops_ret(-1);
	}

	ctx_key.address = addr;
	ctx_key.tail_call_index = KMESH_TAIL_CALL_CLUSTER + bpf_get_current_task();
	struct bpf_mem_ptr data_tmp = {
		.ptr = ctx_val_1.data
	};
	if (!bpf__strncpy(&data_tmp, sizeof(struct bpf_mem_ptr), cluster, BPF_DATA_MAX_LEN)) {
		BPF_LOG(ERR, ROUTER_CONFIG, "failed to copy cluster %s\n", cluster);
		return convert_sockops_ret(-1);
	}

	ret = kmesh_tail_update_ctx(&ctx_key, &ctx_val_1);
	if (ret != 0)
		return convert_sockops_ret(ret);

	kmesh_tail_call(ctx, KMESH_TAIL_CALL_CLUSTER);
	kmesh_tail_delete_ctx(&ctx_key);
	return 0;
}

char _license[] SEC("license") = "GPL";
int _version SEC("version") = 1;

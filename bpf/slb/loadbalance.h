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
#include "slb_common.h"
#include "map.h"
#include "bpf_log.h"

#include "lb_method/lb_roundrobin.h"
#include "lb_method/lb_random.h"

static inline struct endpoint_entry_t *
handle_loadbalance(struct service_entry_t *service)
{
	struct endpoint_entry_t *endpoint = NULL;
	BPF_LOG(DEBUG, KMESH, "debug load balance type:%d",
		service->policy);
	switch (service->policy) {
		case LB_POLICY_RANDOM:
			endpoint = slb_lb_random(service);
			break;
		case LB_POLICY_ROUND_ROBIN:
			endpoint = slb_lb_roundrobin(service);
			break;
		defalut:
			BPF_LOG(INFO, KMESH, "unsupport lb policy! type:%u\n", service->policy);
			break;
	}

	return endpoint;
}

// TODO xdp lb_roundbin
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

static inline struct endpoint_entry_t *
slb_lb_roundrobin(struct service_entry_t *service)
{
	struct backend_key_t backend_key = {
		.service_id = service->service_id,
	};
	struct backend_entry_t *backend;
	endpoint_key_t endpoint_key;

	backend_key.backend_slot = service->current_backend_slot %
			service->count;

	__u32 *current_backend_slot = &service->current_backend_slot;
	__sync_fetch_and_add(current_backend_slot, 1);

	backend = map_lookup_backend(&backend_key);
	if (!backend) {
		BPF_LOG(INFO, KMESH, "can not found reachable backend");
		return NULL;
	}

	endpoint_key = backend->endpoint_id;

	return map_lookup_endpoint(&endpoint_key);
}
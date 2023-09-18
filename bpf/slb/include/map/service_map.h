/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
 * MeshAccelerating is licensed under the Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Author: bitcoffee
 * Create: 2023-05-12
 */
#ifndef _SERVICE_MAP_H_
#define _SERVICE_MAP_H_

#include "slb_common.h"
#include "map_data_v1/service.h"
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, struct service_key_t);
	__type(value, struct service_entry_t);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, MAP_SIZE_OF_SERVICE);
	__uint(map_flags, 0);
} map_of_service SEC(".maps");

static inline struct service_entry_t *map_lookup_service(const struct service_key_t *map_key)
{
	return bpf_map_lookup_elem(&map_of_service, map_key);
}

#endif /* _SERVICE_MAP_H_ */
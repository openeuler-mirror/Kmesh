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
 * Create: 2023-07-21
 */
#ifndef _USEDPORT_MAP_H_
#define _USEDPORT_MAP_H_

#include "slb_common.h"
#include "tuple.h"
#include <linux/bpf.h>

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, tuple_t);
	__type(value, __u32); // record create time
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, MAP_SIZE_OF_CONNTRACK);
	__uint(map_flags, 0);
} map_of_usedport SEC(".maps");

static inline __u32 *map_lookup_usedport(const tuple_t *tuple)
{
	return bpf_map_lookup_elem(&map_of_usedport, tuple);
}

static inline int map_update_usedport(const tuple_t *key, __u32 *value)
{
	return bpf_map_update_elem(&map_of_usedport, key, value, BPF_NOEXIST);
}

static inline void map_delete_usedport(const tuple_t *key)
{
	(void)bpf_map_delete_elem(&map_of_usedport, key);
}

#endif /* _SERVICE_MAP_H_ */
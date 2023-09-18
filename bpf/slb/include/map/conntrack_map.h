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
 * Create: 2023-06-15
 */

#ifndef _CONNTRACK_MAP_H_
#define _CONNTRACK_MAP_H_
#include "slb_common.h"
#include "tuple.h"
#include "conntrack.h"

struct {
	__uint(type, BPF_MAP_TYPE_LRU_HASH);
	__type(key, tuple_t);
	__type(value, struct ct_value_t);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, MAP_SIZE_OF_CONNTRACK);
	__uint(map_flags, 0);
} map_of_ct SEC(".maps");

static inline struct ct_value_t *map_lookup_ct(const tuple_t *map_key)
{
	return bpf_map_lookup_elem(&map_of_ct, map_key);
}

static inline int map_update_ct(const tuple_t *map_key, const struct ct_value_t *value)
{
	return bpf_map_update_elem(&map_of_ct, map_key, value, BPF_ANY);
}

static inline int map_delete_ct(const tuple_t *map_key)
{
	return bpf_map_delete_elem(&map_of_ct, map_key);
}

#endif /* _CONNTRACK_MAP_H_ */

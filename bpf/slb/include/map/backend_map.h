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
#ifndef _BACKEND_MAP_H
#define _BACKEND_MAP_H

#include "slb_common.h"
#include "map_data_v1/backend.h"
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, struct backend_key_t);
	__type(value, struct backend_entry_t);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, MAP_SIZE_OF_BACKEND);
	__uint(map_flags, 0);
} map_of_backend SEC(".maps");

static inline struct backend_entry_t *map_lookup_backend(const struct backend_key_t *map_key)
{
	return bpf_map_lookup_elem(&map_of_backend, map_key);
}

#endif /* _BACKEND_MAP_H */
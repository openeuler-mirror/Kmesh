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
#ifndef _LBCONFIG_MAP_H_
#define _LBCONFIG_MAP_H_

#include "slb_common.h"
#include "map_data_v1/lbconfig.h"

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, lbconfig_key_t);
    __type(value, struct lbconfig_entry_t); 
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(max_entries, MAP_SIZE_OF_LBCONFIG);
    __uint(map_flags, 0);
} map_of_lbconfig SEC(".maps");

static inline struct lbconfig_entry_t *map_lookup_lbconfig(const lbconfig_key_t *map_key)
{
    return bpf_map_lookup_elem(&map_of_lbconfig, map_key);
}

static inline int map_update_lbconfig(const lbconfig_key_t *map_key, const struct lbconfig_entry_t *value)
{
    return bpf_map_update_elem(&map_of_lbconfig, map_key, value, BPF_NOEXIST);
}

static inline void map_delete_lbconfig(const lbconfig_key_t *map_key)
{
    (void)bpf_map_delete_elem(&map_of_lbconfig, map_key);
}

#endif /* _LBCONFIG_MAP_H_ */
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
#ifndef _SERVICE_H_
#define _SERVICE_H_

#include <linux/types.h>

struct service_key_t {
    __u16 protocol;
    // host order
    __u32 port;
    __u32 ipv4;
    __u32 ipv6[4];
}__attribute__((packed));

struct service_entry_t {
    // service id, global unique
    __u32 service_id;
    // endpoint number
    __u32 count;
    /*
     * this field is used in round robin mode
     * to record the ID of the selected backend.
     */
    __u32 current_backend_slot;
    // loadbalance policy
    __u16 policy;
}__attribute__((packed));

#endif /* _SERVICE_H_ */
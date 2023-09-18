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
#ifndef _ENDPOINT_H_
#define _ENDPOINT_H_

#include <linux/types.h>

typedef __u32 endpoint_key_t;
struct endpoint_entry_t {
    __u16 protocol;
    __u32 port;
    __u32 ipv4;
    // reserved field, record the node where the pod is located
    __u32 node_ip;
    // reserved field, record the endpoint belong current node
    __u8  is_local;
    // reserved field, 
    __u8  pad[3];
}__attribute__((packed));

#endif
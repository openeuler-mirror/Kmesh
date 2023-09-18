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
#ifndef _CONFIG_H_
#define _CONFIG_H_

// ************
// options
#define KMESH_MODULE_ON			1
#define KMESH_MODULE_OFF		0

// L3
#define KMESH_ENABLE_IPV4		KMESH_MODULE_ON
#define KMESH_ENABLE_IPV6		KMESH_MODULE_OFF
// L4
#define KMESH_ENABLE_TCP		KMESH_MODULE_ON
#define KMESH_ENABLE_UDP		KMESH_MODULE_ON
// L7
#define KMESH_ENABLE_HTTP		KMESH_MODULE_OFF
#define KMESH_ENABLE_HTTPS		KMESH_MODULE_OFF

#define MAP_SIZE_OF_SERVICE     10240
#define MAP_SIZE_OF_BACKEND     65536
#define MAP_SIZE_OF_ENDPOINT    65536
#define MAP_SIZE_OF_CONNTRACK   102400

#define map_of_service			slb_service
#define map_of_backend			slb_backend
#define map_of_endpoint			slb_endpoint
#define map_of_loadbalance		slb_loadbalance
#define map_of_ct			slb_ct
#define map_of_usedport			slb_usedport

#endif /*_CONFIG_H_*/

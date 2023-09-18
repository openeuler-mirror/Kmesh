/*
 * Copyright (c) 2019 Huawei Technologies Co., Ltd.
 * MeshAccelerating is licensed under the Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Author: superCharge
 * Create: 2023-06-22
 */

package cache_v1

import "openeuler.io/mesh/pkg/logger"

var (
	log = logger.NewLoggerField("cache/v1")
)

const (
	CacheFlagNone   CacheOptionFlag = 0x00
	CacheFlagDelete CacheOptionFlag = 0x01
	CacheFlagUpdate CacheOptionFlag = 0x02
	CacheFlagAll    CacheOptionFlag = CacheFlagDelete | CacheFlagUpdate
)

type CacheOptionFlag uint
type CacheCount map[uint32]uint32                                          // k = port
type EndpointID2BackendSlot map[uint32]uint32                              // k = endpoint_id v = backend_slot
type ServiceCacheEndpointID2BackendSlots map[uint32]EndpointID2BackendSlot // k = sercieid v = endpoint2backendslot map

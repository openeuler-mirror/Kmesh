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

import (
	"fmt"

	map_api_v1 "openeuler.io/mesh/include/map_data_v1/go"
	maps_v1 "openeuler.io/mesh/pkg/cache/v1/maps"
)

type EndpointKeyAndValue struct {
	Key   uint32
	Value map_api_v1.EndpointEntry
}

func (kv *EndpointKeyAndValue) packUpdate() error {
	if err := maps_v1.EndpointUpdate(&kv.Key, &kv.Value); err != nil {
		return fmt.Errorf("update endpoint failed, %v, %s", kv.Key, err)
	}
	return nil
}

func (kv *EndpointKeyAndValue) packDelete() error {
	if err := maps_v1.EndpointDelete(&kv.Key); err != nil {
		return fmt.Errorf("delete endpoint failed, %v, %s", kv.Key, err)
	}
	return nil
}

type EndpointCache map[EndpointKeyAndValue]CacheOptionFlag

func (cache EndpointCache) StatusFlush(flag CacheOptionFlag) {
	var err error

	for kv, f := range cache {
		if f != flag {
			continue
		}

		switch flag {
		case CacheFlagDelete:
			err = kv.packDelete()
		case CacheFlagUpdate:
			err = kv.packUpdate()
			cache[kv] = CacheFlagNone
		default:
		}

		if err != nil {
			log.Errorln(err)
		}
	}

	if flag == CacheFlagDelete {
		cache.StatusDelete(flag)
	}
}

func (cache EndpointCache) StatusDelete(flag CacheOptionFlag) {
	for kv, f := range cache {
		if f == flag {
			delete(cache, kv)
		}
	}
}

func (cache EndpointCache) StatusReset(old, new CacheOptionFlag) {
	for kv, f := range cache {
		if f == old {
			cache[kv] = new
		}
	}
}

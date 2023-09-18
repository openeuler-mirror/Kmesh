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
 * Author: bitcoffee
 * Create: 2023-06-16
 */

package cache_v1

import (
	"fmt"

	map_api_v1 "openeuler.io/mesh/include/map_data_v1/go"
	maps_v1 "openeuler.io/mesh/pkg/cache/v1/maps"
)

type ServiceKeyAndValue struct {
	Key   map_api_v1.ServiceKey
	Value map_api_v1.ServiceEntry
}

func (kv *ServiceKeyAndValue) packUpdate() error {
	if err := maps_v1.ServiceUpdate(&kv.Key, &kv.Value); err != nil {
		return fmt.Errorf("update listener failed, %v, %s", kv.Key, err)
	}
	return nil
}

func (kv *ServiceKeyAndValue) packDelete() error {
	if err := maps_v1.ServiceDelete(&kv.Key); err != nil {
		return fmt.Errorf("delete listener failed, %v, %s", kv.Key, err)
	}
	return nil
}

type ServiceCache map[ServiceKeyAndValue]CacheOptionFlag

func (cache ServiceCache) StatusFlush(flag CacheOptionFlag) {
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
		default:
		}

		if err != nil {
			log.Errorln(err)
		}
	}
}

func (cache ServiceCache) StatusDelete(flag CacheOptionFlag) {
	for kv, f := range cache {
		if f == flag {
			delete(cache, kv)
		}
	}
}

func (cache ServiceCache) StatusReset(old, new CacheOptionFlag) {
	for kv, f := range cache {
		if f == old {
			cache[kv] = new
		}
	}
}

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

package maps

import (
	"github.com/cilium/ebpf"
	map_api_v1 "openeuler.io/mesh/include/map_data_v1/go"
	"openeuler.io/mesh/pkg/bpf"
	"openeuler.io/mesh/pkg/logger"
)

var (
	log = logger.NewLoggerField("cache/v1/maps")
)

func ServiceUpdate(key *map_api_v1.ServiceKey, value *map_api_v1.ServiceEntry) error {
	log.Debugf("Update [%#v], [%#v]", *key, *value)
	return bpf.Obj.CgroupSock.CgroupSockObjects.CgroupSockMaps.SlbService.
		Update(key, value, ebpf.UpdateAny)
}

func ServiceDelete(key *map_api_v1.ServiceKey) error {
	log.Debugf("Delete [%#v]", *key)
	return bpf.Obj.CgroupSock.CgroupSockObjects.CgroupSockMaps.SlbService.
		Delete(key)
}
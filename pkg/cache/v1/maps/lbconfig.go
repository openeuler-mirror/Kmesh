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
 * Author: dongdong9
 * Create: 2023-08-13
 */

package maps

import (
	"github.com/cilium/ebpf"
	map_api_v1 "openeuler.io/mesh/include/map_data_v1/go"
	"openeuler.io/mesh/pkg/bpf"
)

func LbConfigUpdate(key *map_api_v1.LbConfigKey, value *map_api_v1.LbConfigEntry) error {
	log.Debugf("Update [%#v], [%#v]", *key, *value)
	return bpf.Obj.CgroupSock.CgroupSockObjects.CgroupSockMaps.SlbConfig.
		Update(key, value, ebpf.UpdateAny)
}

func LbConfigDelete(key *map_api_v1.LbConfigKey) error {
	log.Debugf("Delete [%#v]", *key)
	return bpf.Obj.CgroupSock.CgroupSockObjects.CgroupSockMaps.SlbConfig.
		Delete(key)
}

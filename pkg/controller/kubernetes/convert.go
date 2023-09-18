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
 * Create: 2023-04-24
 */

package kubernetes

import (
	"sort"
	"strconv"

	api_core_v1 "k8s.io/api/core/v1"
	map_api_v1 "openeuler.io/mesh/include/map_data_v1/go"
	cache_v1 "openeuler.io/mesh/pkg/cache/v1"
	"openeuler.io/mesh/pkg/nets"
)

const (
	DefaultConnectTimeOut = 15
	ConverNumBase         = 10
	MaxEndpointNum        = 2 << 16
)

var (
	ProtocolStrToUint = map[string]uint16{
		"TCP": 6,  // IPPROTO_TCP
		"UDP": 17, // IPPROTO_UDP,
	}
)

func extractEndpointCache(epcache cache_v1.EndpointCache,
	flag cache_v1.CacheOptionFlag, ep *api_core_v1.Endpoints) {
	var epkv cache_v1.EndpointKeyAndValue

	if ep == nil {
		return
	}

	for i, sub := range ep.Subsets {
		for j, epPort := range sub.Ports {
			if !nets.GetConfig().IsEnabledProtocol(string(epPort.Protocol)) {
				continue
			}

			epkv.Value.Protocol = ProtocolStrToUint[string(epPort.Protocol)]
			epkv.Value.Port = nets.ConvertPortToBigEndian(uint32(epPort.Port))
			for k, epAddr := range sub.Addresses {
				epkv.Value.IPv4 = nets.ConvertIpToUint32(epAddr.IP)
				epkv.Key = hashName.StrToNum(epPort.Name +
					strconv.FormatUint(uint64(epkv.Value.IPv4), ConverNumBase) +
					strconv.FormatUint(uint64(epkv.Value.Port), ConverNumBase))
				endpointNum := i + j + k
				if endpointNum > MaxEndpointNum {
					log.Errorln("endpoint too much, over 2^16")
					break
				}
				epcache[epkv] |= flag
			}
		}
	}
}

func extractBackendCache(beCache cache_v1.BackendCache, svcNameID uint32, endpointNum *uint32,
	epCache cache_v1.EndpointCache, endpointIDToBackendSlot cache_v1.EndpointID2BackendSlot) {

	var kv cache_v1.BackendKeyAndValue
	var tailkv cache_v1.BackendKeyAndValue
	var ok bool
	var i int

	var deleteSlotSlice []cache_v1.BackendKeyAndValue
	var keepSlotSlice []cache_v1.BackendKeyAndValue
	var newSlotSlice []cache_v1.BackendKeyAndValue

	kv.Key.Serviceid = svcNameID
	tailkv.Key.Serviceid = svcNameID
	for endpointKeyAndValue, flag := range epCache {
		kv.Value.Endpointid = endpointKeyAndValue.Key
		kv.Key.Backslot, ok = endpointIDToBackendSlot[endpointKeyAndValue.Key]
		switch flag {
		case cache_v1.CacheFlagDelete:
			if !ok {
				log.Errorln("can not found slot when cache in status delete")
				continue
			}
			delete(endpointIDToBackendSlot, endpointKeyAndValue.Key)
			deleteSlotSlice = append(deleteSlotSlice, kv)
		case cache_v1.CacheFlagUpdate:
			if ok {
				log.Errorln("found a exists backslot when insert a new endpoint")
				continue
			}
			kv.Key.Backslot = uint32(len(endpointIDToBackendSlot))
			endpointIDToBackendSlot[endpointKeyAndValue.Key] = kv.Key.Backslot
			newSlotSlice = append(newSlotSlice, kv)
		case cache_v1.CacheFlagAll:
			if !ok {
				log.Errorln("can not found slot when cache in status all")
				continue
			}
			keepSlotSlice = append(keepSlotSlice, kv)
		}
	}
	sort.Slice(deleteSlotSlice, func(i, j int) bool { return deleteSlotSlice[i].Key.Backslot < deleteSlotSlice[j].Key.Backslot })
	sort.Slice(keepSlotSlice, func(i, j int) bool { return keepSlotSlice[i].Key.Backslot > keepSlotSlice[j].Key.Backslot })
	sort.Slice(newSlotSlice, func(i, j int) bool { return newSlotSlice[i].Key.Backslot > newSlotSlice[j].Key.Backslot })

	*endpointNum = uint32(len(keepSlotSlice) + len(newSlotSlice))

	for ; i < len(deleteSlotSlice) && i < len(newSlotSlice); i++ {
		kv.Key.Backslot = deleteSlotSlice[i].Key.Backslot
		kv.Value.Endpointid = newSlotSlice[i].Value.Endpointid
		beCache[kv] |= cache_v1.CacheFlagUpdate
		endpointIDToBackendSlot[newSlotSlice[i].Value.Endpointid] = kv.Key.Backslot
		delete(endpointIDToBackendSlot, deleteSlotSlice[i].Value.Endpointid)
	}

	for ; i < len(deleteSlotSlice) && i < len(keepSlotSlice); i++ {
		if deleteSlotSlice[i].Key.Backslot > keepSlotSlice[i].Key.Backslot {
			break
		}
		kv.Key.Backslot = deleteSlotSlice[i].Key.Backslot
		kv.Value.Endpointid = keepSlotSlice[i].Value.Endpointid
		beCache[kv] |= cache_v1.CacheFlagUpdate

		tailkv.Key.Backslot = keepSlotSlice[i].Key.Backslot
		tailkv.Value.Endpointid = keepSlotSlice[i].Value.Endpointid
		beCache[tailkv] |= cache_v1.CacheFlagDelete
		endpointIDToBackendSlot[keepSlotSlice[i].Value.Endpointid] = kv.Key.Backslot
		delete(endpointIDToBackendSlot, deleteSlotSlice[i].Value.Endpointid)
	}

	for ; i < len(newSlotSlice); i++ {
		beCache[newSlotSlice[i]] |= cache_v1.CacheFlagUpdate
	}

	for ; i < len(deleteSlotSlice); i++ {
		beCache[deleteSlotSlice[i]] |= cache_v1.CacheFlagDelete
		delete(endpointIDToBackendSlot, deleteSlotSlice[i].Value.Endpointid)
	}
}

func extractServiceCache(svcKeys map[map_api_v1.ServiceKey]cache_v1.CacheOptionFlag,
	svcFlag cache_v1.CacheOptionFlag, svcNameID uint32, svc *api_core_v1.Service, addr nodeAddress) {
	if svc == nil {
		return
	}
	var kv map_api_v1.ServiceKey
	for k, _ := range svcKeys {
		svcKeys[k] = cache_v1.CacheFlagDelete
	}

	for _, serPort := range svc.Spec.Ports {
		if !nets.GetConfig().IsEnabledProtocol(string(serPort.Protocol)) {
			continue
		}
		kv.Protocol = ProtocolStrToUint[string(serPort.Protocol)]
		switch svc.Spec.Type {
		case api_core_v1.ServiceTypeNodePort:
			kv.Port = nets.ConvertPortToBigEndian(uint32(serPort.NodePort))
			for ip, nodeFlag := range addr {
				kv.IPv4 = nets.ConvertIpToUint32(ip)

				if svcFlag != cache_v1.CacheFlagNone {
					svcKeys[kv] |= svcFlag
				} else if nodeFlag != cache_v1.CacheFlagNone {
					svcKeys[kv] |= nodeFlag
				}
			}
			fallthrough
		case api_core_v1.ServiceTypeClusterIP:
			if svcFlag != cache_v1.CacheFlagNone {
				kv.Port = nets.ConvertPortToBigEndian(uint32(serPort.Port))
				// TODO: Service.Spec.ExternalIPs ??
				kv.IPv4 = nets.ConvertIpToUint32(svc.Spec.ClusterIP)

				svcKeys[kv] |= svcFlag
			}
		case api_core_v1.ServiceTypeLoadBalancer:
			// TODO
		case api_core_v1.ServiceTypeExternalName:
			// TODO
		default:
			// ignore
		}
	}
	for k, flag := range svcKeys {
		if flag == cache_v1.CacheFlagDelete {
			delete(svcKeys, k)
		}
	}
}

func updateServiceEndpointNum(svcache cache_v1.ServiceCache, endpointNum uint32,
	svcKeys map[map_api_v1.ServiceKey]cache_v1.CacheOptionFlag, svcNameID uint32) {

	var kv cache_v1.ServiceKeyAndValue

	kv.Value.EndpointNum = endpointNum
	kv.Value.LoadbalanceType = 0
	kv.Value.Serviceid = svcNameID
	for k, _ := range svcKeys {
		kv.Key = k
		svcache[kv] |= cache_v1.CacheFlagUpdate
	}
}

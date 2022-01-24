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
 * Author: LemmyHuang
 * Create: 2022-01-24
 */

package kubernetes

// #cgo CFLAGS: -I../../../bpf/include
// #include "listener_type.h"
import "C"
import "C"
import (
	apiCoreV1 "k8s.io/api/core/v1"
	"openeuler.io/mesh/pkg/api"
	"openeuler.io/mesh/pkg/api/types"
	"openeuler.io/mesh/pkg/nets"
)

func extractEndpointCache(cache api.EndpointCache, flag api.CacheOptionFlag, nameID uint32, ep *apiCoreV1.Endpoints) {
	var kv api.EndpointKeyAndValue

	if ep == nil {
		return
	}
	kv.Key.NameID = nameID

	for _, sub := range ep.Subsets {
		for _, epPort := range sub.Ports {
			if !nets.GetConfig().IsEnabledProtocol(string(epPort.Protocol)) {
				continue
			}

			kv.Value.Address.Protocol = types.ProtocolStrToC[string(epPort.Protocol)]
			kv.Value.Address.Port = nets.ConvertPortToLittleEndian(uint32(epPort.Port))
			kv.Key.Port = kv.Value.Address.Port

			for _, epAddr := range sub.Addresses {
				kv.Value.Address.IPv4 = nets.ConvertIpToUint32(epAddr.IP)
				cache[kv] |= flag
			}
		}
	}
}

func extractClusterCache(cache api.ClusterCache, flag api.CacheOptionFlag, nameID uint32, svc *apiCoreV1.Service) {
	var kv api.ClusterKeyAndValue

	if svc == nil {
		return
	}

	kv.Key.NameID = nameID
	kv.Value.LoadAssignment.MapKeyOfEndpoint.NameID = nameID
	// TODO
	kv.Value.Type = 0
	kv.Value.ConnectTimeout = 15

	for _, serPort := range svc.Spec.Ports {
		if !nets.GetConfig().IsEnabledProtocol(string(serPort.Protocol)) {
			continue
		}

		kv.Value.LoadAssignment.MapKeyOfEndpoint.Port = nets.ConvertPortToLittleEndian(uint32(serPort.TargetPort.IntVal))
		kv.Key.Port = nets.ConvertPortToLittleEndian(uint32(serPort.Port))

		cache[kv] |= flag
	}
}

func extractListenerCache(cache api.ListenerCache, svcFlag api.CacheOptionFlag, nameID uint32,
						  svc *apiCoreV1.Service, addr nodeAddress) {
	var kv api.ListenerKeyAndValue

	if svc == nil {
		return
	}

	kv.Value.MapKey.NameID = nameID
	kv.Value.Type = C.LISTENER_TYPE_DYNAMIC
	kv.Value.State = C.LISTENER_STATE_ACTIVE

	for _, serPort := range svc.Spec.Ports {
		if !nets.GetConfig().IsEnabledProtocol(string(serPort.Protocol)) {
			continue
		}

		// TODO: goListener.Address.Protocol = ProtocolStrToC[serPort.Protocol]
		kv.Value.MapKey.Port = nets.ConvertPortToLittleEndian(uint32(serPort.Port))

		switch svc.Spec.Type {
		case apiCoreV1.ServiceTypeNodePort:
			kv.Key.Port = nets.ConvertPortToLittleEndian(uint32(serPort.NodePort))
			for ip, nodeFlag := range addr {
				kv.Key.IPv4 = nets.ConvertIpToUint32(ip)
				kv.Value.Address = kv.Key

				if svcFlag != api.CacheFlagNone {
					cache[kv] |= svcFlag
				} else if nodeFlag != api.CacheFlagNone {
					cache[kv] |= nodeFlag
				}
			}
			fallthrough
		case apiCoreV1.ServiceTypeClusterIP:
			if svcFlag != 0 {
				kv.Key.Port = nets.ConvertPortToLittleEndian(uint32(serPort.Port))
				// TODO: Service.Spec.ExternalIPs ??
				kv.Key.IPv4 = nets.ConvertIpToUint32(svc.Spec.ClusterIP)

				kv.Value.Address = kv.Key
				cache[kv] |= svcFlag
			}
		case apiCoreV1.ServiceTypeLoadBalancer:
			// TODO
		case apiCoreV1.ServiceTypeExternalName:
			// TODO
		default:
			// ignore
		}
	}
}

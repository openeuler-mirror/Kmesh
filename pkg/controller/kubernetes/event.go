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
 * Create: 2023-05-09
 */

package kubernetes

import (
	api_core_v1 "k8s.io/api/core/v1"
	map_api_v1 "openeuler.io/mesh/include/map_data_v1/go"
	cache_v1 "openeuler.io/mesh/pkg/cache/v1"
)

var (
	hashName                                                                          = cache_v1.NewHashName()
	nodeHdl                                                                           = newNodeHandle()
	service2SlotCache cache_v1.ServiceCacheEndpointID2BackendSlots                    = make(cache_v1.ServiceCacheEndpointID2BackendSlots)
	serviceKeyCaches  map[uint32](map[map_api_v1.ServiceKey]cache_v1.CacheOptionFlag) = make(map[uint32](map[map_api_v1.ServiceKey]cache_v1.CacheOptionFlag))
	endpointCaches    map[uint32]cache_v1.EndpointCache                               = make(map[uint32]cache_v1.EndpointCache)
)

type serviceHandle struct {
	name           string
	service        *serviceEvent
	endpointEvents []*endpointEvent

	serviceEventCount   cache_v1.CacheCount
	endpointsEventCount cache_v1.CacheCount

	calcEndpointCount uint32
	endpointsCount    uint32
	// When you want to delete endpoint from the map,
	// you need to convert the address to key first.

}

func newServiceHandle(name string) *serviceHandle {
	return &serviceHandle{
		name:              name,
		calcEndpointCount: 0,
		endpointsCount:    0,
	}
}

func (svc *serviceHandle) destroy() {
	hashName.Delete(svc.name)
	*svc = serviceHandle{}
}

func (svc *serviceHandle) isEmpty() bool {
	for _, c := range svc.serviceEventCount {
		if c > 0 {
			return false
		}
	}
	for _, c := range svc.endpointsEventCount {
		if c > 0 {
			return false
		}
	}

	return true
}

func (svc *serviceHandle) isChange() bool {
	if svc.service != nil {
		return true
	}
	if len(svc.endpointEvents) > 0 {
		return true
	}

	return false
}

func (svc *serviceHandle) batchProcess(addr nodeAddress) {
	svcNameID := hashName.StrToNum(svc.name)

	// get service cache
	svcKeyCache, ok := serviceKeyCaches[svcNameID]
	if !ok {
		serviceKeyCaches[svcNameID] = make(map[map_api_v1.ServiceKey]cache_v1.CacheOptionFlag)
		svcKeyCache = serviceKeyCaches[svcNameID]
	}
	defer func() {
		if len(svcKeyCache) == 0 {
			delete(serviceKeyCaches, svcNameID)
		}
	}()

	// get endpoint cache
	epCache, ok := endpointCaches[svcNameID]
	if !ok {
		endpointCaches[svcNameID] = make(cache_v1.EndpointCache)
		epCache = endpointCaches[svcNameID]
	}
	defer func() {
		if len(epCache) == 0 {
			delete(endpointCaches, svcNameID)
		}
	}()

	beCache := make(cache_v1.BackendCache)
	defer func() { beCache = nil }()
	svcCache := make(cache_v1.ServiceCache)
	defer func() { svcCache = nil }()
	var endpointNum uint32 = 0

	epCache.StatusReset(cache_v1.CacheFlagNone, cache_v1.CacheFlagDelete)

	for k, epEvent := range svc.endpointEvents {
		if epEvent == nil {
			continue
		}
		if k == (len(svc.endpointEvents) - 1) {
			extractEndpointCache(epCache, cache_v1.CacheFlagUpdate, epEvent.newObj)
		}
		epEvent.destroy()
		svc.endpointEvents[k] = nil
	}
	// clear endpoints all elem
	if svc.endpointEvents != nil {
		svc.endpointEvents = svc.endpointEvents[:0]
	}

	endpointIDToBackendSlot, ok := service2SlotCache[svcNameID]
	if !ok {
		service2SlotCache[svcNameID] = make(cache_v1.EndpointID2BackendSlot)
		endpointIDToBackendSlot = service2SlotCache[svcNameID]
	}
	defer func() {
		if len(endpointIDToBackendSlot) == 0 {
			delete(service2SlotCache, svcNameID)
		}
	}()

	extractBackendCache(beCache, svcNameID, &endpointNum, epCache, endpointIDToBackendSlot)

	if svc.service != nil {
		extractServiceCache(svcKeyCache, cache_v1.CacheFlagUpdate, svcNameID, svc.service.newObj, addr)

		svc.service.destroy()
		svc.service = nil
	}

	updateServiceEndpointNum(svcCache, endpointNum, svcKeyCache, svcNameID)

	// update all map
	epCache.StatusFlush(cache_v1.CacheFlagUpdate)
	beCache.StatusFlush(cache_v1.CacheFlagUpdate)
	svcCache.StatusFlush(cache_v1.CacheFlagUpdate)

	svcCache.StatusFlush(cache_v1.CacheFlagDelete)
	beCache.StatusFlush(cache_v1.CacheFlagDelete)
	epCache.StatusFlush(cache_v1.CacheFlagDelete)

	// clear status

	epCache.StatusDelete(cache_v1.CacheFlagDelete)
	epCache.StatusReset(cache_v1.CacheFlagAll, cache_v1.CacheFlagNone)
	epCache.StatusReset(cache_v1.CacheFlagUpdate, cache_v1.CacheFlagNone)
}

type endpointEvent struct {
	oldObj *api_core_v1.Endpoints
	newObj *api_core_v1.Endpoints
}

func newEndpointEvent(oldObj, newObj interface{}) *endpointEvent {
	event := &endpointEvent{}

	if oldObj == nil && newObj == nil {
		return nil
	}

	if oldObj != nil {
		event.oldObj = oldObj.(*api_core_v1.Endpoints)
	}
	if newObj != nil {
		event.newObj = newObj.(*api_core_v1.Endpoints)
	}

	return event
}

func (event *endpointEvent) destroy() {
	*event = endpointEvent{}
}

type serviceEvent struct {
	oldObj *api_core_v1.Service
	newObj *api_core_v1.Service
}

func newServiceEvent(oldObj, newObj interface{}) *serviceEvent {
	event := &serviceEvent{}

	if oldObj == nil && newObj == nil {
		return nil
	}

	if oldObj != nil {
		event.oldObj = oldObj.(*api_core_v1.Service)
	}
	if newObj != nil {
		event.newObj = newObj.(*api_core_v1.Service)
	}

	return event
}

func (event *serviceEvent) destroy() {
	*event = serviceEvent{}
}

// k = name
type nodeService map[string]*api_core_v1.Service

// k = ip
type nodeAddress map[string]cache_v1.CacheOptionFlag

type nodeHandle struct {
	// Mark node changes only
	isChange bool
	service  nodeService
	address  nodeAddress
}

func newNodeHandle() *nodeHandle {
	return &nodeHandle{
		isChange: false,
		service:  make(nodeService),
		address:  make(nodeAddress),
	}
}

func (nd *nodeHandle) destroy() {
	nd.service = nil
	nd.address = nil
}

func (nd *nodeHandle) refreshService(name string, oldObj, newObj *api_core_v1.Service) {
	if oldObj != nil && newObj == nil {
		delete(nd.service, name)
	} else if newObj != nil {
		// TODO: handle other type
		if newObj.Spec.Type == api_core_v1.ServiceTypeNodePort {
			nd.service[name] = newObj
		}
	}
}

func (nd *nodeHandle) extractNodeCache(flag cache_v1.CacheOptionFlag, obj interface{}) {
	if obj == nil {
		return
	}
	node := obj.(*api_core_v1.Node)

	for _, addr := range node.Status.Addresses {
		// TODO: Type == api_core_v1.NodeExternalIP ???
		if addr.Type != api_core_v1.NodeInternalIP {
			continue
		}

		nd.isChange = true
		nd.address[addr.Address] |= flag
		if nd.address[addr.Address] == cache_v1.CacheFlagAll {
			nd.address[addr.Address] = 0
		}
	}
}

func (nd *nodeHandle) batchProcess() {
	if !nd.isChange {
		return
	}

	for name, svc := range nd.service {
		svCache := make(cache_v1.ServiceCache)
		svcNameID := hashName.StrToNum(name)

		// get service cache
		svcKey, ok := serviceKeyCaches[svcNameID]
		if !ok {
			serviceKeyCaches[svcNameID] = make(map[map_api_v1.ServiceKey]cache_v1.CacheOptionFlag)
			svcKey = serviceKeyCaches[svcNameID]
		}

		// get endpoint cache
		epCache, ok := endpointCaches[svcNameID]
		if !ok {
			endpointCaches[svcNameID] = make(cache_v1.EndpointCache)
			epCache = endpointCaches[svcNameID]
		}

		extractServiceCache(svcKey, cache_v1.CacheFlagNone, svcNameID, svc, nd.address)
		updateServiceEndpointNum(svCache, uint32(len(epCache)), svcKey, svcNameID)

		svCache.StatusFlush(cache_v1.CacheFlagUpdate)
		svCache.StatusFlush(cache_v1.CacheFlagDelete)
		svCache = nil
	}

	for ip, flag := range nd.address {
		if flag == cache_v1.CacheFlagDelete {
			delete(nd.address, ip)
		} else {
			nd.address[ip] = cache_v1.CacheFlagNone
		}
	}

	nd.isChange = false
}

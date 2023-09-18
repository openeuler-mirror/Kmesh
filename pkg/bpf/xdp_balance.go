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
 * Create: 2023-06-07
 */

package bpf

import (
	"fmt"
	"os"
	"reflect"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"openeuler.io/mesh/bpf/slb/bpf2go"
)

type BpfXdpBalance struct {
	Info BpfInfo
	Link link.Link
	bpf2go.XdpLoadBalanceObjects
}

func NewXdpBalance(cfg *Config) (BpfXdpBalance, error) {
	xdpBalance := BpfXdpBalance{}
	xdpBalance.Info.Config = *cfg

	xdpBalance.Info.MapPath = SharedMapPath
	if err := os.MkdirAll(xdpBalance.Info.MapPath, 0750); err != nil && !os.IsExist(err) {
		return xdpBalance, err
	}

	xdpBalance.Info.BpfFsPath += "/" + pinName + "/xdp/"
	if err := os.MkdirAll(xdpBalance.Info.BpfFsPath, 0750); err != nil && !os.IsExist(err) {
		return xdpBalance, err
	}

	return xdpBalance, nil
}

func (xdpB *BpfXdpBalance) LoadXdpBalanceObjects() (*ebpf.CollectionSpec, error) {
	var (
		err  error
		spec *ebpf.CollectionSpec
		opts ebpf.CollectionOptions
	)
	opts.Maps.PinPath = xdpB.Info.MapPath

	if spec, err = bpf2go.LoadXdpLoadBalance(); err != nil {
		return nil, err
	}

	if err = spec.LoadAndAssign(&xdpB.XdpLoadBalanceObjects, &opts); err != nil {
		return nil, fmt.Errorf("LoadAndAssign return err %s", err)
	}

	value := reflect.ValueOf(xdpB.XdpLoadBalanceObjects.XdpLoadBalancePrograms)
	if err = pinPrograms(&value, xdpB.Info.BpfFsPath); err != nil {
		return nil, err
	}

	return spec, nil
}

func (xdpB *BpfXdpBalance) Load() error {
	_, err := xdpB.LoadXdpBalanceObjects()
	if err != nil {
		return err
	}

	return nil
}

func (xdpB *BpfXdpBalance) Attach() error {
	xdpOpts := link.XDPOptions{
		Program:   xdpB.XdpLoadBalanceObjects.XdpLoadBalance,
		Interface: config.xdpLinkDev,
		Flags:     link.XDPDriverMode,
	}

	lk, err := link.AttachXDP(xdpOpts)
	if err != nil {
		return err
	}
	xdpB.Link = lk

	return nil
}

func (xdpB *BpfXdpBalance) close() error {
	if err := xdpB.XdpLoadBalanceObjects.Close(); err != nil {
		return err
	}

	return nil
}

func (xdpB *BpfXdpBalance) Detach() error {
	var value reflect.Value

	if err := xdpB.close(); err != nil {
		return err
	}

	value = reflect.ValueOf(xdpB.XdpLoadBalanceObjects.XdpLoadBalancePrograms)
	if err := unpinPrograms(&value); err != nil {
		return err
	}
	value = reflect.ValueOf(xdpB.XdpLoadBalanceObjects.XdpLoadBalanceMaps)
	if err := unpinMaps(&value); err != nil {
		return err
	}

	if err := os.RemoveAll(xdpB.Info.BpfFsPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	if xdpB.Link != nil {
		return xdpB.Link.Close()
	}
	return nil
}

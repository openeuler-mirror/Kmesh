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
 * Create: 2023-04-07
 */

package bpf

import (
	"os"
	"reflect"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"openeuler.io/mesh/bpf/slb/bpf2go"
)

const pinName = "bpf_slb"

type BpfCgroupSock struct {
	Info     BpfInfo
	LinkList []link.Link
	bpf2go.CgroupSockObjects
}

func NewCgroupSock(cfg *Config) (BpfCgroupSock, error) {
	b := BpfCgroupSock{}
	b.Info.Config = *cfg
	b.Info.MapPath = SharedMapPath
	if err := os.MkdirAll(b.Info.MapPath, 0750); err != nil && !os.IsExist(err) {
		return b, err
	}

	b.Info.BpfFsPath += "/" + pinName + "/socket/"
	if err := os.MkdirAll(b.Info.BpfFsPath, 0750); err != nil && !os.IsExist(err) {
		return b, err
	}

	return b, nil
}

func (b *BpfCgroupSock) loadCgroupSockObjects() (*ebpf.CollectionSpec, error) {
	var (
		err  error
		spec *ebpf.CollectionSpec
		opts ebpf.CollectionOptions
	)
	opts.Maps.PinPath = b.Info.MapPath

	if spec, err = bpf2go.LoadCgroupSock(); err != nil {
		return nil, err
	}

	setMapPinType(spec, ebpf.PinByName)
	if err = spec.LoadAndAssign(&b.CgroupSockObjects, &opts); err != nil {
		return nil, err
	}

	value := reflect.ValueOf(b.CgroupSockObjects.CgroupSockPrograms)
	if err = pinPrograms(&value, b.Info.BpfFsPath); err != nil {
		return nil, err
	}

	return spec, nil
}

func (b *BpfCgroupSock) Load() error {
	spec, err := b.loadCgroupSockObjects()
	if err != nil {
		return err
	}

	prog := spec.Programs["sock_connect4"]
	b.Info.Type = prog.Type
	b.Info.AttachType = prog.AttachType

	return nil
}

func (b *BpfCgroupSock) Attach() error {
	var err error
	cgroupOpt := link.CgroupOptions{
		Path:    b.Info.Cgroup2Path,
		Attach:  b.Info.AttachType,
		Program: b.CgroupSockObjects.SockConnect4,
	}

	lk, err := link.AttachCgroup(cgroupOpt)
	if err != nil {
		return err
	}
	b.LinkList = append(b.LinkList, lk)

	return nil
}

func (b *BpfCgroupSock) close() error {
	if err := b.CgroupSockObjects.Close(); err != nil {
		return err
	}

	return nil
}

func (b *BpfCgroupSock) Detach() error {
	var value reflect.Value

	if err := b.close(); err != nil {
		return err
	}

	value = reflect.ValueOf(b.CgroupSockObjects.CgroupSockPrograms)
	if err := unpinPrograms(&value); err != nil {
		return err
	}
	value = reflect.ValueOf(b.CgroupSockObjects.CgroupSockMaps)
	if err := unpinMaps(&value); err != nil {
		return err
	}

	if err := os.RemoveAll(b.Info.BpfFsPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	if b.LinkList != nil {
		for _, l := range b.LinkList {
			if err := l.Close(); err != nil {
				return err
			}
		}
		b.LinkList = nil
	}
	return nil
}

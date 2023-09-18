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
 * Create: 2023-04-09
 */

package bpf

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/rlimit"
)

type BpfInfo struct {
	Config
	MapPath    string
	Type       ebpf.ProgramType
	AttachType ebpf.AttachType
}

type BpfObject struct {
	CgroupSock BpfCgroupSock
	XdpBalance BpfXdpBalance
}

var Obj BpfObject

func StartCgroupSock() error {
	var err error

	if Obj.CgroupSock, err = NewCgroupSock(&config); err != nil {
		return err
	}

	if err = Obj.CgroupSock.Load(); err != nil {
		Stop()
		return fmt.Errorf("bpf Load failed, %s", err)
	}

	if err = Obj.CgroupSock.Attach(); err != nil {
		Stop()
		return fmt.Errorf("bpf Attach failed, %s", err)
	}

	return nil
}

func StartXdpBalance() error {
	var err error
	if Obj.XdpBalance, err = NewXdpBalance(&config); err != nil {
		return err
	}

	if err = Obj.XdpBalance.Load(); err != nil {
		return fmt.Errorf("bpf xdp_banlance Load failed, %s", err)
	}

	if err = Obj.XdpBalance.Attach(); err != nil {
		return fmt.Errorf("bpf xdp_banlance Attach failed, %s", err)
	}
	return nil
}

func Start() error {
	var err error

	if err = rlimit.RemoveMemlock(); err != nil {
		return err
	}

	if err = StartCgroupSock(); err != nil {
		return err
	}
	if err = StartXdpBalance(); err != nil {
		Stop()
		return fmt.Errorf("bpf StartXdpBalance failed, %s", err)
	}

	return nil
}

func Stop() error {
	if err := Obj.XdpBalance.Detach(); err != nil {
		return fmt.Errorf("failed to detach XdpBalance, err:%s", err)
	}

	if err := Obj.CgroupSock.Detach(); err != nil {
		return fmt.Errorf("failed to detach cgroup, err:%s", err)
	}

	return nil
}

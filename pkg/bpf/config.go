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
 * Create: 2023-04-08
 */

package bpf

import (
	"flag"
	"openeuler.io/mesh/pkg/options"
	"os"
	"path/filepath"
)

var config Config

const (
	SharedMapPath = "/sys/fs/bpf/tc/globals"
)

func init() {
	options.Register(&config)
}

type Config struct {
	BpfFsPath      string `json:"-bpf-fs-path"`
	Cgroup2Path    string `json:"-cgroup2-path"`
	xdpLinkDev     int    `json:"xdp-link-dev"`
}

func (c *Config) SetArgs() error {
	flag.StringVar(&c.BpfFsPath, "bpf-fs-path", "/sys/fs/bpf", "bpf fs path")
	flag.StringVar(&c.Cgroup2Path, "cgroup2-path", "/mnt/cgroup2", "cgroup2 path")

	flag.IntVar(&c.xdpLinkDev, "xdp-link-dev", 2, "dev number for xdp program to link")

	return nil
}

func (c *Config) ParseConfig() error {
	var err error

	if c.Cgroup2Path, err = filepath.Abs(c.Cgroup2Path); err != nil {
		return err
	}
	if _, err = os.Stat(c.Cgroup2Path); err != nil {
		return err
	}

	if c.BpfFsPath, err = filepath.Abs(c.BpfFsPath); err != nil {
		return err
	}
	if _, err = os.Stat(c.BpfFsPath); err != nil {
		return err
	}

	return nil
}

func GetConfig() *Config {
	return &config
}

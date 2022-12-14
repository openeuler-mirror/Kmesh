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
 * Create: 2021-10-09
 */

package controller

import (
	"openeuler.io/mesh/pkg/bpf"
	"openeuler.io/mesh/pkg/controller/envoy"
	"openeuler.io/mesh/pkg/controller/interfaces"
)

var (
	stopCh     = make(chan struct{})
	client     interfaces.ClientFactory
)

func Start() error {
	var err error

	client, err = config.NewClient()
	if err != nil {
		return err
	}

	return client.Run(stopCh)
}

func Stop() {
	var obj struct{}
	stopCh <- obj
	close(stopCh)
	client.Close()
}

func GetAdsClient() *envoy.AdsClient {
	if !bpf.GetConfig().EnableKmesh {
		return nil
	}
	return client.(*envoy.AdsClient)
}

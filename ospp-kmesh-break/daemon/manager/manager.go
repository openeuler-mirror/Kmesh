/*
 * Copyright (c) 2019 Huawei Technologies Co., Ltd.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

 * Author: LemmyHuang
 * Create: 2021-10-09
 */

// Package manager: kmesh daemon manager
package manager

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf"

	"openeuler.io/mesh/cmd/command"
	"openeuler.io/mesh/pkg/bpf"
	"openeuler.io/mesh/pkg/controller"
	"openeuler.io/mesh/pkg/logger"
	"openeuler.io/mesh/pkg/options"
	"openeuler.io/mesh/pkg/pid"
	"openeuler.io/mesh/api/v2/core"
)

const (
	pkgSubsys = "manager"
)

var (
	log = logger.NewLoggerField(pkgSubsys)
)

// 初始化 ebpf map
var ebpfMap1 *ebpf.Map
var ebpfMap2 *ebpf.Map
var ticker *time.Ticker

func initMaps() {
	var err error
	options := &ebpf.LoadPinOptions{
	}
	ebpfMap1, err = ebpf.LoadPinnedMap("map_of_break", options)
	if err != nil {
		log.Fatal("Failed to load map_of_break:", err)
	}
	
	ebpfMap2, err = ebpf.LoadPinnedMap("map_of_break_count", options)
	if err != nil {
		log.Fatal("Failed to load map_of_break_count:", err)
	}
}

// 获取限额数
func getLimit() (int, error) {
	var limit int
	var str string
	str = "break_config"
	
	err := ebpfMap1.Lookup(&str, &limit)
	if err != nil {
		return 0, err
	}
	return limit, nil
}

// 重置计数
func resetCount() error {
    var nextKey, currentKey SocketAddress
    var zeroValue = 0

    for {
        err := ebpfMap2.NextKey(&currentKey, &nextKey)
        if err != nil {
            // No more keys to iterate
            break
        }

        // Set the value for this key to zero
        if err := ebpfMap2.Put(&nextKey, &zeroValue); err != nil {
            return fmt.Errorf("failed to reset count for key %v: %w", nextKey, err)
        }

        currentKey = nextKey
    }

    return nil
}

// 检查 ebpf map 中键值对是否存在
func exists() (bool, error) {
	var someKey = SocketAddress{
		Protocol: 0,
		Port: 22,
		Ipv4: 127,
	}
	err := ebpfMap2.Lookup(&someKey, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func handleTicker() bool {
	limit, err := getLimit()
	if err != nil {
		log.Error("获取限额数失败：", err)
		return false
	}
	done := make(chan struct{})
	ticker := time.NewTicker(time.Duration(limit) * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				processTicker()
			}
			case <-done:
            	return
        	}
		}
	}()
	return true
}

func processTicker() {
	if exists, err := exists(); err != nil {
		log.Error("检查ebpfMap2存在性失败：", err)
	} else if !exists {
		createEbpfMap2Entry()
	}
	resetMap2Count()
}

func createEbpfMap2Entry() {
	var someKey = SocketAddress{
		Protocol: 0,
		Port:     22,
		Ipv4:    127,
	}
	initialValue := 0
	if err := ebpfMap2.Put(&someKey, &initialValue); err != nil {
		log.Error("创建ebpfMap2失败：", err)
	}
}

func resetMap2Count() {
	if err := resetCount(); err != nil {
		log.Error("重置ebpfMap2的count变量失败：", err)
	}
}

// Execute start daemon manager process
func Execute() {
	var err error

	if err = options.InitDaemonConfig(); err != nil {
		log.Error(err)
		return
	}
	log.Info("options InitDaemonConfig successful")

	if err = pid.CreatePidFile(); err != nil {
		log.Errorf("failed to start, reason: %v", err)
		return
	}
	defer pid.RemovePidFile()

	if err = bpf.Start(); err != nil {
		fmt.Println(err)
		return
	}
	log.Info("bpf Start successful")

	if err = controller.Start(); err != nil {
		log.Error(err)
		bpf.Stop()
		return
	}
	log.Info("controller Start successful")
	initMaps()
	if !handleTicker() {
		return
	}
	ticker = time.NewTicker(time.Duration(limit) * time.Second)
	if err = command.StartServer(); err != nil {
		log.Error(err)
		controller.Stop()
		bpf.Stop()
		return
	}
	log.Info("command StartServer successful")
	setupCloseHandler()
	return
}

func setupCloseHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGABRT, syscall.SIGTSTP)

	<-ch
	command.StopServer()
	controller.Stop()
	bpf.Stop()
	// 定时器清理
	ticker.Stop()
	close(done)
	fmt.Println("定时器已经清理")
	log.Warn("signal Notify exit")
}

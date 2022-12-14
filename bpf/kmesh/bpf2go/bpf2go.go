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
 * Author: nlgwcy
 * Create: 2022-02-27
 */

// Package bpf2go  generate c to go struct
package bpf2go

// go run github.com/cilium/ebpf/cmd/bpf2go --help
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang  --cflags $EXTRA_CFLAGS --cflags $EXTRA_CDEFINE KmeshCgroupSock ../cgroup_sock.c -- -I../include -I../../include -I../../../api/v2-c
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang  --cflags $EXTRA_CFLAGS --cflags $EXTRA_CDEFINE KmeshSockops ../sockops.c -- -I../include -I../../include -I../../../api/v2-c
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang  --cflags $EXTRA_CFLAGS --cflags $EXTRA_CDEFINE KmeshFilter ../filter.c -- -I../include -I../../include -I../../../api/v2-c
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang  --cflags $EXTRA_CFLAGS --cflags $EXTRA_CDEFINE KmeshRouteConfig ../route_config.c -- -I../include -I../../include -I../../../api/v2-c
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang  --cflags $EXTRA_CFLAGS --cflags $EXTRA_CDEFINE KmeshCluster ../cluster.c -- -I../include -I../../include -I../../../api/v2-c

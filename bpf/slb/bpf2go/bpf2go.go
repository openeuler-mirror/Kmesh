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
 * Create: 2023-05-10
 */

package bpf2go

// go run github.com/cilium/ebpf/cmd/bpf2go --help
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang --cflags $EXTRA_CFLAGS --cflags $EXTRA_CDEFINE CgroupSock ../cgroup_sock.c -- -I../include -I../../../include -DCGROUP_SOCK_CTX
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang --cflags $EXTRA_CFLAGS --cflags $EXTRA_CDEFINE XdpLoadBalance ../xdp_load_balance.c -- -I../include -I../../../include -DXDP_CTX
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang --cflags $EXTRA_CFLAGS --cflags $EXTRA_CDEFINE RevTC ../tc.c -- -I../include -I../../../include  -DTC_CTX

/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
 * MeshAccelerating is licensed under the Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Author: superCharge
 * Create: 2023-04-25
 */

#ifndef _COMMON_H_
#define _COMMON_H_

#define bpf_unused __attribute__((__unused__))

#define BPF_MAX(x, y)		(((x) > (y)) ? (x) : (y))
#define BPF_MIN(x, y)		(((x) < (y)) ? (x) : (y))

#ifndef bpf_memset
#define bpf_memset(dest, chr, n)   __builtin_memset((dest), (chr), (n))
#endif

#ifndef bpf_memcpy
#define bpf_memcpy(dest, src, n)   __builtin_memcpy((dest), (src), (n))
#endif

#ifndef __stringify
#define __stringify(X)					#X
#endif
#define SEC_TAIL(ID, KEY)				SEC(__stringify(ID) "/" __stringify(KEY))

// bpf return value
#define CGROUP_SOCK_ERR		0
#define CGROUP_SOCK_OK		1

// bpf tail call type
#define SLB_SOCKET_CALLS	(cgroup/connect4)
#define SLB_XDP_CALLS

// loadbalance type
enum LB_POLICY_TYPE{
	LB_POLICY_ROUND_ROBIN = 0,
	LB_POLICY_RANDOM,
};

#endif // _COMMON_H_

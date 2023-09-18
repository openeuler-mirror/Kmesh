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
 * Author: bitcoffee
 * Create: 2023-05-22
 */
#ifndef _CONNTRACK_H_
#define _CONNTRACK_H_
enum NATTYPE {
	NAT_TYPE_DNAT = 0,
	NAT_TYPE_SNAT,
	NAT_TYPE_FULL_NAT,
};

struct mac_info_t {
	/* fullnat: MAC address of the local egress network adapter after fib_lookup
	 * rev-fullnat: MAC address of the ct map recorded origin dmac
	 */
	__u8  nat_smac[6]; 
	/* fullnat: MAC address of the next hop after fib_lookup
	 * rev-fullnat: MAC address of the ct map recorded origin smac
	 */
	__u8  nat_dmac[6];
	/* fullnat: egress if index after fib_lookup
	 * rev-fullnat: origin ingress index
	 */
	__u32 nat_ifindex;
} __attribute__((packed));

struct nat_info_t {
	__u8 protocol;
	union {
		__u32 v4_saddr;
		// todo ipv6
		// struct in6_addr v6_saddr;
	} nat_saddr;
	union {
		__u32 v4_daddr;
		// todo ipv6
		// struct in6_addr v6_daddr;
	} nat_daddr;
	__u32 nat_sport;
	__u32 nat_dport;
	enum NATTYPE nat_type;
	struct mac_info_t nat_mac_info;
} __attribute__((packed));

struct ct_value_t {
	__u64 create_time;
	__u64 last_use;
	struct nat_info_t nat_info;
	// record fin package seq, it is 0 if not recv a fin package
	__u32 fin_seq;
} __attribute__((packed));

#endif /* _CONNTRACK_H_ */

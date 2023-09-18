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

package nets

import (
	"encoding/binary"
	"net"
)

func ConvertIpToUint32(ip string) uint32 {
	netIP := net.ParseIP(ip)
	if len(netIP) == net.IPv6len {
		return binary.LittleEndian.Uint32(netIP.To4())
	}
	return binary.LittleEndian.Uint32(netIP)
}

func ConvertUint32ToIp(num uint32) string {
	netIP := make(net.IP, 4)
	binary.LittleEndian.PutUint32(netIP, num)
	return netIP.String()
}

func ConvertPortToLittleEndian(num uint32) uint32 {
	// FIXME
	tmp := make([]byte, 2)
	big16 := uint16(num)
	binary.BigEndian.PutUint16(tmp, big16)
	little16 := binary.LittleEndian.Uint16(tmp)
	return uint32(little16)
}

// ConvertPortToBigEndian convert uint32 to network order
func ConvertPortToBigEndian(little uint32) uint32 {
	// first convert to uint16, then convert the byte order,
	// and finally switch back to uint32
	tmp := make([]byte, 2)
	little16 := uint16(little)
	binary.BigEndian.PutUint16(tmp, little16)
	big16 := binary.LittleEndian.Uint16(tmp)
	return uint32(big16)
}

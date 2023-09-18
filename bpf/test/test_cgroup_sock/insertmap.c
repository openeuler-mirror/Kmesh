/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
 * MeshAccelerating is licensed under the Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *	 http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Author: Bitcoffee
 * Create: 2023-07-01
 */

#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>
#include <bpf/libbpf.h>
#include <bpf/bpf.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <arpa/inet.h>
// need include include/map_data_v1
#include "service.h"
#include "backend.h"
#include "endpoint.h"

#define SERVICE_MAP_NAME "slb_service"
#define BACKEND_MAP_NAME "slb_backend"
#define ENDPOINT_MAP_NAME "slb_endpoint"

void insert_map_service(int obj_fd)
{
    struct service_key_t service_key = {0};
    struct service_entry_t service_entry = {0};
    service_key.protocol = 6;
    service_key.port = htons(30000);
    char *ip = "127.0.0.1";
    struct in_addr dst;
    inet_pton(AF_INET, ip, (void *)&dst);
    service_key.ipv4 = dst.s_addr;
    service_entry.service_id = 1;
    service_entry.count = 2;
    service_entry.policy = 0;
    bpf_map_update_elem(obj_fd, &service_key, &service_entry, 0);
}

void insert_map_backend(int obj_fd)
{
    struct backend_key_t backend_key = {0};
    struct backend_entry_t backend_entry = {0};
    backend_key.backend_slot = 0;
    backend_key.service_id = 1;
    backend_entry.endpoint_id = 1;
    bpf_map_update_elem(obj_fd, &backend_key, &backend_entry, 0);

    backend_key.backend_slot = 1;
    backend_entry.endpoint_id = 2;
    bpf_map_update_elem(obj_fd, &backend_key, &backend_entry, 0);
}

void insert_map_endpoint(int obj_fd)
{
    endpoint_key_t endpoint_key;
    struct endpoint_entry_t endpoint_entry = {0};
    endpoint_key = 1;
    endpoint_entry.protocol = htonl(6);
    endpoint_entry.port = htons(30010);
    char *ip = "127.0.0.1";
    struct in_addr dst;
    inet_pton(AF_INET, ip, (void *)&dst);
    endpoint_entry.ipv4 = dst.s_addr;
    bpf_map_update_elem(obj_fd, &endpoint_key, &endpoint_entry, 0);

    endpoint_key = 2;
    endpoint_entry.port = htons(30020);
    bpf_map_update_elem(obj_fd, &endpoint_key, &endpoint_entry, 0);
}

int main(int argc, char *argv[])
{
    int obj_id = 0;
    int obj_fd;
    struct bpf_map_info map_info = {0};
    int info_length = sizeof(struct bpf_map_info);
    while(!bpf_map_get_next_id(obj_id, &obj_id)) {
        obj_fd = bpf_map_get_fd_by_id(obj_id);
        bpf_obj_get_info_by_fd(obj_fd, &map_info, &info_length);
        if (strcmp(map_info.name, SERVICE_MAP_NAME) == 0) {
            insert_map_service(obj_fd);
        } else if (strcmp(map_info.name, BACKEND_MAP_NAME) == 0) {
            insert_map_backend(obj_fd);
        } else if (strcmp(map_info.name, ENDPOINT_MAP_NAME) == 0) {
            insert_map_endpoint(obj_fd);
        }
    }
    return 0;
}

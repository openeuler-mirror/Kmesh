# Copyright (c) 2019 Huawei Technologies Co., Ltd.
# MeshAccelerating is licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Author: superCharge
# Create: 2023-06-08

ROOT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

include ./mk/bpf.vars.mk
include ./mk/bpf.print.mk

# compiler flags
GOFLAGS := $(EXTRA_GOFLAGS)

# target
APPS1 := kmesh-daemon

.PHONY: all install clean

all:
#$(QUIET) $(ROOT_DIR)/mk/pkg-config.sh set
	
	$(QUIET) $(GO) mod download
	$(QUIET) $(GO) generate bpf/slb/bpf2go/bpf2go.go
	
	$(call printlog, BUILD, $(APPS1))
	$(QUIET) $(GO) build -o $(APPS1) $(GOFLAGS) ./daemon/main.go
	
#$(QUIET) $(ROOT_DIR)/mk/pkg-config.sh unset

install:
#$(QUIET) make install -C api/v2-c
#$(QUIET) make install -C bpf/deserialization_to_bpf_map

	$(call printlog, INSTALL, $(INSTALL_BIN)/$(APPS1))
##$(QUIET) install -dp -m 0750 $(INSTALL_BIN)
	$(QUIET) install -Dp -m 0550 $(APPS1) $(INSTALL_BIN)
	
clean:
	$(call printlog, CLEAN, $(APPS1))
	$(QUIET) rm -rf $(APPS1) $(APPS1)
	

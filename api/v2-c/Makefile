# Copyright (c) 2019 Huawei Technologies Co., Ltd.
# MeshAccelerating is licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Author: LemmyHuang
# Create: 2021-09-17

ROOT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

include ../../mk/bpf.vars.mk
include ../../mk/bpf.print.mk

INCLUDES = -I./

# compiler flags
LDFLAGS := -lprotobuf-c
CFLAGS := $(EXTRA_CFLAGS) $(EXTRA_CDEFINE)
CFLAGS += -fstack-protector -fPIC
CFLAGS += -Wall -Werror

SOURCES = $(wildcard */*.c)
OBJECTS = $(subst .c,.o,$(SOURCES))
# target
APPS := libkmesh_api_v2_c.so

.PHONY: all install clean

all: $(APPS)

$(APPS): $(OBJECTS)
	$(call printlog, BUILD, api/v2-c/$@)
	$(QUIET) $(CLANG) $(CFLAGS) $(LDFLAGS) -shared $^ -o $@

%.o: %.c
	$(call printlog, BUILD, api/v2-c/$@)
	$(QUIET) $(CLANG) $(CFLAGS) $(INCLUDES) -c $^ -o $@

install:
	$(call printlog, INSTALL, $(INSTALL_LIB)/$(APPS))
	$(QUIET) install -Dp -m 0550 $(APPS) $(INSTALL_LIB)

uninstall:
	$(call printlog, UNINSTALL, $(INSTALL_LIB)/$(APPS))
	$(QUIET) rm -rf $(INSTALL_LIB)/$(APPS)

clean:
	$(call printlog, CLEAN, $(APPS))
	$(QUIET) rm -rf $(APPS) $(APPS) $(OBJECTS)


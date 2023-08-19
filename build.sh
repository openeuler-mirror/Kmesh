#!/bin/bash
ROOT_DIR=$(dirname $(readlink -f ${BASH_SOURCE[0]}))

function prepare() {
    ARCH=$(arch)
    if [ "$ARCH" == "x86_64" ]; then
            export EXTRA_GOFLAGS="-gcflags=\"-N -l\""
            export EXTRA_CFLAGS="-O0 -g"
            export EXTRA_CDEFINE="-D__x86_64__"
    fi

    export PATH=$PATH:$ROOT_DIR/vendor/google.golang.org/protobuf/cmd/protoc-gen-go/
    cp $ROOT_DIR/depends/include/5.10.0-60.18.0.50.oe2203/bpf_helper_defs_ext.h $ROOT_DIR/bpf/include/
    cp /sys/kernel/btf/vmlinux /usr/lib/modules/`uname -r`/build/
}

function install() {
    mkdir -p /etc/kmesh
    chmod 750 /etc/kmesh
    cp $ROOT_DIR/config/kmesh.json /etc/kmesh
    chmod 640 /etc/kmesh/kmesh.json

    cp $ROOT_DIR/build/kmesh-start-pre.sh /usr/bin
    chmod 550 /usr/bin/kmesh-start-pre.sh
    cp $ROOT_DIR/build/kmesh-stop-post.sh /usr/bin
    chmod 550 /usr/bin/kmesh-stop-post.sh

    mkdir -p /etc/oncn-mda
    chmod 700 /etc/oncn-mda
    cp $ROOT_DIR/oncn-mda/etc/oncn-mda.conf /etc/oncn-mda/
    chmod 600 /etc/oncn-mda/oncn-mda.conf

    cp $ROOT_DIR/build/kmesh.service /usr/lib/systemd/system/
    systemctl daemon-reload
}

function uninstall() {
    rm -rf /etc/kmesh
    rm -rf /usr/bin/kmesh-start-pre.sh
    rm -rf /usr/bin/kmesh-stop-post.sh
    rm -rf /etc/oncn-mda
    rm -rf /usr/share/oncn-mda
    rm -rf /usr/lib/systemd/system/kmesh.service
    systemctl daemon-reload
}

function clean() {
    rm -rf /etc/kmesh
    rm -rf /usr/bin/kmesh-start-pre.sh
    rm -rf /usr/bin/kmesh-stop-post.sh
}

if [ "$1" == "-h"  -o  "$1" == "--help" ]; then
    echo build.sh -h/--help : Help.
    echo build.sh -b/--build: Build Kmesh.
    echo build.sh -i/--install: Install Kmesh.
    echo build.sh -c/--clean: Clean the built binary.
    echo build.sh -u/--uninstall: Uninstall Kmesh.
    exit
fi

if [ -z "$1" -o "$1" == "-b"  -o  "$1" == "--build" ]; then
    prepare
    make
    exit
fi

if [ "$1" == "-i"  -o  "$1" == "--install" ]; then
    make install
    install
    exit
fi

if [ "$1" == "-u"  -o  "$1" == "--uninstall" ]; then
    make uninstall
    uninstall
    exit
fi

if [ "$1" == "-c"  -o  "$1" == "--clean" ]; then
    make clean
    clean
    exit
fi

#
#Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
#MeshAccelerating is licensed under the Mulan PSL v2.
#You can use this software according to the terms and conditions of the Mulan PSL v2.
#You may obtain a copy of Mulan PSL v2 at:
#	 http://license.coscl.org.cn/MulanPSL2
#THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
#IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
#PURPOSE.
#See the Mulan PSL v2 for more details.
#Author: bitcoffee
#Create: 2023-07-01

# test cgroup sock
# step 1: init environment
tmpdir=`mktemp -d`

mount none -t cgroup2 $tmpdir
attached_prog_lists=`bpftool cgroup tree | grep connect4 | awk '{print $1}'`
for prog in $attached_prog_lists; do
    bpftool cgroup detach $tmpdir connect4 id $prog > /dev/null
done
rm -rf /sys/fs/bpf/*
gcc insertmap.c -I ../../../include/map_data_v1/ -lbpf -o insert

clang -g -O2 -target bpf -c ../../slb/cgroup_sock.c -I../../slb/include -I../../../include -D__x86_64__ -DCGROUP_SOCK_CTX -o cgroup_sock.o

# step 2: load/attach ebpf program, update service / backend / endpoint map
bpftool prog load cgroup_sock.o /sys/fs/bpf/test
bpftool cgroup attach $tmpdir connect4 pinned /sys/fs/bpf/test

./insert
# step 3: start nc server
nc -v -4 -l 30010 < ./http_reply &
nc -v -4 -l 30020 < ./http_reply2 &
# step 4: connect to nc server and get reply
number1=`curl -v 127.0.0.1:30000 2>&1 | grep "Slb-test1" | wc -l`
number2=`curl -v 127.0.0.1:30000 2>&1 | grep "Slb-test2" | wc -l`
if [ $number1 -eq "1" ] && [ $number2 -eq "1" ]; then
    echo "test SUCCESS"
else
    echo "test FAILED"
fi

# step 5: clean temp dir
bpftool cgroup detach $tmpdir connect4 pinned /sys/fs/bpf/test
killall nc
rm -rf /sys/fs/bpf/test
umount $tmpdir
rm -rf $tmpdir
cgroup:
	clang -g -O2 -target bpf -c cgroup_sock.c -Iinclude -I../../include -D__x86_64__ -DCGROUP_SOCK_CTX

xdp:
	clang -g -O2 -target bpf -c xdp_load_balance.c -Iinclude -I../include -I../../include -D__x86_64__ -DXDP_CTX

tc:
	clang -g -O2 -target bpf -c tc.c -Iinclude -I../include -I../../include -D__x86_64__ -DTC_CTX

all:tc xdp cgroup

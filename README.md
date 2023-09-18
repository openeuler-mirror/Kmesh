# MeshAccelerating

## Usage Tutorial

build

```sh
make clean
make
make install
```

config

```sh
mkdir /mnt/cgroup2
mount -t cgroup2 none /mnt/cgroup2/

kubectl exec xxx-pod-name -c istio-proxy -- cat etc/istio/proxy/envoy-rev0.json > envoy-rev0.json
---
       , {
        "name": "xds-grpc",
        "type" : "STATIC",
        "connect_timeout": "1s",
        "lb_policy": "ROUND_ROBIN",
        "load_assignment": {
          "cluster_name": "xds-grpc",
          "endpoints": [{
            "lb_endpoints": [{
              "endpoint": {
                "address":{
                  "socket_address": {
                    "protocol": "TCP",
                    "address": "192.168.123.249", # istiod pod IP
                    "port_value": 15010
                  }
                }
              }
            }]
          }]
        },
---
```

kmesh-daemon

```sh
# kmesh-daemon -h
Usage of kmesh-daemon:
  -bpf-fs-path string
    	bpf fs path (default "/sys/fs/bpf")
  -cgroup2-path string
    	cgroup2 path (default "/mnt/cgroup2")
  -config-file string
    	[if -enable-kmesh] deploy in kube cluster (default "/etc/istio/proxy/envoy-rev0.json")
  -enable-ads
    	[if -enable-kmesh] enable control-plane from ads (default true)
  -enable-in-cluster
    	[if -enable-slb] deploy in kube cluster by DaemonSet
  -enable-kmesh
    	enable bpf kmesh
  -enable-slb
    	enable bpf slb
  -service-cluster string
    	[if -enable-kmesh] TODO (default "TODO")
  -service-node string
    	[if -enable-kmesh] TODO (default "TODO")

# example
./kmesh-daemon -enable-slb=true
# example
./kmesh-daemon -enable-kmesh=true -enable-ads=true -config-file=envoy-rev0.json
./kmesh-daemon -enable-kmesh=true -enable-ads=false
```

kmesh-cmd

```sh
# kmesh-cmd -h
Usage of kmesh-cmd:
  -config-file string
    	input config-resources to bpf maps (default "./config-resources.json")

# example
./kmesh-cmd -config-file=examples/api-v2-config/config-resources.json
```

admin

```sh
# curl http://localhost:15200/help
	/help: print list of commands
	/options: print config options
	/bpf/slb/maps: print bpf slb maps in kernel
	/bpf/kmesh/maps: print bpf kmesh maps in kernel
	/controller/envoy: print control-plane in envoy cache
	/controller/kubernetes: print control-plane in kubernetes cache

# example
curl http://localhost:15200/bpf/kmesh/maps
curl http://localhost:15200/controller/envoy
```


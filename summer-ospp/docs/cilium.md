## Openeuler  cilium cni

version: 1.14.0

### bare metal cluster install cilium cni

```shell
CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)

CLI_ARCH=amd64

curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}

sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
sudo tar xzvfC cilium-linux-${CLI_ARCH}.tar.gz /usr/local/bin
rm cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}

cilium install --version 1.14.0
cilium status --wait
cilium connectivity test

./helm install cilium cilium/cilium --version 1.14.0 \
   --namespace kube-system \
   --set operator.replicas=1 \
   --set k8sServiceHost=192.168.43.152 \
   --set k8sServicePort=6443 
```

### 修改为 native route模式

./helm upgrade cilium cilium/cilium \
   --namespace kube-system \
   --reuse-values \
   --set tunnel=disabled \
   --set autoDirectNodeRoutes=true \
   --set ipv4NativeRoutingCIDR=10.0.0.0/16

执行`kubectl -n kube-system edit configmap cilium-config -o yaml`

改为`routing-mode: native`



### 基于kind cluster开发调试Cilium

vscode

openeuler : 22.03 LTS

go version : 1.20

kind version : v0.20.0

docker version ：24.0.5

docker-buildx : v0.11.2

cgroup v2

```shell
# 修改 Euler OS 系统使用 cgroupv2
stat -fc %T /sys/fs/cgroup/

grubby --update-kernel=ALL --args=systemd.unified_cgroup_hierarchy=1
reboot
```

```shell
cd cilium
make kind IMAGE=kindest/node:v1.20.2
make kind-image 

make kind: Creates a kind cluster based on the configuration passed in. For more information, see Configuration for clusters.

make kind-image: Builds all Cilium images and loads them into the cluster.

make kind-image-agent: Builds the Cilium Agent image only and loads it into the cluster.

make kind-image-operator: Builds the Cilium Operator (generic) image only and loads it into the cluster.

make kind-image-debug: Builds all Cilium images with optimizations disabled and dlv embedded for live debugging enabled and loads the images into the cluster.

make kind-install-cilium: Installs Cilium into the cluster using the Cilium CLI.

make kind-down: Tears down and deletes the cluster.
```

```
# 安转到kind cluster中
cilium install --chart-directory=/root/cilium/install/kubernetes/cilium --helm-values=/root/cilium/contrib/testing/kind-values.yaml --version= >/dev/null 2>&1 &
```

### 踩坑

#### 安装24.0.5版本的docker-ce, 安装openeuler的默认docker版本构建镜像cilium不过

```shell
 yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
 cd /etc/yum.repos.d/
 sed -i 's/\$releasever/7/g' docker-ce.repo
 yum list docker-ce --showduplicates|sort -r
 wget -O /etc/yum.repos.d/CentOS-Base.repo https://repo.huaweicloud.com/repository/conf/CentOS-7-reg.repo
 sed -i 's/\$releasever/7/g' CentOS-Base.repo
 yum clean all
 yum makecache
 yum install docker-ce docker-ce-cli containerd.io
```

### cilium direct routing mode 部分源码解析

```
/* Value of endpoint map */
struct endpoint_info {
	__u32		ifindex;  
	__u16		unused; /* used to be sec_label, no longer used */
	__u16		lxc_id;   容器ID？
	__u32		flags;
	mac_t		mac;      容器端 veth mac 地址
	mac_t		node_mac; 主机端veth mac地址
	__u32		sec_id;
	__u32		pad[3];
};

```


```
/*
 * from-host is attached as a tc egress filter to the node's 'cilium_host'
 * interface if present.
 */
__section_entry
int cil_from_host(struct __ctx_buff *ctx)
{
	/* Traffic from the host ns going through cilium_host device must
	 * not be subject to EDT rate-limiting.
	 */
	edt_set_aggregate(ctx, 0);
	return handle_netdev(ctx, true);
}
```

```
/*
 * from-netdev is attached as a tc ingress filter to one or more physical devices
 * managed by Cilium (e.g., eth0). This program is only attached when:
 * - the host firewall is enabled, or
 * - BPF NodePort is enabled, or
 * - L2 announcements are enabled, or
 * - WireGuard's host-to-host encryption and BPF NodePort are enabled
 */
__section_entry
int cil_from_netdev(struct __ctx_buff *ctx)
{
	__u32 __maybe_unused src_id = 0;

#ifdef ENABLE_NODEPORT_ACCELERATION
	__u32 flags = ctx_get_xfer(ctx, XFER_FLAGS);
#endif
	int ret;

	/* Filter allowed vlan id's and pass them back to kernel.
	 * We will see the packet again in from-netdev@eth0.vlanXXX.
	 */
	if (ctx->vlan_present) {
		__u32 vlan_id = ctx->vlan_tci & 0xfff;

		if (vlan_id) {
			if (allow_vlan(ctx->ifindex, vlan_id))
				return CTX_ACT_OK;

			ret = DROP_VLAN_FILTERED;
			goto drop_err;
		}
	}

	ctx_skip_nodeport_clear(ctx);

#ifdef ENABLE_NODEPORT_ACCELERATION
	if (flags & XFER_PKT_NO_SVC)
		ctx_skip_nodeport_set(ctx);

#ifdef HAVE_ENCAP
	if (flags & XFER_PKT_SNAT_DONE)
		ctx_snat_done_set(ctx);
#endif
#endif

#ifdef ENABLE_HIGH_SCALE_IPCACHE
	ret = decapsulate_overlay(ctx, &src_id);
	if (IS_ERR(ret))
		return send_drop_notify_error(ctx, src_id, ret, CTX_ACT_DROP,
				       METRIC_INGRESS);
	if (ret == CTX_ACT_REDIRECT)
		return ret;
#endif /* ENABLE_HIGH_SCALE_IPCACHE */

	return handle_netdev(ctx, false);

drop_err:
	return send_drop_notify_error(ctx, 0, ret, CTX_ACT_DROP, METRIC_INGRESS);
}
```

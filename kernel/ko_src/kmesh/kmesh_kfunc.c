#include <linux/btf.h>
#include <linux/btf_ids.h>
#include <linux/module.h>
#include <linux/init.h>
#include <linux/string.h>
#include "kmesh_parse_protocol_data.h"
#include "kmesh_parse_http_1_1.h"
#include "kmesh_kfunc.h"

__diag_push();
__diag_ignore_all("-Wmissing-prototypes",
		"Global functions as their definitions will be in BTF");

__bpf_kfunc int
bpf__strnlen(const char *str, int len)
{
	return strnlen(str, len);
}	

__bpf_kfunc struct bpf_mem_ptr *
bpf__strnstr(void *dst, int dst__sz, const char *src, int len)
{
	struct bpf_mem_ptr *tmp = dst;
	char *res = strnstr(tmp->ptr, src, len);
	if (!res)
		return NULL;
	return tmp;
}

__bpf_kfunc struct bpf_mem_ptr *
bpf__strncpy(void *dst, int dst__sz, const char *src, int len)
{
	struct bpf_mem_ptr *tmp = dst;
	char *res = strncpy(tmp->ptr, src, len);
	if (!res)
		return NULL;
	tmp->ptr = res;
	return tmp;
}

__bpf_kfunc int
bpf__strncmp(char *dst, int len, void *src, int src__sz)
{
        struct bpf_mem_ptr *tmp = src;
	return strncmp(dst, tmp->ptr, len);
}

__bpf_kfunc __u32
bpf_parse_header_msg(void *src, int src__sz)
{
	return parse_protocol_impl(src);
}

__bpf_kfunc struct bpf_mem_ptr *
bpf_get_msg_header_element(const char *src)
{
	return get_protocol_element_impl(src);
}

__diag_pop();

BTF_SET8_START(bpf_kmesh_kfunc)
BTF_ID_FLAGS(func, bpf__strnlen)
BTF_ID_FLAGS(func, bpf__strnstr)
BTF_ID_FLAGS(func, bpf__strncpy)
BTF_ID_FLAGS(func, bpf__strncmp)
BTF_ID_FLAGS(func, bpf_parse_header_msg)
BTF_ID_FLAGS(func, bpf_get_msg_header_element)
BTF_SET8_END(bpf_kmesh_kfunc)

static const struct btf_kfunc_id_set bpf_kmesh_kfunc_set = {
	.owner	= THIS_MODULE,
	.set	= &bpf_kmesh_kfunc,
};

int __init bpf_kmesh_kfunc_init(void)
{
	int ret;
	ret = register_btf_kfunc_id_set(BPF_PROG_TYPE_UNSPEC, &bpf_kmesh_kfunc_set);
	if (ret < 0) {
		pr_err("ret is not zero:%d\n", ret);
		return ret;
	}
	return 0;
}

void __exit bpf_kmesh_kfunc_exit(void)
{
	return;
}

MODULE_LICENSE("Dual BSD/GPL");

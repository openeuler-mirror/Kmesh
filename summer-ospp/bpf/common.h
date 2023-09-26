
struct endpoint_key {
	__u32   ip4;
}__attribute__ ((__packed__));


/* Value of endpoint map */
struct endpoint_info {
	__u32		ifindex;
	__u64 		mac;
  	__u64		nodeMac;
} __attribute__ ((__packed__));

struct lb4_key {
	__u32 address;		/* Service virtual IPv4 address */
	// __be16 dport;		/* L4 port filter, if unset, all ports apply */
	__u32 backend_slot;	/* Backend iterator, 0 indicates the svc frontend */
} __attribute__ ((__packed__));

struct simple_ct_key {
    __u32 src_ip;
    __u32 dst_ip;
};

static __always_inline __u16
csum_fold_helper(__u64 csum)
{
    int i;
#pragma unroll
    for (i = 0; i < 4; i++)
    {
        if (csum >> 16)
            csum = (csum & 0xffff) + (csum >> 16);
    }
    return ~csum;
}

static __always_inline __u16
iph_csum(struct iphdr *iph)
{
    iph->check = 0;
    unsigned long long csum = bpf_csum_diff(0, 0, (unsigned int *)iph, sizeof(struct iphdr), 0);
    return csum_fold_helper(csum);
}

static __always_inline int ipv4_hdrlen(const struct iphdr *ip4)
{
	return ip4->ihl * 4;
}
// Copyright Authors of HActiV

// bpfcode package for eBPF Code
package bpfcode

const NetworkCcode string = `
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <uapi/linux/ptrace.h>
#include <net/sock.h>
#include <net/net_namespace.h>
#include <linux/nsproxy.h>

struct event_t {
    u32 pid;
    __u32 src_ip;
    __u32 dst_ip;
    __u8  protocol;
    __u64 packet_count;
    u32 namespaceinum;
    __u32 packet_size;
    __u16 dst_port;
    __u8 is_outgoing;
};

BPF_PERF_OUTPUT(events);

static inline int process_packet(struct pt_regs *ctx, struct sk_buff *skb, __u8 is_outgoing) {
    struct event_t event = {};
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    
    struct nsproxy *nsproxy;
    struct net *net_ns;
    unsigned int inum;
    if (bpf_probe_read_kernel(&nsproxy, sizeof(nsproxy), &task->nsproxy))
        return 0;
    // net_ns 읽기
    bpf_probe_read(&net_ns, sizeof(net_ns), &nsproxy->net_ns);
    
    // net namespace inode 번호 읽기
    bpf_probe_read(&inum, sizeof(inum), &net_ns->ns.inum);
    
    event.namespaceinum = inum;

    struct iphdr ip;
    unsigned char *head;
    __u16 network_header;

    bpf_probe_read(&head, sizeof(head), &skb->head);
    bpf_probe_read(&network_header, sizeof(network_header), &skb->network_header);
    bpf_probe_read(&ip, sizeof(ip), head + network_header);

    if (ip.saddr == 0x0100007F || ip.daddr == 0x0100007F) {
        return 0;
    }

    event.src_ip = ip.saddr;
    event.dst_ip = ip.daddr;
    event.protocol = ip.protocol;
    event.packet_count = 1;
    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.packet_size = skb->len;
    event.is_outgoing = is_outgoing;

    if (ip.protocol == IPPROTO_TCP) {
        struct tcphdr tcp;
        bpf_probe_read(&tcp, sizeof(tcp), head + network_header + sizeof(ip));
        event.dst_port = ntohs(tcp.dest);
    } else if (ip.protocol == IPPROTO_UDP) {
        struct udphdr udp;
        bpf_probe_read(&udp, sizeof(udp), head + network_header + sizeof(ip));
        event.dst_port = ntohs(udp.dest);
    } else {
        event.dst_port = 0;
    }

    events.perf_submit(ctx, &event, sizeof(event));
    return 0;
}

int kprobe__ip_rcv(struct pt_regs *ctx, struct sk_buff *skb) {
    return process_packet(ctx, skb, 0);
}

int kprobe__ip_output(struct pt_regs *ctx, struct net *net, struct sock *sk, struct sk_buff *skb) {
    return process_packet(ctx, skb, 1);
}
`

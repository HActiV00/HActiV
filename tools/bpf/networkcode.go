package bpfcode

const NetworkCcode string = `
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <uapi/linux/ptrace.h>
#include <net/sock.h>

struct event_t {
    u32 pid;
    __u32 src_ip;
    __u32 dst_ip;
    __u8  protocol;
    __u64 packet_count;
};

BPF_PERF_OUTPUT(events);

int kprobe__ip_rcv(struct pt_regs *ctx, struct sk_buff *skb) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    

    struct iphdr ip;
    unsigned char *head;
    __u16 network_header;

    bpf_probe_read(&head, sizeof(head), &skb->head);
    bpf_probe_read(&network_header, sizeof(network_header), &skb->network_header);
    bpf_probe_read(&ip, sizeof(ip), head + network_header);

    if (ip.saddr == 0x0100007F || ip.daddr == 0x0100007F) {
        return 0;
    }

    struct event_t event = {};
    event.src_ip = ip.saddr;
    event.dst_ip = ip.daddr;
    event.protocol = ip.protocol;
    event.packet_count = 1;

    event.pid = pid_tgid >> 32;
    events.perf_submit(ctx, &event, sizeof(event));
    return 0;
}
`

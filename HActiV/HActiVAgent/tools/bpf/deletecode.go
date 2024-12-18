// Copyright Authors of HActiV

// bpfcode package for eBPF Code
package bpfcode

const Delete string = `
#include <linux/nsproxy.h>
#include <uapi/linux/ptrace.h>
#include <linux/fs.h>
#include <net/net_namespace.h>
#include <linux/cred.h>

struct event_t {
    u32 uid;
    u32 gid;
    u32 pid;
    u32 ppid;
    char comm[16];
    char filename[200];
    u32 op;
    u32 namespaceinum;
};

BPF_PERF_OUTPUT(events);

int trace_unlinkat(struct pt_regs *ctx, int dfd, struct filename *name) {
    struct event_t event = {};
    struct task_struct *task;

    task = (struct task_struct *)bpf_get_current_task();
    u64 ugid = bpf_get_current_uid_gid();

    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.ppid = task->real_parent->tgid;
    event.uid = ugid & 0xFFFF;
    event.gid = ugid >> 32;
    event.op = 2;

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
    
    bpf_probe_read_str(&event.filename, sizeof(event.filename), name->name);

    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    events.perf_submit(ctx, &event, sizeof(event));
    
    return 0;
}

int trace_truncate(struct pt_regs *ctx, const char *pathname) {
    struct event_t event = {};
    struct task_struct *task;

    task = (struct task_struct *)bpf_get_current_task();
    u64 ugid = bpf_get_current_uid_gid();

    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.ppid = task->real_parent->tgid;
    event.uid = ugid & 0xFFFF;
    event.gid = ugid >> 32;
    event.op = 1;

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

    if (pathname != NULL) {
        bpf_probe_read_user_str(&event.filename, sizeof(event.filename), pathname);
    }
    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    events.perf_submit(ctx, &event, sizeof(event));
    
    return 0;
}
`

// Copyright Authors of HActiV

// bpfcode package for eBPF Code
package bpfcode

const Delete string = `
#include <uapi/linux/ptrace.h>
#include <linux/fs.h>
#include <linux/sched.h>
#include <linux/nsproxy.h>
#include <linux/ns_common.h>
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

struct mnt_namespace {
    #if LINUX_VERSION_CODE < KERNEL_VERSION(5, 11, 0)
        atomic_t count;
    #endif
    struct ns_common ns;
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
    struct mnt_namespace *mnt_ns;
    unsigned int inum;
    if (bpf_probe_read_kernel(&nsproxy, sizeof(nsproxy), &task->nsproxy))
        return 0;
    if (bpf_probe_read_kernel(&mnt_ns, sizeof(mnt_ns), &nsproxy->mnt_ns))
        return 0;
    if (bpf_probe_read_kernel(&inum, sizeof(inum), &mnt_ns->ns.inum))
        return 0;
    event.namespaceinum = inum;

    bpf_probe_read_str(&event.filename, sizeof(event.filename), name->name);

    bpf_trace_printk("Test %s\\n", event.filename);
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
    struct mnt_namespace *mnt_ns;
    unsigned int inum;
    if (bpf_probe_read_kernel(&nsproxy, sizeof(nsproxy), &task->nsproxy))
        return 0;
    if (bpf_probe_read_kernel(&mnt_ns, sizeof(mnt_ns), &nsproxy->mnt_ns))
        return 0;
    if (bpf_probe_read_kernel(&inum, sizeof(inum), &mnt_ns->ns.inum))
        return 0;
    event.namespaceinum = inum;

    if (pathname != NULL) {
        bpf_probe_read_user_str(&event.filename, sizeof(event.filename), pathname);
    }
    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    events.perf_submit(ctx, &event, sizeof(event));
    
    return 0;
}
`

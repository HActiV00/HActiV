// Copyright Authors of HActiV

// bpfcode package for eBPF Code
package bpfcode

const OpenCcode = `
#include <uapi/linux/ptrace.h>
#include <linux/nsproxy.h>
#include <linux/sched.h>
#include <linux/ns_common.h>

struct event_t {
    u32 pid;
    u32 ppid;
    u32 uid;
    u32 gid;
    int ret;
    char comm[TASK_COMM_LEN];
    char filename[256];
    u32 namespaceinum;
};

struct mnt_namespace {
    #if LINUX_VERSION_CODE < KERNEL_VERSION(5, 11, 0)
        atomic_t count;
    #endif
    struct ns_common ns;
};

BPF_PERF_OUTPUT(events);

int trace_sys_enter_openat(struct tracepoint__syscalls__sys_enter_openat *args)
{
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t event = {};

    event.pid = pid_tgid >> 32;
    event.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    event.gid = bpf_get_current_uid_gid() >> 32;
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    event.ppid = task->real_parent->tgid;

    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;
    
    struct nsproxy *nsproxy;
    struct mnt_namespace *mnt_ns;
    unsigned int inum;
    u64 ns_id;

    if (bpf_probe_read_kernel(&nsproxy, sizeof(nsproxy), &task->nsproxy))
        return 0;
    if (bpf_probe_read_kernel(&mnt_ns, sizeof(mnt_ns), &nsproxy->mnt_ns))
        return 0;
    if (bpf_probe_read_kernel(&inum, sizeof(inum), &mnt_ns->ns.inum))
        return 0;
    event.namespaceinum =  inum;

    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    bpf_probe_read_user_str(event.filename, sizeof(event.filename), args->filename);

    events.perf_submit(args, &event, sizeof(event));
    return 0;
}
`

// Copyright Authors of HActiV

// bpfcode package for eBPF Code
package bpfcode

const MemoryCcode = `
#include <linux/nsproxy.h>
#include <uapi/linux/ptrace.h>
#include <linux/mm_types.h>
#include <net/net_namespace.h>
#include <linux/cred.h>
#include <linux/mm.h>

BPF_PERF_OUTPUT(events);

struct event_t {
    u32 uid;
    u32 gid;
    u32 pid;
    u32 ppid;
    char comm[TASK_COMM_LEN];
    char syscall[16];
    u64 start_addr;
    u64 end_addr;
    u64 size;
    u32 prot;
    u32 namespaceinum;
    char mapping_type[16];
};

TRACEPOINT_PROBE(syscalls, sys_enter_mmap) {
    struct event_t event = {};
    u64 ugid = bpf_get_current_uid_gid();
    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.uid = ugid & 0xFFFFFFFF;
    event.gid = ugid >> 32;

    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    struct task_struct *parent_task;
    bpf_probe_read_kernel(&parent_task, sizeof(parent_task), &task->real_parent);
    u32 ppid;
    bpf_probe_read_kernel(&ppid, sizeof(ppid), &parent_task->tgid);
    event.ppid = ppid;

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

    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    __builtin_memcpy(&event.syscall, "mmap", 4);
    event.start_addr = (u64)args->addr;
    event.size = args->len;
    event.prot = args->prot;
    event.end_addr = event.start_addr + event.size;

    events.perf_submit(args, &event, sizeof(event));
    return 0;
}

TRACEPOINT_PROBE(syscalls, sys_enter_mprotect) {
    struct event_t event = {};
    u64 ugid = bpf_get_current_uid_gid();
    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.uid = ugid & 0xFFFFFFFF;
    event.gid = ugid >> 32;

    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    struct task_struct *parent_task;
    bpf_probe_read_kernel(&parent_task, sizeof(parent_task), &task->real_parent);
    u32 ppid;
    bpf_probe_read_kernel(&ppid, sizeof(ppid), &parent_task->tgid);
    event.ppid = ppid;

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

    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    __builtin_memcpy(&event.syscall, "mprotect", 8);
    event.start_addr = (u64)args->start;
    event.size = args->len;
    event.prot = args->prot;
    event.end_addr = event.start_addr + event.size;

    events.perf_submit(args, &event, sizeof(event));
    return 0;
}
`

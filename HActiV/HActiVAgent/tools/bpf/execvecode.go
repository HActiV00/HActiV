// Copyright Authors of HActiV

// bpfcode package for eBPF Code
package bpfcode

const ExecveCcode = `
#include <linux/nsproxy.h>
#include <linux/cred.h>
#include <net/net_namespace.h>

#define MAX_ARGS_SIZE 100

BPF_PERF_OUTPUT(events);

struct event_t {
	u32 uid;
	u32 gid;
    u32 pid;
    u32 ppid;
    u32 puid;
    u32 pgid;
    char comm[TASK_COMM_LEN];
    char filename[100];
    char arg1[200];
    u32 namespaceinum;
};

TRACEPOINT_PROBE(syscalls, sys_enter_execve) {
    struct event_t event = {};

	u64 ugid = bpf_get_current_uid_gid();
    event.pid = bpf_get_current_pid_tgid() >> 32;
	event.uid = ugid & 0xFFFFFFFF;
	event.gid = ugid>>32;

    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    struct task_struct *parent_task;
    bpf_probe_read(&parent_task, sizeof(parent_task), &task->real_parent);

    u64 ppid_tgid, parent_uid_gid;
    bpf_probe_read(&ppid_tgid, sizeof(ppid_tgid), &parent_task->tgid);
    event.ppid = (u32)ppid_tgid;

    struct cred *parent_cred;
    bpf_probe_read(&parent_cred, sizeof(parent_cred), &parent_task->cred);
    bpf_probe_read(&parent_uid_gid, sizeof(parent_uid_gid), &parent_cred->uid);

    event.puid = (u32)parent_uid_gid;
    event.pgid = (u32)(parent_uid_gid >> 32);
	
    const char **argp;
    bpf_probe_read(&argp, sizeof(argp), &args->argv);    
    
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
    bpf_probe_read_user_str(event.filename, sizeof(event.filename), args->filename);

    int offset = 0;
    char *arg;

    #pragma unroll
    for (int i = 1; i < 20 && offset < MAX_ARGS_SIZE - 1; i++) {
        if (bpf_probe_read(&arg, sizeof(arg), &argp[i]) || !arg) break;

        int len = bpf_probe_read_user_str(&event.arg1[offset], MAX_ARGS_SIZE - offset, arg);
        if (len <= 0) break;

        offset += len - 1;
        event.arg1[offset++] = ' ';
    }

    if (offset > 0)
        event.arg1[offset - 1] = '\0';
    else
        event.arg1[0] = '\0';

    events.perf_submit(args, &event, sizeof(event));
    return 0;
    }

`

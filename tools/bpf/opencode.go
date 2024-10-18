package bpfcode

const OpenCcode = `
#include <uapi/linux/ptrace.h>
#include <linux/sched.h>

struct event_t {
    u32 pid;
    u32 ppid;
    u32 uid;
    u32 gid;
    int ret;
    char comm[TASK_COMM_LEN];
    char filename[256];
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

    // HActiV Code Exception
    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;
    
    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    bpf_probe_read_user_str(event.filename, sizeof(event.filename), args->filename);

    events.perf_submit(args, &event, sizeof(event));
    return 0;
}
`

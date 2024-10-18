package bpfcode

const LogFileAccess string = `
#include <uapi/linux/ptrace.h>
#include <linux/fs.h>

struct event_t {
	u32 uid;
	u32 gid;
    u32 pid;
    u32 ppid;
    char comm[16];
    char filename[200];
};

BPF_PERF_OUTPUT(events);

// 파일이 열릴 때 호출되는 함수
int trace_open(struct pt_regs *ctx, struct file *file) {
    struct event_t event = {};
    struct task_struct *task;

    task = (struct task_struct *)bpf_get_current_task();
	u64 ugid = bpf_get_current_uid_gid();

    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.ppid = task->real_parent->tgid;
	event.uid = ugid & 0xFFFF;
	event.gid = ugid>>32;

    // HActiV Code Exception
    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;

    bpf_get_current_comm(&event.filename, sizeof(event.filename));
    // 프로세스 이름 가져오기
    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    events.perf_submit(ctx, &event, sizeof(event));
    return 0;
}
`

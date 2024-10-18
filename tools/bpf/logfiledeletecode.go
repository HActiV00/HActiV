package bpfcode

const LogFileDelete string = `
#include <uapi/linux/ptrace.h>
#include <linux/fs.h>

struct event_t {
	uid_t uid;
	gid_t gid;
    u32 pid;
    u32 ppid;
    char comm[16];
    char filename[200];
    u32 op;  // 1: truncate, 2: delete
};

BPF_PERF_OUTPUT(events);

// unlink 또는 unlinkat 시스템 콜이 호출될 때 실행되는 함수
int trace_unlink(struct pt_regs *ctx) {

    struct event_t event = {};

    struct task_struct *task;

    task = (struct task_struct *)bpf_get_current_task();
	u64 ugid = bpf_get_current_uid_gid();

    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.ppid = task->real_parent->tgid;
	event.uid = ugid & 0xFFFF;
	event.gid = ugid>>32;
    event.op = 2; // delete operation

    // HActiV Code Exception
    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;
    
    //filename 미 구현이여서 프로세스 이름 가져옴
    bpf_get_current_comm(&event.filename, sizeof(event.filename));
    
    // 프로세스 이름 가져오기
    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    // 이벤트 전달
    events.perf_submit(ctx, &event, sizeof(event));
    
    return 0;
}

// truncate 또는 ftruncate 시스템 콜이 호출될 때 실행되는 함수
int trace_truncate(struct pt_regs *ctx) {
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

    event.op = 1; // truncate operation
    //파일 이름 가져오기 임시적으로 프로세스 이름 가져옴
    bpf_get_current_comm(&event.filename, sizeof(event.filename));
    
    // 프로세스 이름 가져오기
    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    // 이벤트 전달
    events.perf_submit(ctx, &event, sizeof(event));
    
    return 0;
}
`

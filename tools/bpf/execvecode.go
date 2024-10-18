package bpfcode

const ExecveCcode = `
#include <linux/sched.h>
#include <linux/nsproxy.h>
#include <linux/ns_common.h>

#define MAX_ARGS_SIZE 100

BPF_PERF_OUTPUT(events);

struct event_t {
	uid_t uid;
	gid_t gid;
    u32 pid;
    u32 ppid;
    char comm[TASK_COMM_LEN];
    char filename[200];
    char arg1[200];
};

TRACEPOINT_PROBE(syscalls, sys_enter_execve) {
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

	bpf_get_current_comm(&event.comm, sizeof(event.comm));
    bpf_probe_read_user_str(event.filename, sizeof(event.filename), args->filename);
	const char **argp;
    bpf_probe_read(&argp, sizeof(argp), &args->argv);

	// GPT Code
    int offset = 0;
    #pragma unroll
    for (int i = 1; i < 20; i++) {  // 최대 16개 인자로 제한
        const char *arg;
        bpf_probe_read(&arg, sizeof(arg), &argp[i]);
        if (!arg)
            break;
        
        int len = bpf_probe_read_user_str(&event.arg1[offset], MAX_ARGS_SIZE - offset, arg);
        if (len <= 0)
            break;
        
        offset += len;
        event.arg1[offset - 1] = ' ';  // 인자 사이에 공백 추가
        
        if (offset >= MAX_ARGS_SIZE - 1)
            break;
    }
    event.arg1[offset] = '\0';  // 문자열 종료

    events.perf_submit(args, &event, sizeof(event));
    return 0;
}
`

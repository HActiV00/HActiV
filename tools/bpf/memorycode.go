package bpfcode

const MemoryCcode = `
#include <uapi/linux/ptrace.h>
#include <linux/sched.h>

struct event_t {
    u32 pid;
    u32 ppid;
    u32 uid;
    u32 gid;
    int ret;
    char comm[TASK_COMM_LEN];
    char syscall[16];
    char event_type[16];
    u64 start_addr;
    u64 end_addr;
} ;

BPF_PERF_OUTPUT(events);
BPF_HASH(call_events, u64, struct event_t);
BPF_HASH(memory_ranges, u64, u64);

// Detect Mmap
int kprobe__sys_mmap(struct pt_regs *ctx) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t event = {};
    event.pid = pid_tgid >> 32;
    event.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    event.gid = bpf_get_current_uid_gid() >> 32;
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    event.ppid = task->real_parent->tgid;
    
    //HActiV Code Exception
    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;

    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    event.start_addr = PT_REGS_PARM1(ctx);
    event.end_addr = event.start_addr + PT_REGS_PARM2(ctx);
    __builtin_memcpy(event.syscall, "mmap", sizeof("mmap"));
    __builtin_memcpy(event.event_type, "none", sizeof("none"));
    call_events.update(&pid_tgid, &event);
    memory_ranges.update(&pid_tgid, &event.start_addr);
    return 0;
}

// mmap 시스템 호출 종료 탐지
int kretprobe__sys_mmap(struct pt_regs *ctx) {
    int ret = PT_REGS_RC(ctx);
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t *eventp = call_events.lookup(&pid_tgid);
    if (eventp == 0) {
        return 0;
    }
    eventp->ret = ret;
    events.perf_submit(ctx, eventp, sizeof(*eventp));
    return 0;
}

// mprotect 시스템 호출 탐지
int kprobe__sys_mprotect(struct pt_regs *ctx) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t event = {};
    event.pid = pid_tgid >> 32;
    event.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    event.gid = bpf_get_current_uid_gid() >> 32;
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    event.ppid = task->real_parent->tgid;

    //HActiV Code Exception
    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;

    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    event.start_addr = PT_REGS_PARM1(ctx);
    event.end_addr = event.start_addr + PT_REGS_PARM2(ctx);
    __builtin_memcpy(event.syscall, "mprotect", sizeof("mprotect"));
    __builtin_memcpy(event.event_type, "none", sizeof("none"));
    call_events.update(&pid_tgid, &event);
    memory_ranges.update(&pid_tgid, &event.start_addr);
    return 0;
}

// mprotect 시스템 호출 종료 탐지
int kretprobe__sys_mprotect(struct pt_regs *ctx) {
    int ret = PT_REGS_RC(ctx);
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t *eventp = call_events.lookup(&pid_tgid);
    if (eventp == 0) {
        return 0;
    }
    eventp->ret = ret;
    events.perf_submit(ctx, eventp, sizeof(*eventp));
    return 0;
}

// read 시스템 호출 탐지
int kprobe__sys_read(struct pt_regs *ctx) {
u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;
    
    struct event_t *eventp = call_events.lookup(&pid_tgid);
    if (eventp == 0) {
        return 0;
    }

    // HActiV Code Exception
    if (pid == Host_Pid || eventp->ppid == Host_Pid)
        return 0;

    // 메모리 범위 검사
    u64 *start_addr = memory_ranges.lookup(&pid_tgid);
    if (start_addr != 0 && PT_REGS_PARM1(ctx) >= *start_addr && PT_REGS_PARM1(ctx) < *start_addr + (eventp->end_addr - *start_addr)) {
        __builtin_memcpy(eventp->event_type, "read", sizeof("read"));
    } else {
        __builtin_memcpy(eventp->event_type, "none", sizeof("none"));
    }
    events.perf_submit(ctx, eventp, sizeof(*eventp));
    return 0;
}

// write 시스템 호출 탐지
int kprobe__sys_write(struct pt_regs *ctx) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;
    
    struct event_t *eventp = call_events.lookup(&pid_tgid);
    if (eventp == 0) {
        return 0;
    }

    // HActiV Code Exception
    if (pid == Host_Pid || eventp->ppid == Host_Pid)
        return 0;

    // 메모리 범위 검사
    u64 *start_addr = memory_ranges.lookup(&pid_tgid);
    if (start_addr != 0 && PT_REGS_PARM1(ctx) >= *start_addr && PT_REGS_PARM1(ctx) < *start_addr + (eventp->end_addr - *start_addr)) {
        __builtin_memcpy(eventp->event_type, "write", sizeof("write"));
    } else {
        __builtin_memcpy(eventp->event_type, "none", sizeof("none"));
    }
    events.perf_submit(ctx, eventp, sizeof(*eventp));
    return 0;
}
`

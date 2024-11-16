// Copyright Authors of HActiV

// bpfcode package for eBPF Code
package bpfcode

const MemoryCcode = `
#include <uapi/linux/ptrace.h>
#include <linux/sched.h>
#include <linux/nsproxy.h>
#include <linux/ns_common.h>
#include <linux/mman.h>

struct event_t {
    u32 pid;
    u32 ppid;
    u32 uid;
    u32 gid;
    // u32 ret;
    char comm[TASK_COMM_LEN];
    char syscall[16];
    char event_type[16]; // 208
    u32 namespaceinum; // 변경: u32에서 u64로 변경 다시 u32로
    u64 start_addr;
    u64 end_addr;
};

struct last_event {
    u32 pid;
    u32 namespaceinum;
    u64 timestamp;
    u64 start_addr;
    u64 end_addr;
};

#define UINT64_MAX 0xFFFFFFFFFFFFFFFFULL

struct mnt_namespace {
    #if LINUX_VERSION_CODE < KERNEL_VERSION(5, 11, 0)
        atomic_t count;
    #endif
    struct ns_common ns;
};

BPF_PERF_OUTPUT(events);
BPF_HASH(last_events, u32, struct last_event); // 중복 제거용 Hash
//BPF_HASH(call_events, u64, struct event_t); kret 미사용

// mmap 시스템 호출 추적
int kprobe__sys_mmap(struct pt_regs *ctx) {
    u64 ts = bpf_ktime_get_ns();
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t event = {};
    event.pid = pid_tgid >> 32;
    event.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    event.gid = bpf_get_current_uid_gid() >> 32;
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    event.ppid = task->real_parent->tgid;

    //HActiV Except
    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;

    // 네임스페이스 정보 가져오기
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
    event.namespaceinum = inum;

    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    event.start_addr = (u64)PT_REGS_PARM1(ctx);
    u64 size = (u64)PT_REGS_PARM2(ctx);
    if (size > UINT64_MAX - event.start_addr) {
        event.end_addr = UINT64_MAX;
    } else {
        event.end_addr = event.start_addr + size;
    }

    __builtin_memcpy(event.syscall, "mmap", sizeof("mmap"));
    //call_events.update(&pid_tgid, &event); //kret 미사용
    events.perf_submit(ctx, &event, sizeof(event));
    return 0;
}
    //return 값 미출력으로 필요X
/*
    int kretprobe__sys_mmap(struct pt_regs *ctx) {
        u32 ret = PT_REGS_RC(ctx);
        u64 pid_tgid = bpf_get_current_pid_tgid();
        struct event_t *eventp = call_events.lookup(&pid_tgid);
        if (eventp == 0) {
            return 0;
        }
    
        //bpf_trace_printk("Test %u\\n", eventp->namespaceinum);

        eventp->ret = ret;
        bpf_trace_printk("mmap return: PID %d, ret %d\\n", eventp->pid, ret);
        events.perf_submit(ctx, eventp, sizeof(*eventp));
        call_events.delete(&pid_tgid);
        return 0;
    }
*/
// brk 시스템 호출 추적
int kprobe__sys_brk(struct pt_regs *ctx) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t event = {};
    event.pid = pid_tgid >> 32;
    event.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    event.gid = bpf_get_current_uid_gid() >> 32;
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    event.ppid = task->real_parent->tgid;
    
    //HActiV Process Except
    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;
    
    // 네임스페이스 정보 가져오기
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
    event.namespaceinum = inum;

    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    event.start_addr = (u64)PT_REGS_PARM1(ctx);
    u64 size = (u64)PT_REGS_PARM2(ctx);
    if (size > UINT64_MAX - event.start_addr) {
        event.end_addr = UINT64_MAX;
    } else {
        event.end_addr = event.start_addr + size;
    }
    __builtin_memcpy(event.syscall, "brk", sizeof("brk"));
    events.perf_submit(ctx, &event, sizeof(event));
    // call_events.update(&pid_tgid, &event); //kret 미사용
    return 0;
}
    //return 값 미출력으로 필요X
/*
    int kretprobe__sys_brk(struct pt_regs *ctx) {
        u32 ret = PT_REGS_RC(ctx);
        u64 pid_tgid = bpf_get_current_pid_tgid();
        struct event_t *eventp = call_events.lookup(&pid_tgid);
        if (eventp == 0) {
            return 0;
        }

            // 네임스페이스 정보 가져오기
        struct nsproxy *nsproxy;
        struct mnt_namespace *mnt_ns;
        u32 inum;
        struct task_struct *task = (struct task_struct *)bpf_get_current_task();
        if (bpf_probe_read_kernel(&nsproxy, sizeof(nsproxy), &task->nsproxy))
            return 0;
        if (bpf_probe_read_kernel(&mnt_ns, sizeof(mnt_ns), &nsproxy->mnt_ns))
            return 0;
        if (bpf_probe_read_kernel(&inum, sizeof(inum), &mnt_ns->ns.inum))
            return 0;
        eventp->namespaceinum = inum;
        bpf_trace_printk("Detected mmap call: PID, Namespace %u\\n", inum);

        eventp->ret = ret;
        bpf_trace_printk("brk return: PID %d, ret %d\\n", eventp->pid, ret);
        events.perf_submit(ctx, eventp, sizeof(*eventp));
        call_events.delete(&pid_tgid);
        return 0;
    }
*/
// mprotect 시스템 호출 추적
int kprobe__sys_mprotect(struct pt_regs *ctx) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t event = {};
    event.pid = pid_tgid >> 32;
    event.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    event.gid = bpf_get_current_uid_gid() >> 32;
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    event.ppid = task->real_parent->tgid;

    if (event.pid == Host_Pid || event.ppid == Host_Pid)
        return 0;

    // 네임스페이스 정보 가져오기
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
    event.namespaceinum = inum;

    unsigned long prot;
    prot = (unsigned long)PT_REGS_PARM3(ctx);
    bpf_trace_printk("mmap called with prot: 0x%lx\n", prot);

    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    event.start_addr = (u64)PT_REGS_PARM1(ctx);
    u64 size = (u64)PT_REGS_PARM2(ctx);
    if (size > UINT64_MAX - event.start_addr) {
        event.end_addr = UINT64_MAX;
    } else {
        event.end_addr = event.start_addr + size;
    }
    
    __builtin_memcpy(event.syscall, "mprotect", sizeof("mprotect"));
    events.perf_submit(ctx, &event, sizeof(event));
    // call_events.update(&pid_tgid, &event); //kret 미사용
    return 0;
}
    //return 값 미출력으로 필요X
/*
int kretprobe__sys_mprotect(struct pt_regs *ctx) {
    u32 ret = PT_REGS_RC(ctx);
    u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t *eventp = call_events.lookup(&pid_tgid);
    if (eventp == 0) {
        return 0;
    }

    // 네임스페이스 정보 가져오기
    struct nsproxy *nsproxy;
    struct mnt_namespace *mnt_ns;
    u32 inum;
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    if (bpf_probe_read_kernel(&nsproxy, sizeof(nsproxy), &task->nsproxy))
        return 0;
    if (bpf_probe_read_kernel(&mnt_ns, sizeof(mnt_ns), &nsproxy->mnt_ns))
        return 0;
    if (bpf_probe_read_kernel(&inum, sizeof(inum), &mnt_ns->ns.inum))
        return 0;
    eventp->namespaceinum = inum;
    bpf_trace_printk("Detected mmap call: PID, Namespace %u\\n", inum);

    eventp->ret = ret;
    bpf_trace_printk("mprotect return: PID %d, ret %d\\n", eventp->pid, ret);
    events.perf_submit(ctx, eventp, sizeof(*eventp));
    call_events.delete(&pid_tgid);
    return 0;
}
*/
`

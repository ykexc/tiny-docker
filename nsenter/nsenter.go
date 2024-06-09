package nsenter

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void enter_namespace(void) {
   // 这里的代码会在Go运行时启动前执行，它会在单线程的C上下文中运行
	char *tinydocker_pid;
	tinydocker_pid = getenv("tinydocker_pid");
	if (tinydocker_pid) {
		fprintf(stdout, "got tinydocker_pid=%s\n", tinydocker_pid);
	} else {
		fprintf(stdout, "missing tinydocker_pid env skip nsenter");
		// 如果没有指定PID就不需要继续执行，直接退出
		return;
	}
	char *tinydocker_cmd;
	tinydocker_cmd = getenv("tinydocker_cmd");
	if (tinydocker_cmd) {
		fprintf(stdout, "got tinydocker_cmd=%s\n", tinydocker_cmd);
	} else {
		fprintf(stdout, "missing tinydocker_cmd env skip nsenter");
		// 如果没有指定命令也是直接退出
		return;
	}
	int i;
	char nspath[1024];
	// 需要进入的5种namespace
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		// 拼接对应路径，类似于/proc/pid/ns/ipc这样
		sprintf(nspath, "/proc/%s/ns/%s", tinydocker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);
		// 执行setns系统调用，进入对应namespace
		if (setns(fd, 0) == -1) {
			fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	// 在进入的Namespace中执行指定命令，然后退出
	int res = system(tinydocker_cmd);
	exit(0);
	return;
}
*/
import "C"

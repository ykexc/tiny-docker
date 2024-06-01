package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"tiny-docker/cgroups"
	"tiny-docker/cgroups/subsystems"
	"tiny-docker/container"
)

// Run 执行具体 command
/*
这里的Start方法是真正开始执行由NewParentProcess构建好的command的调用，它首先会clone出来一个namespace隔离的
进程，然后在子进程中，调用/proc/self/exe,也就是调用自己，发送init参数，调用我们写的init方法，
去初始化容器的一些资源。
*/
func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, wp := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Run parent.Start err:%v", err)
	}
	// 创建cgroup manager, 并通过调用set和apply设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	_ = cgroupManager.Set(res)
	_ = cgroupManager.Apply(parent.Process.Pid, res)
	// 在子进程创建后才能通过pipe来发送参数
	sendInitCommand(comArray, wp)
	_ = parent.Wait()
}

// 向管道中写
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is: %s", command)
	_, err := writePipe.WriteString(command)
	if err != nil {
		log.Error("write pipe fail: %s", command)
	}
	writePipe.Close()
}

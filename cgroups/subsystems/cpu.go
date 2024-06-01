package subsystems

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"tiny-docker/constant"
)

type CpuSubsystem struct{}

const (
	PeriodDefault = 100000
	Percent       = 100
)

func (s *CpuSubsystem) Name() string {
	return "cpu"
}

/**
cgroupV1使用
(1) mkdir mkdir /sys/fs/cgroup/cpu/cpu_test
(2) echo 20000 > /sys/fs/cgroup/cpu/cpu_test/cpu.cfs_quota_us  设置cpu最大占20%
(3) while : ; do : ; done &  开启一个死循环
(4) echo [pid] > /sys/fs/cgroup/cpu/cpu_test/tasks
(5) 观察top,cpu最大不会超过20%


cgroupsV2使用(底层与V1差别挺大的,但整体上也就那几个流程)
(1) 确保自己的linux有使用v2
(2) mkdir /sys/fs/cgroup/cpu_test
(3)设置cpu最大20%
echo "20000 100000" > /sys/fs/cgroup/cpu_test/cpu.max
(4) 开启一个死循环
while : ; do : ; done &
(5)echo [pid] > /sys/fs/cgroup/cpu_test/cgroup.procs
(6)观察top,cpu最大不会超过20%
*/

func (s *CpuSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.CpuCfsQuota == 0 && res.CpuShare == "" {
		return nil
	}
	subsysCgroupPath, err := getCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	// cpu.shares 控制的是CPU使用比例，不是绝对值
	if res.CpuShare != "" {
		if err = os.WriteFile(path.Join(subsysCgroupPath, "cpu.shares"), []byte(res.CpuShare), 0644); err != nil {
			return fmt.Errorf("set cgroup cpu share fail %v", err)
		}
	}
	// cpu.cfs_period_us & cpu.cfs_quota_us 控制的是CPU使用时间，单位是微秒，比如每1秒钟，这个进程只能使用200ms，相当于只能用20%的CPU
	if res.CpuCfsQuota != 0 {
		// cpu.cfs_period_us 默认为100000，即100ms, 所以直接写 20000 = 20000, 100000
		if err = os.WriteFile(path.Join(subsysCgroupPath, "cpu.cfs_period_us"), []byte(strconv.Itoa(PeriodDefault)), constant.Perm0644); err != nil {
			return fmt.Errorf("set cgroup cpu share fail %v", err)
		}

		// cpu.cfs_quota_us 则根据用户传递的参数来控制，比如参数为20，就是限制为20%CPU，所以把cpu.cfs_quota_us设置为cpu.cfs_period_us的20%就行
		// 这里只是简单的计算了下，并没有处理一些特殊情况，比如负数什么的
		if err = os.WriteFile(path.Join(subsysCgroupPath, "cpu.cfs_quota_us"), []byte(strconv.Itoa(PeriodDefault/Percent*res.CpuCfsQuota)), constant.Perm0644); err != nil {
			return fmt.Errorf("set cgroup cpu share fail %v", err)
		}
	}
	return nil
}

//流程是这样的 先设置(顺便先创建) -> 在申请(就是往task里面加入pid)  echo [pid] > /sys/fs/cgroup/cpu/cpu_test/tasks

func (s *CpuSubsystem) Apply(cgroupPath string, pid int, res *ResourceConfig) error {
	if res.CpuCfsQuota == 0 && res.CpuShare == "" {
		return nil
	}
	subsysCgroupPath, err := getCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
	if err = os.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), constant.Perm0644); err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}
	return nil
}

// Remove 删除文件夹就相当于取消的这次限制
func (s *CpuSubsystem) Remove(cgroupPath string) error {
	subsysCgroupPath, err := getCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsysCgroupPath)
}

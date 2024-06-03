package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"syscall"
)

const (
	lower    = "ubuntu-base-16.04.6-base-amd64"
	lowerTar = "ubuntu-base-16.04.6-base-amd64.tar"
)

// NewParentProcess 构建 command 用于启动一个新进程
/*
这里是父进程，也就是当前进程执行的内容。
1.这里的/proc/se1f/exe调用中，/proc/self/ 指的是当前运行进程自己的环境，exec 其实就是自己调用了自己，使用这种方式对创建出来的进程进行初始化
2.后面的args是参数，其中init是传递给本进程的第一个参数，在本例中，其实就是会去调用initCommand去初始化进程的一些环境和资源
3.下面的clone参数就是去fork出来一个新进程，并且使用了namespace隔离新创建的进程和外部环境。
4.如果用户指定了-it参数，就需要把当前进程的输入输出导入到标准输入输出上
*/
func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {

	// 创建匿名管道用于传递参数，将readPipe作为子进程的ExtraFiles，子进程从readPipe中读取参数
	// 父进程中则通过writePipe将参数写入管道
	rp, wp, err := os.Pipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{rp}
	rootPath := "/root"
	NewWorkSpace(rootPath, volume)
	cmd.Dir = path.Join(rootPath, "merged")
	return cmd, wp
}

func NewWorkSpace(rootPath, volume string) {
	createLower(rootPath)
	createDirs(rootPath)
	mountOverlayFS(rootPath)
	if volume != "" {
		mntPath := path.Join(rootPath, "merged")
		hostPath, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mountVolume(mntPath, hostPath, containerPath)
	}
}

// createLower 将busybox作为overlayfs的lower层
func createLower(rootPath string) {

	rfPath, rfTarPath := path.Join(rootPath, lower), path.Join(rootPath, lowerTar)
	log.Infof("rf:%s rfTar:%s", rfPath, rfTarPath)
	exists, err := PathExists(rfPath)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", rfPath, err)
	}
	if !exists {
		if err := os.Mkdir(rfPath, 0755); err != nil {
			log.Errorf("Mkdir dir %s error. %v", rfPath, err)
		}
		if _, err = exec.Command("tar", "-xvf", rfTarPath, "-C", rfPath).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", rfPath, err)
		}
	}
}

// createDirs 创建overlayfs需要的的merged、upper、worker目录
func createDirs(rootPath string) {
	for _, dir := range []string{path.Join(rootPath, "merged"), path.Join(rootPath, "upper"), path.Join(rootPath, "work")} {
		if err := os.Mkdir(dir, 0755); err != nil {
			log.Errorf("mkdir dir %s error. %v", dir, err)
		}
	}
}

// mountOverlayFS 挂载overlayfs
func mountOverlayFS(rootPath string) {
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", path.Join(rootPath, lower),
		path.Join(rootPath, "upper"), path.Join(rootPath, "work"))
	// 完整命令：mount -t overlay overlay -o lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work /root/merged
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, path.Join(rootPath, "merged"))
	log.Infof("mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootPath, volume string) {
	mntPath := path.Join(rootPath, "merged")

	//先umount volume,再overlay
	if volume != "" {
		_, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		unmountVolume(mntPath, containerPath)
	}
	umountOverlayFS(path.Join(rootPath, "merged"))
	deleteDirs(rootPath)
}

func umountOverlayFS(mntPath string) {
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func deleteDirs(rootPath string) {
	dirs := []string{
		path.Join(rootPath, "merged"),
		path.Join(rootPath, "upper"),
		path.Join(rootPath, "work"),
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Errorf("Remove dir %s error %v", dir, err)
		}
	}
}

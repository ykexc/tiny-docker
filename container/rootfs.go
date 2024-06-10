package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"tiny-docker/utils"
)

func NewWorkSpace(containerId, imageName, volume string) {
	createLower(containerId, imageName)
	createDirs(containerId)
	mountOverlayFS(containerId)
	if volume != "" {
		mntPath := utils.GetMerged(containerId)
		hostPath, containerPath, err := utils.VolumeUrlExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mountVolume(mntPath, hostPath, containerPath)
	}
}

// createLower 将busybox作为overlayfs的lower层
func createLower(containerId, imageName string) {

	// 根据 containerID 拼接出 lower 目录
	// 根据 imageName 找到镜像 tar，并解压到 lower 目录中
	lowerPath := utils.GetLower(containerId)
	imagePath := utils.GetImage(imageName)
	log.Infof("lower:%s image.tar:%s", lowerPath, imagePath)
	//检查目录是否存在
	exist, err := utils.PathExist(lowerPath)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", lowerPath, err)
	}
	// 不存在的话将image.tar解压到lower文件夹中
	if !exist {
		if err = os.MkdirAll(lowerPath, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", lowerPath, err)
		}
		if _, err = exec.Command("tar", "-xvf", imagePath, "-C", lowerPath).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", lowerPath, err)
		}
	}
}

// createDirs 创建overlayfs需要的的merged、upper、worker目录
func createDirs(containerId string) {
	for _, dir := range []string{utils.GetMerged(containerId), utils.GetUpper(containerId), utils.GetWork(containerId)} {
		if err := os.Mkdir(dir, 0777); err != nil {
			log.Errorf("mkdir dir %s error. %v", dir, err)
		}
	}
}

// mountOverlayFS 挂载overlayfs
func mountOverlayFS(containerId string) {
	// 拼接参数
	// e.g. lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work
	dirs := utils.GetOverlayFSDirs(utils.GetLower(containerId), utils.GetUpper(containerId), utils.GetWork(containerId))
	mergedPath := utils.GetMerged(containerId)
	// 完整命令：mount -t overlay overlay -o lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work /root/merged
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mergedPath)
	log.Infof("mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(containerId, volume string) {

	//先umount volume,再overlay
	if volume != "" {
		_, containerPath, err := utils.VolumeUrlExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mntPath := utils.GetMerged(containerId)
		unmountVolume(mntPath, containerPath)
	}
	umountOverlayFS(containerId)
	deleteDirs(containerId)
}

func umountOverlayFS(containerId string) {
	mntPath := utils.GetMerged(containerId)
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func deleteDirs(containerId string) {
	dirs := []string{
		utils.GetMerged(containerId),
		utils.GetUpper(containerId),
		utils.GetWork(containerId),
		utils.GetLower(containerId),
		utils.GetRoot(containerId), // root 目录也要删除
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Errorf("Remove dir %s error %v", dir, err)
		}
	}
}

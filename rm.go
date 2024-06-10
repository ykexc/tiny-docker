package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"tiny-docker/container"
	"tiny-docker/utils"
)

func removeContainer(containerId string, force bool) {
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerId, err)
		return
	}

	switch containerInfo.Status {
	case container.STOP:
		// 先删除配置目录,再删除rootfs
		if err = container.DeleteContainerInfo(containerId); err != nil {
			log.Errorf("Remove container [%s]'s config failed, detail: %v", containerId, err)
			return
		}
		container.DeleteWorkSpace(containerId, containerInfo.Volume)
	case container.RUNNING:
		if !force {
			log.Errorf("Couldn't remove running container [%s], Stop the container before attempting removal or"+
				" force remove", containerId)
			return
		}
		log.Infof("force delete running container [%s]", containerId)
		stopContainer(containerId)
		removeContainer(containerId, force)
	default:
		log.Errorf("Couldn't remove container,invalid status %s", containerInfo.Status)
		return
	}
}

func removeImage(imageName string) {
	imagePath := utils.GetImage(imageName)
	exist, err := utils.PathExist(imagePath)
	if err != nil {
		log.Errorf("Check image %s exist error: %v", imageName, err)
	}
	if !exist {
		log.Infof("Image %s has not", imageName)
	}
	_ = os.Remove(imagePath)

}

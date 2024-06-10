package main

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"tiny-docker/utils"
)

var ErrImageAlreadyExists = errors.New("image already exists")

func commitContainer(containerId, imageName string) error {
	mntPath := utils.GetMerged(containerId)
	imageTar := utils.GetImage(imageName)
	exists, err := utils.PathExist(imageTar)
	if err != nil {
		return errors.WithMessagef(err, "check is image [%s/%s] exist failed", imageName, imageTar)
	}
	if exists {
		return ErrImageAlreadyExists
	}
	log.Infof("commitContainer imageTar:%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput(); err != nil {
		log.Errorf("tar folder %s error %v", mntPath, err)
	}
	return nil
}

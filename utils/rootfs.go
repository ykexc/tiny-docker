package utils

import "fmt"

const (
	ImagePath       = "/var/lib/tiny-docker/image/"
	RootPath        = "/var/lib/tiny-docker/overlay2/"
	lowerDirFormat  = RootPath + "%s/lower"
	upperDirFormat  = RootPath + "%s/upper"
	workDirFormat   = RootPath + "%s/work"
	mergedDirFormat = RootPath + "%s/merged"
	overlayFsFormat = "lowerdir=%s,upperdir=%s,workdir=%s"
)

func GetRoot(containerID string) string {
	return RootPath + containerID
}

func GetImage(imageName string) string {
	return fmt.Sprintf("%s%s.tar", ImagePath, imageName)
}

func GetLower(containerID string) string {
	return fmt.Sprintf(lowerDirFormat, containerID)
}

func GetUpper(containerID string) string {
	return fmt.Sprintf(upperDirFormat, containerID)
}

func GetWork(containerID string) string {
	return fmt.Sprintf(workDirFormat, containerID)
}

func GetMerged(containerID string) string {
	return fmt.Sprintf(mergedDirFormat, containerID)
}

func GetOverlayFSDirs(lowerDir string, upperDir string, workDir string) string {
	return fmt.Sprintf(overlayFsFormat, lowerDir, upperDir, workDir)
}

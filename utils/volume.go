package utils

import (
	"fmt"
	"strings"
)

// VolumeUrlExtract 通过冒号分割解析volume目录，比如 -v /tmp:/tmp
func VolumeUrlExtract(volume string) (sourcePath, destinationPath string, err error) {
	parts := strings.Split(volume, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid volume [%s], must split by `:`", volume)
	}
	sourcePath, destinationPath = parts[0], parts[1]

	if sourcePath == "" || destinationPath == "" {
		return "", "", fmt.Errorf("invalid volume [%s], path can't be empty", volume)
	}

	return sourcePath, destinationPath, nil
}

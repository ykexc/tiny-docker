package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"tiny-docker/cgroups/subsystems"
	"tiny-docker/container"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			tinydocker run -it [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it", // 简单起见，这里把 -i 和 -t 参数合并成一个
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit,e.g.: -mem 100m",
		},
		cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu limit,e.g.: -cpu 100m",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit,e.g.: -cpuset 100m",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume,e.g.: -v /ect/conf:/etc/conf",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container,run background",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name，e.g.: -name mycontainer",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment,e.g. -e name=mydocker",
		},
	},
	/*
		这里是run命令执行的真正函数。
		1.判断参数是否包含command
		2.获取用户指定的command
		3.调用Run function去准备启动容器:
	*/
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}

		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]

		tty := context.Bool("it")
		detach := context.Bool("d")

		if tty && detach {
			return fmt.Errorf("it and d flag can not both provided")
		}

		if !detach {
			tty = true
		}

		log.Infof("createTty %v", tty)
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuSet:      context.String("cpuset"),
			CpuCfsQuota: context.Int("cpu"),
		}

		log.Info("resConf:", resConf)

		volume := context.String("v")
		containerName := context.String("name")
		envSlice := context.StringSlice("e")
		Run(tty, cmdArray, envSlice, resConf, volume, containerName, imageName)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	/*
		1.获取传递过来的 command 参数
		2.执行容器初始化操作
	*/
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		log.Infof("command: %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit container to image,e.g. mydocker commit 123456789 myimage",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing image name")
		}
		containerId := context.Args().Get(0)
		imageName := context.Args().Get(1)
		return commitContainer(containerId, imageName)
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "List all containers",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}

var logsCommand = cli.Command{
	Name:  "logs",
	Usage: "Tail logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("please input your container name")
		}
		containerName := context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "Execute a command in a container",
	Action: func(context *cli.Context) error {
		// 如果环境变量存在，说明C代码已经运行过了，即setns系统调用已经执行了，这里就直接返回，避免重复执行
		if os.Getenv(EnvExecPid) != "" {
			log.Infof("pid callback pid %v", os.Getpid())
			return nil
		}
		// 格式：mydocker exec 容器名字 命令，因此至少会有两个参数
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container name or command")
		}
		commandArray := context.Args().Tail()
		containerName := context.Args().Get(0)
		ExecContainer(containerName, commandArray)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "Stop a container",
	Action: func(context *cli.Context) error {
		// 期望输入是：tiny-docker stop 容器Id，如果没有指定参数直接打印错误
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container id")
		}
		containerId := context.Args().Get(0)
		stopContainer(containerId)
		return nil
	},
}

var rmCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers,e.g. mydocker rm 1234567890",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "f",
			Usage: "force remove container",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container id")
		}
		containerId := context.Args().Get(0)
		force := context.Bool("f")
		removeContainer(containerId, force)
		return nil
	},
}

var imagesCommand = cli.Command{
	Name:  "images",
	Usage: "List all images",
	Action: func(context *cli.Context) error {
		ListImages()
		return nil
	},
}

var rmiCommand = cli.Command{
	Name:  "rmi",
	Usage: "Remove s image e.g. tinydocker rmi nginx",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing image name")
		}
		imageName := context.Args().Get(0)
		removeImage(imageName)
		return nil
	},
}

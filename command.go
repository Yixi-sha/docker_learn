package main

import (
	"./cgroup/subsystem"
	"./container"
	"./nsenter"
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `create a container with namespace and cgroup limit
			mydocker run -ti [command]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		&cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		&cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		&cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		&cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
	},

	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing container command")
		}
		var cmdArray []string
		for i := 0; i < context.NArg(); i++ {
			cmdArray = append(cmdArray, context.Args().Get(i))
		}
		tty := context.Bool("ti")
		detach := context.Bool("d")
		if tty && detach {
			log.Fatal("-ti and -d can not both provided")
		}
		resConf := &subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		volume := context.String("v")
		containerName := context.String("name")
		Run(tty, cmdArray, resConf, volume, containerName)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: `Init contanier process run user's process in container. Do not call it outside`,

	Action: func(context *cli.Context) error {
		log.Println("init come on")
		cmd := context.Args().Get(0)
		log.Printf("command %s\n", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}

var commmitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",

	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing container command")
		}
		imageName := context.Args().Get(0)
		commmitContainer(imageName)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the container",
	Action: func(context *cli.Context) error {
		fmt.Println(container.ListContainers())
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "log",
	Usage: "print logs of container",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing container name")
		}
		containerName := context.Args().Get(0)
		content, err := container.GetLogContainer(containerName)
		if err != nil {
			return err
		}
		fmt.Println(content)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "into container exec a command",
	Action: func(context *cli.Context) error {
		if os.Getenv(nsenter.ENV_EXEC_PID) != "" {
			log.Println("pid callback pid ", os.Getgid())
			return nil
		}
		if context.NArg() < 2 {
			return fmt.Errorf("missing container name or command")
		}
		containerName := context.Args().Get(0)
		var commandArray []string
		for _, arg := range context.Args().Tail() {
			commandArray = append(commandArray, arg)
		}
		nsenter.ExecContainer(containerName, commandArray)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("miss container name")
		}
		containerName := context.Args().Get(0)
		container.StopContainer(containerName)
		return nil
	},
}

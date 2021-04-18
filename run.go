package main

import (
	"log"
	"github.com/urfave/cli"
	"os"
	"fmt"
	"./container"
	"./cgroup"
	"./cgroup/subsystem"
	"strings"
)

const usage = `mydocker is simple container`

var runCommand = cli.Command{
	Name: "run",
	Usage: `create a container with namespace and cgroup limit
			mydocker run -ti [command]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "ti",
			Usage: "enable tty",
		},
		&cli.StringFlag{
			Name: "m",
			Usage: "memory limit",
		},
		&cli.StringFlag{
			Name: "cpushare",
			Usage: "cpushare limit",
		},
		&cli.StringFlag{
			Name: "cpuset",
			Usage: "cpuset limit",
		},
	},
	
	Action: func(context *cli.Context) error{
		if context.NArg() < 1{
			return fmt.Errorf("Missing container command")
		}
		var cmdArray []string
		for i := 0; i < context.NArg(); i++  {
			cmdArray = append(cmdArray, context.Args().Get(i))
		}
		tty := context.Bool("ti")
		resConf := &subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet: context.String("cpuset"),
			CpuShare:context.String("cpushare"),
		}
		Run(tty, cmdArray, resConf)
		return nil
	},
}

var initCommand = cli.Command{
	Name: "init",
	Usage: `Init contanier process run user's process in container. Do not call it outside`,

	Action: func(context *cli.Context) error{
		log.Println("init come on")
		cmd := context.Args().Get(0)
		log.Printf("command %s\n", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}



func Run(tty bool, comArray []string, res *subsystem.ResourceConfig){
	parent, writePipe := container.NewParentProcess(tty)

	if parent == nil{
		log.Println("new parent process error")
		return
	}
	if err := parent.Start(); err != nil{
		log.Fatal(err)
	}
	log.Println(comArray,os.Getpid(),parent.Process.Pid)
	cgroupManager := cgroup.NewCgroupManager("mydocker")
	defer cgroupManager.Destroy()

	_ = cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(comArray, writePipe)
	parent.Wait()
	//os.Exit(0)
}

func sendInitCommand(comArray []string, writePipe *os.File){
	command := strings.Join(comArray, " ")
	log.Printf("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

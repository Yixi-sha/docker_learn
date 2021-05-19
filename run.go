package main

import (
	"./cgroup"
	"./cgroup/subsystem"
	"./container"
	"fmt"
	"log"
	"os"
	"strings"
)

const usage = `mydocker is simple container`

func Run(tty bool, comArray []string, res *subsystem.ResourceConfig, volume string, containerName string) {
	id := container.RandStringBytes(10)
	if containerName == "" {
		containerName = id
	}
	parent, writePipe, rootURL, mntURL, volume := container.NewParentProcess(tty, volume, containerName)

	if parent == nil {
		log.Println("new parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Fatal(err)
	}

	containerInfo, err := container.RecordContainerInfo(id, parent.Process.Pid, comArray, containerName, rootURL, mntURL, volume)
	if err != nil {
		fmt.Println(err)
	}

	log.Println(comArray, os.Getpid(), parent.Process.Pid)
	cgroupManager := cgroup.NewCgroupManager("mydocker"+containerName)

	_ = cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(comArray, writePipe)
	if tty {
		parent.Wait()
		container.DeleteWorkSpace(containerInfo.RootURL, containerInfo.MntURL, containerInfo.Volume, containerInfo.Name)
		container.DeleteContainerInfo(containerName)
		cgroupManager.Destroy()
	}

	//os.Exit(0)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Printf("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

package cgroup

import (
	"fmt"
	"log"

	"mydocker/cgroup/subsystem"
)

type CgroupManager struct {
	Path     string
	Resource *subsystem.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subsysIns := range subsystem.SubsystemsIns {
		subsysIns.Apply(c.Path, pid)
	}
	return nil
}

func (c *CgroupManager) Set(res *subsystem.ResourceConfig) error {
	for _, subsysIns := range subsystem.SubsystemsIns {
		subsysIns.Set(c.Path, res)
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	fmt.Println(":Destroy")
	for _, subsysIns := range subsystem.SubsystemsIns {
		if err := subsysIns.Remove(c.Path); err != nil {
			log.Println("remove cgroup fail %v", err)
		}
	}
	fmt.Println(":Destroy")
	return nil
}

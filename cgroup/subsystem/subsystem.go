package subsystem

type ResourceConfig struct{
	MemoryLimit string
	CpuShare string
	CpuSet string
}

type Subsystem interface {
	Name() string

	//set cgourp in system
	Set(path string, res *ResourceConfig) error

	//add cgroup
	Apply(path string, pid int) error

	//remove cgourp
	Remove(path string) error
}

var (
	SubsystemsIns = []Subsystem{
		&CpusetSubSystem{},
		&MemorySubsystem{},
		&CpuSubSystem{},
	}
)
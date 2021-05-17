package container

const RootURL string = "/home/winks/work/expore/docker/mydocker/test/"

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"createTime"`
	Status      string `json:"status"`
}

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/mydocker/%s"
	ConfigName          string = "config.json"
	LogFile             string = "LogFile.txt"
)

package container

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"createTime"`
	Status      string `json:"status"`
	RootURL     string `json:"rootURL"`
	MntURL      string `json:"mntURL"`
	Volume      string `json:"volume"`
}

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/mydocker/%s"
	ConfigName          string = "config.json"
	LogFile             string = "LogFile.txt"
)

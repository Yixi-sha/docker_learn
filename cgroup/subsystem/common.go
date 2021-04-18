package subsystem

import (
	"os"
	"bufio"
	"strings"
	"fmt"
	"path"
)

func FindCgourpMountpoint(subsystem string) string{
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil{
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan(){
		text := scanner.Text()
		fields := strings.Split(text, " ")
		for _, opt := range strings.Split(fields[len(fields) - 1], ","){
			if opt == subsystem{
				return fields[4]
			}
		}
		if err := scanner.Err(); err != nil{
			return ""
		}
	}
	return ""
}

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error){
	cgroupRoot := FindCgourpMountpoint(subsystem)
	path := path.Join(cgroupRoot, cgroupPath)
	if _, err := os.Stat(path); err == nil || (autoCreate && os.IsNotExist(err)){
		if os.IsNotExist(err){
			if err := os.Mkdir(path, 0755); err == nil{

			}else{
				return "", fmt.Errorf("error create cgourp %v", err)
			}
		}
		return path, nil
	}else{
		return  "", fmt.Errorf("cgroup path error %v", err)
	}
}
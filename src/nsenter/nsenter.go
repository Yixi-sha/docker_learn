package nsenter

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"mydocker/container"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func getEnvByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return strings.Split(string(contentBytes), "\u0000")
}

func ExecContainer(containerName string, commandArray []string) {
	pid, err := container.GetContainerPidByName(containerName)
	if err != nil {
		log.Printf("exec container name %s err %s\n", containerName, err)
		return
	}
	cmdstr := strings.Join(commandArray, " ")
	fmt.Println("container pid ", pid)
	fmt.Println("command: ", cmdstr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdstr)

	cmd.Env = append(os.Environ(), getEnvByPid(pid)...)

	if err := cmd.Run(); err != nil {
		log.Println(err)
	}
}

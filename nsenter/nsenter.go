package nsenter

import (
	"../container"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

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

	if err := cmd.Run(); err != nil {
		log.Println(err)
	}
}

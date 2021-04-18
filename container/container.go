package container

import (
	"os/exec"
	"os"
	"syscall"
	"log"
	"io/ioutil"
	"fmt"
	"strings"
	"path/filepath"
)

func NewParentProcess(tty bool) (*exec.Cmd, *os.File){
	readPipe, writePipe, err := NewPipe()
	if err != nil{
		log.Println("new pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET |
		syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Dir = "/home/winks/work/expore/docker/mydocker/busybox"
	cmd.ExtraFiles = []*os.File{readPipe} //keep open in file fd is 3 + i
	return cmd, writePipe

}

func NewPipe() (*os.File, *os.File, error){
	read, write, err := os.Pipe();
	if err != nil{
		return nil, nil, err
	}
	return read, write, nil
}

func readUserCommand() []string{
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil{
		log.Printf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func RunContainerInitProcess() error{

	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0{
		return fmt.Errorf("Run container get user command errorm cmdArray is nil")
	}
	
	setUpMount()
	
	path, err := exec.LookPath(cmdArray[0])
	if err != nil{
		log.Printf("exec loop path error %v", err)
		return nil
	}
	log.Printf("find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil{
		log.Fatal(err)
	}
	return nil
}


func pivotRoot(root string) error{
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil{
		return fmt.Errorf("Mount rootfs to itself error %v ", err)
	}
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil{
		return err
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil{
		return fmt.Errorf("pivotRoot %v %s %s", err, root, pivotDir)
	}

	if err := syscall.Chdir("/"); err != nil{
		return fmt.Errorf("syscall.Chdir %v ", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil{
		return fmt.Errorf("syscall.Unmount %v ", err)
	}
	return os.Remove(pivotDir)
}

func setUpMount(){
	pwd, err := os.Getwd()
	if err != nil{
		log.Fatal("os.Getwd() error %v", err)
	}
	log.Println("Current location is %s", pwd)
	if err := pivotRoot(pwd); err != nil{
		log.Fatal(err)
	}

	syscall.Mount("", "/", "", syscall.MS_REC | syscall.MS_PRIVATE, "")

	defaultMountFlag := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlag), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID | syscall.MS_STRICTATIME, "mode=755")
}
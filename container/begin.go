package container

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func NewParentProcess(tty bool, volume string, containerName string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
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
	} else {
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			log.Println(err)
			return nil, nil
		}
		stdLogFile, err := os.Create(dirURL + "/" + LogFile)
		if err != nil {
			log.Println(err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}

	cmd.ExtraFiles = []*os.File{readPipe} //keep open in file fd is 3 + i
	mntURL := RootURL + "mnt/"
	NewWorkSpace(RootURL, mntURL, volume)
	cmd.Dir = mntURL
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Printf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func RunContainerInitProcess() error {

	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command errorm cmdArray is nil")
	}

	setUpMount()

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Printf("exec loop path error %v", err)
		return nil
	}
	log.Printf("find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Fatal(err)
	}
	return nil
}

func pivotRoot(root string) error {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error %v ", err)
	}
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivotRoot %v %s %s", err, root, pivotDir)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("syscall.Chdir %v ", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("syscall.Unmount %v ", err)
	}
	return os.Remove(pivotDir)
}

func setUpMount() {
	syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("os.Getwd() error %v", err)
	}
	log.Println("Current location is %s", pwd)
	if err := pivotRoot(pwd); err != nil {
		log.Fatal(err)
	}

	defaultMountFlag := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlag), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Printf("Fail to judge whether dir is %s exist, %v\n", busyboxURL, err)
	}
	if exist == false {
		busyboxTarURL := rootURL + "busybox.tar"
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Fatal("os.Mkdir err ", busyboxURL)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Fatal("exec.Command ", busyboxTarURL)
		}
	}
}

func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		if os.IsNotExist(err) {
			log.Fatal(err)
		}
	}
}

func CreateMountPoint(rootURL string, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		if os.IsNotExist(err) {
			log.Fatal(err)
		}
	}
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func MountVolume(mntURL string, volumeURLs []string) {
	parentURL := volumeURLs[0]

	if err := os.Mkdir(parentURL, 0777); err != nil {
		if os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	containerVolumeURL := mntURL + volumeURLs[1]
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		if os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	cmd := exec.Command("mount", "-t", "aufs", "-o", "dirs="+parentURL, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func NewWorkSpace(rootURL, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)

	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(mntURL, volumeURLs)
			log.Println("%q", volumeURLs)
		} else {
			log.Fatal("volume parameter err")
		}
	}
}

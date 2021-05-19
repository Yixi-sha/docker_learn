package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"../cgroup"
)

func DeleteWorkSpace(rootURL string, mntURL string, volume string, containerName string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteVolumeMount(mntURL, volumeURLs)
			log.Println("%q", volumeURLs)
		} else {
			log.Fatal("volume parameter err")
		}
	}
	DeleteMountPoint(mntURL)
	DeleteWriteLayer(rootURL, containerName)
	if err := os.RemoveAll(rootURL + containerName); err != nil {
		log.Fatal(err)
	} 
}

func DeleteVolumeMount(mntURL string, volumeURLs []string) {
	containerURL := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func DeleteMountPoint(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Fatal(err)
	}
}

func DeleteWriteLayer(rootURL, containerName  string) {
	writeURL := rootURL + "writeLayer/" + containerName
	if err := os.RemoveAll(writeURL); err != nil {
		log.Fatal(err)
	}
}

func StopContainer(containerName string) {
	info, err := GetContainerInfobyName(containerName)
	if err != nil {
		log.Println(err)
		return
	}

	pidInt, err := strconv.Atoi(info.Pid)

	if err != nil {
		log.Println(err)
		return
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Println(err)
		return
	}

	info.Status = STOP
	info.Pid = " "
	newContentBytes, err := json.Marshal(info)
	if err != nil {
		log.Println(err)
		return
	}

	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := ioutil.WriteFile(dirURL+"/"+ConfigName, newContentBytes, 0622); err != nil {
		log.Println(err)
	}
	fmt.Println("end")
}

func RemoveContainer(containerName string){
	containerInfo, err := GetContainerInfobyName(containerName)
	if err != nil{
		log.Println(err)
		return
	}

	if containerInfo.Status != STOP{
		log.Println("state is not ", STOP)
		return
	}
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil{
		log.Println(err)
		return
	}
	DeleteWorkSpace(containerInfo.RootURL, containerInfo.MntURL, containerInfo.Volume, containerInfo.Name)
	cgroupManager := cgroup.NewCgroupManager("mydocker"+containerInfo.Name)
	cgroupManager.Destroy()
}

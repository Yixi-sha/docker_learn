package main

import (
	"fmt"
	"log"
	"os/exec"

	"mydocker/container"
)

func commmitContainer(containerName string) {
	conatinerInfo, err := container.GetContainerInfobyName(containerName)
	if err != nil {
		log.Println(err)
		return
	}
	imageTar := conatinerInfo.RootURL + containerName + ".tar"
	fmt.Printf("%s\n", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", conatinerInfo.MntURL, ".").CombinedOutput(); err != nil {
		log.Fatal(err)
	}
}

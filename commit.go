package main

import (
	"./container"
	"fmt"
	"log"
	"os/exec"
)

func commmitContainer(imageName string) {
	mntURL := container.RootURL + "mnt/"
	imageTar := container.RootURL + imageName + ".tar"
	fmt.Printf("%s\n", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Fatal(err)
	}
}

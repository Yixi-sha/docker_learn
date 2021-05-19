package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

func RandStringBytes(n int) string { //confie to n
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RecordContainerInfo(id string, containerPid int, commandArray []string, containerName, rootURL, mntURL ,volume string) (*ContainerInfo, error) {
	createTime := time.Now().Format("2020-01-02 15:04:05")
	command := strings.Join(commandArray, "")

	containerInfo := ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPid),
		Command:     command,
		CreatedTime: createTime,
		Status:      RUNNING,
		Name:        containerName,
		RootURL:     rootURL,
		MntURL:      mntURL,
		Volume:      volume,
	}

	jsonByte, err := json.Marshal(containerInfo)
	if err != nil {
		return nil, err
	}
	jsonStr := string(jsonByte)

	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		return nil, err
	}

	fileName := dirUrl + "/" + ConfigName

	file, err := os.Create(fileName)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	if _, err := file.WriteString(jsonStr); err != nil {
		return nil, err
	}

	return &containerInfo, nil
}

func DeleteContainerInfo(containerName string) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		fmt.Println(err)
	}
}

func getContainerInfo(file os.FileInfo) (*ContainerInfo, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(DefaultInfoLocation, containerName)
	configFileDir = configFileDir + "/" + ConfigName

	content, err := ioutil.ReadFile(configFileDir)

	if err != nil {
		return nil, err
	}

	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		return nil, err
	}

	return &containerInfo, nil
}

func ListContainers() string {
	dirURL := DefaultInfoLocation[:len(DefaultInfoLocation)-3]

	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var containerInfos []*ContainerInfo

	for _, file := range files {
		tmpContainerInfo, err := getContainerInfo(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		containerInfos = append(containerInfos, tmpContainerInfo)
	}
	var builder strings.Builder
	w := tabwriter.NewWriter(&builder, 12, 1, 3, ' ', 0)

	fmt.Fprint(w, "ID\tName\tPID\tStatus\tCommand\tCreated\n")
	for _, item := range containerInfos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id, item.Pid, item.Name, item.Status, item.Command, item.CreatedTime)
	}
	if err := w.Flush(); err != nil {
		fmt.Println(err)
	}
	return builder.String()

}

func GetLogContainer(containerName string) (string, error) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	file, err := os.Open(dirURL + "/" + LogFile)
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GetContainerInfobyName(containerName string) (*ContainerInfo, error) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	contentBytes, err := ioutil.ReadFile(dirURL + "/" + ConfigName)
	if err != nil {
		return nil, err
	}

	var containerInfo ContainerInfo

	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}

func GetContainerPidByName(containerName string) (string, error) {
	containerInfo, err := GetContainerInfobyName(containerName)
	if err != nil {
		return "", err
	}

	return containerInfo.Pid, nil
}

package network

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/mydocker/ipam/subset.json"

type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (this *IPAM) load() error {
	if _, err := os.Stat(this.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	subnetConfig, err := os.Open(this.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	defer subnetConfig.Close()

	subnetJson, err := ioutil.ReadAll(subnetConfig)
	if err != nil {
		return err
	}

	err = json.Unmarshal(subnetJson, this.Subnets)

	return err
}

func (this *IPAM) dump() error {

	ipamConfigDir, _ := path.Split(this.SubnetAllocatorPath)
	if _, err := os.Stat(this.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigDir, 0644)
		} else {

			return err
		}
	}

	ipamConfig, err := os.OpenFile(this.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer ipamConfig.Close()

	ipamConfigJson, err := json.Marshal(this.Subnets)
	if err != nil {
		return err
	}

	_, err = ipamConfig.Write(ipamConfigJson)

	return err
}

func (this *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	this.Subnets = &map[string]string{}
	if err = this.load(); err != nil {
		return
	}
	one, size := subnet.Mask.Size()

	if _, exist := (*this.Subnets)[subnet.String()]; !exist {
		(*this.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}
	for c := range (*this.Subnets)[subnet.String()] {
		if (*this.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*this.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*this.Subnets)[subnet.String()] = string(ipalloc)

			ip = subnet.IP

			for t := int(4); t > 0; t-- {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}
	this.dump()
	return
}

func (this *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	this.Subnets = &map[string]string{}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	err := this.load()
	if err != nil {
		return err
	}

	c := 0

	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := int(4); t > 0; t-- {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}
	ipalloc := []byte((*this.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*this.Subnets)[subnet.String()] = string(ipalloc)
	this.dump()
	return nil
}

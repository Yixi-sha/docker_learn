package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"text/tabwriter"
)

type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}

var (
	defaultNetworkPath = "/var/run/mydocker/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

func (this *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}

	nwPath := path.Join(dumpPath, this.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return err
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(this)
	if err != nil {
		return err
	}

	_, err = nwFile.Write(nwJson)
	return err
}

func (this *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer nwConfigFile.Close()

	jsonByte, err := ioutil.ReadAll(nwConfigFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonByte, this)
	return err
}

func (this *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, this.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, this.Name))
	}
}

func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	fmt.Println(driver, subnet, name)
	gatewayIp, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIp
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(defaultNetworkPath)
}

/*func Connect(networkName string, cinfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("No Such Network %s", networkName)
	}

	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}

	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}

	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}

	return configPortMapping(ep, cinfo)
}*/

func Init() error {
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}
		if err := nw.load(nwPath); err != nil {
			return err
		}
		networks[nwName] = nw
		return nil
	})
	return nil
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")

	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange.String(), nw.Driver)
	}
	if err := w.Flush(); err != nil {
		fmt.Println(err)
	}
}

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	fmt.Println("t", nw.IpRange, nw.IpRange.IP)
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return err
	}
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return err
	}

	return nw.remove(defaultNetworkPath)
}

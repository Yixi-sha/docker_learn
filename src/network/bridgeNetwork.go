package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func (this *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip

	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  this.Name(),
	}

	err := this.initBrige(n)
	if err != nil {
		return nil, err
	}
	return n, err
}

func createBridgeinterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	bridge := &netlink.Bridge{
		LinkAttrs: la,
	}

	if err := netlink.LinkAdd(bridge); err != nil {
		return fmt.Errorf("could not add %s: %v\n", la.Name, err)
	}
	return nil
}

func setinterfaceIP(name string, rawIP string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	ipNet, err := netlink.ParseAddr(rawIP)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(iface, ipNet)
}

func setinterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(iface)
	return err
}

func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	_, err := cmd.Output()
	fmt.Println(err)
	return err
}

func (this *BridgeNetworkDriver) initBrige(n *Network) error {

	if err := createBridgeinterface(n.Name); err != nil {
		return err
	}

	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP

	if err := setinterfaceIP(n.Name, gatewayIP.String()); err != nil {
		return err
	}

	if err := setinterfaceUP(n.Name); err != nil {
		return err
	}

	err := setupIPTables(n.Name, n.IpRange)

	return err
}

func (this *BridgeNetworkDriver) Delete(network Network) error {

	br, err := netlink.LinkByName(network.Name)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

func (this *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (this *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	return nil
}

func (this *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	return nil
}

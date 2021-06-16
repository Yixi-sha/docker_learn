package network

import (
	"fmt"
	"net"
	"testing"
)

func TestAllocate(t *testing.T) {
	for i := 0; i < 10; i++ {
		_, ipnet, _ := net.ParseCIDR("192.168.0.0/24")
		ip, _ := ipAllocator.Allocate(ipnet)
		fmt.Printf("TestAllocate ip: %v\n", ip)
	}

}

func TestRelease(t *testing.T) {
	ip, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	ipAllocator.Release(ipnet, &ip)
	//t.Logf("TestRelease ip: %v", ip)
}

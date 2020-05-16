package hsocks5

import (
	"net"
	"strings"
)

// IPList type
type IPList struct {
	nets []*net.IPNet
}

// LoadIPListFrom assets
func LoadIPListFrom(path string) *IPList {
	l, _ := Asset(path)
	return NewIPList(string(l))
}

// NewIPList object
func NewIPList(cidrs string) *IPList {
	nets := []*net.IPNet{}

	for _, cidr := range strings.Split(cidrs, "\n") {
		if _, net, err := net.ParseCIDR(cidr); err == nil {
			nets = append(nets, net)
		}
	}

	return &IPList{nets: nets}
}

// ContainsIP value
func (l *IPList) ContainsIP(ip net.IP) (rt bool) {
	rt = false
	for _, subNet := range l.nets {
		if subNet.Contains(ip) {
			rt = true
			break // quick break
		}
	}
	return
}

// Contains ip in IPList
func (l *IPList) Contains(ipOrHostname string) (rt bool) {
	rt = false
	if oIP := net.ParseIP(ipOrHostname); oIP != nil { // is ip
		rt = l.ContainsIP(oIP)
	} else {
		if oIPs, err := net.LookupIP(ipOrHostname); err == nil {
			rt = l.MultiContainsIPs(oIPs)
		}
	}
	return
}

// MultiContainsIPs value
func (l *IPList) MultiContainsIPs(ips []net.IP) (rt bool) {
	rt = false
	for _, subNet := range l.nets {
		for _, oIP := range ips {
			if subNet.Contains(oIP) {
				rt = true
				break
			}
		}
		if rt {
			break
		}
	}
	return
}

// MultiContains any on any in IPList
func (l *IPList) MultiContains(ipOrHostnames []string) (rt bool) {
	rt = false

	for _, ip := range ipOrHostnames {
		if l.Contains(ip) {
			rt = true
			break // quick break
		}
	}

	return
}

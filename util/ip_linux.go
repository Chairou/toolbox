package util

import (
"net"
"strings"
"syscall"
)

// GetInternalIPv4Address 获取内部IPv4地址
func GetInternalIPv4Address() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	var buf syscall.Utsname
	err = syscall.Uname(&buf)
	if err != nil {
		return "127.0.0.1"
	}

	release := charsToString(buf.Release[:])
	isTlinux := strings.Contains(release, "tlinux")

	for _, iface := range ifaces {
		if isTlinux && iface.Name != "eth1" {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipaddr, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			if ipaddr.IsLoopback() {
				continue
			}
			if ipaddr.To4() != nil {
				return ipaddr.String()
			}
		}
	}
	return "127.0.0.1"
}

func charsToString(ca []int8) string {
	s := make([]byte, len(ca))
	var lens int
	for ; lens < len(ca); lens++ {
		if ca[lens] == 0 {
			break
		}
		s[lens] = uint8(ca[lens])
	}
	return string(s[0:lens])
}

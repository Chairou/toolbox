package util

// GetInternalIPv4Address 获取内部IPv4地址
//func GetInternalIPv4Address() string {
//	addrs, err := net.InterfaceAddrs()
//	if err != nil {
//		return "127.0.0.1"
//	}
//	for _, addr := range addrs {
//		ipaddr, _, err := net.ParseCIDR(addr.String())
//		if err != nil {
//			continue
//		}
//		if ipaddr.IsLoopback() {
//			continue
//		}
//		if ipaddr.To4() != nil {
//			return ipaddr.String()
//		}
//	}
//	return "127.0.0.1"
//}

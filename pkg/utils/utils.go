package utils

import (
	"net"
	"strings"
)

func Ternary[V interface{}](condition bool, trueVal, falseVal V) V {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

func IsContain[V comparable](list []V, target V) (exists bool, index int) {
	for i, v := range list {
		if v == target {
			return true, i
		}
	}
	return false, -1
}

func GetTargetIP(prefix string) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				sIP := ipnet.IP.String()
				if strings.HasPrefix(sIP, prefix) {
					ipnet.IP[15] = 51
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}

package utils

import (
	"log"
	"net"
)

func GetLocalIP() (ip string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal("net.Interfaces failed, err:", err.Error())
		return
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ip = ipnet.IP.String()
					}
				}
			}
		}
	}
	return

}

package main

import (
	"net"
	"net/http"
	"strings"
)

// status returns server IP information.
func status(w http.ResponseWriter, r *http.Request) {
	primaryIP := ""
	// Determine primary IP via UDP dial.
	conn, err := net.Dial("udp", "192.0.2.1:80")
	if err == nil {
		if local, ok := conn.LocalAddr().(*net.UDPAddr); ok && local.IP.To4() != nil {
			primaryIP = local.IP.To4().String()
		}
		_ = conn.Close()
	}
	if primaryIP == "" {
		interfaces, _ := net.Interfaces()
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					candidate := ipnet.IP.To4().String()
					if strings.HasPrefix(candidate, "169.254.") || strings.HasPrefix(candidate, "192.168.56.") {
						continue
					}
					primaryIP = candidate
					break
				}
			}
			if primaryIP != "" {
				break
			}
		}
	}
	ips := []string{}
	if primaryIP != "" {
		ips = append(ips, primaryIP)
	}
	jsonResponse(w, map[string]any{"port": port, "ip": primaryIP, "ips": ips, "homepage": home})
}

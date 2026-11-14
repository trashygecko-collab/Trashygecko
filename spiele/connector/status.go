package main

import (
	"net"
	"net/http"
	"strings"
)

// localIPv4s sammelt alle nutzbaren IPv4-Adressen dieses PCs, die primäre zuerst.
func localIPv4s() []string {
	primary := ""
	// Determine primary IP via UDP dial.
	if conn, err := net.Dial("udp", "192.0.2.1:80"); err == nil {
		if local, ok := conn.LocalAddr().(*net.UDPAddr); ok && local.IP.To4() != nil {
			primary = local.IP.To4().String()
		}
		_ = conn.Close()
	}

	ips := []string{}
	if primary != "" {
		ips = append(ips, primary)
	}

	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			candidate := ipnet.IP.To4().String()
			if strings.HasPrefix(candidate, "169.254.") || strings.HasPrefix(candidate, "192.168.56.") {
				continue
			}
			duplicate := false
			for _, known := range ips {
				if known == candidate {
					duplicate = true
					break
				}
			}
			if !duplicate {
				ips = append(ips, candidate)
			}
		}
	}
	return ips
}

// status returns server IP information.
func status(w http.ResponseWriter, r *http.Request) {
	ips := localIPv4s()
	primary := ""
	if len(ips) > 0 {
		primary = ips[0]
	}
	// url ist die Adresse, die Mitspieler eintippen sollen. Lokal ist das die
	// LAN-Adresse; die Online-Fassung setzt hier ihre öffentliche Adresse ein.
	url := ""
	if primary != "" {
		url = "http://" + primary + ":" + port + "/connector"
	}
	jsonResponse(w, map[string]any{
		"port":     port,
		"ip":       primary,
		"ips":      ips,
		"url":      url,
		"homepage": home,
		"online":   false,
	})
}

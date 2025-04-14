package common

import (
	"fmt"
	"net"
	"time"
)

const (
	// DefaultWorkDirPath is the base path for all box-related directories
	DefaultWorkDirPath = "/var/gbox"
	// DefaultShareDirPath is the path for the shared directory within a box
	DefaultShareDirPath = DefaultWorkDirPath + "/share"
)

// GetLocalIPs returns all local IP addresses
func GetLocalIPs() []string {
	var ips []string
	// Add localhost first
	ips = append(ips, "localhost", "127.0.0.1")

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	// Track if we've found a main IPv4 address
	foundMainIPv4 := false

	for _, i := range interfaces {
		// Skip loopback, down, and point-to-point interfaces
		if i.Flags&net.FlagLoopback != 0 || // Skip loopback
			i.Flags&net.FlagUp == 0 || // Skip down interfaces
			i.Flags&net.FlagPointToPoint != 0 { // Skip point-to-point interfaces
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				// Only add IPv4 addresses
				if ipnet.IP.To4() != nil {
					if !foundMainIPv4 {
						ips = append(ips, ipnet.IP.String())
						foundMainIPv4 = true
					}
				}
			}
		}
	}
	return ips
}

// FormatDurationConcise provides a simpler string representation for durations.
func FormatDurationConcise(d time.Duration) string {
	if d%(24*time.Hour) == 0 {
		h := d / (24 * time.Hour)
		if h > 0 {
			return fmt.Sprintf("%dd", h)
		} // Days
	}
	if d%time.Hour == 0 {
		h := d / time.Hour
		if h > 0 {
			return fmt.Sprintf("%dh", h)
		} // Hours
	}
	if d%time.Minute == 0 {
		m := d / time.Minute
		if m > 0 {
			return fmt.Sprintf("%dm", m)
		} // Minutes
	}
	if d%time.Second == 0 {
		s := d / time.Second
		if s > 0 {
			return fmt.Sprintf("%ds", s)
		} // Seconds
	}
	return d.String() // Fallback to default
}

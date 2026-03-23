package algorithm

import (
	"hllb/checks"
	"log"
	"net/netip"
)

// Round Robin

var ServerIndex int

func RR() (netip.Addr, error) {
	if len(checks.ValidPoolHost) == 0 {
		return netip.Addr{}, nil
	}

	ServerIndex++
	index := ServerIndex % len(checks.ValidPoolHost)
	valid, err := netip.ParseAddr(checks.ValidPoolHost[index])
	if err != nil {
		log.Printf("Failed to parse IP %s: %v", valid, err)
	}
	return valid, nil
}

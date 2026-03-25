package algorithm

import (
	"hllb/checks"
	"log"
	"net/netip"
	"sync/atomic"
)

// Round Robin

var serverIndex atomic.Int64

func RR() (netip.Addr, error) {
	pool := checks.GetValidPoolHost()
	if len(pool) == 0 {
		return netip.Addr{}, nil
	}

	idx := serverIndex.Add(1)
	index := int(idx) % len(pool)
	valid, err := netip.ParseAddr(pool[index])
	if err != nil {
		log.Printf("Failed to parse IP %s: %v", pool[index], err)
	}
	return valid, nil
}

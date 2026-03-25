package utils

import (
	"log"
	"os"
	"strings"
	"sync/atomic"

	"codeberg.org/miekg/dns"
)

type ZoneRecord struct {
	A  []string
	NS []string
}

var zonePtr atomic.Pointer[map[string]ZoneRecord]

func init() {
	empty := make(map[string]ZoneRecord)
	zonePtr.Store(&empty)
}

// GetZoneSnapshot возвращает текущую карту зон без копирования и без блокировки.
func GetZoneSnapshot() map[string]ZoneRecord {
	return *zonePtr.Load()
}

// GetZone возвращает запись из зоны по ключу без блокировки.
func GetZone(key string) (ZoneRecord, bool) {
	z := *zonePtr.Load()
	rec, ok := z[key]
	return rec, ok
}

func InitZone() {
	tmpZone := make(map[string]ZoneRecord)
	allFileZone, err := os.ReadDir("./zone")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range allFileZone {
		if file.IsDir() {
			continue
		}
		f, err := os.Open("./zone/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		zp := dns.NewZoneParser(f, file.Name(), file.Name())
		for rr, err := range zp.RRs() {
			if err != nil {
				log.Fatal(err)
			}
			if rr == nil {
				continue
			}

			name := strings.TrimSuffix(strings.ToLower(rr.Header().Name), ".")

			record := tmpZone[name]
			switch rr := rr.(type) {
			case *dns.A:
				record.A = append(record.A, rr.A.String())
			case *dns.NS:
				record.NS = append(record.NS, rr.Ns)
			}
			tmpZone[name] = record
		}
	}
	zonePtr.Store(&tmpZone)
}

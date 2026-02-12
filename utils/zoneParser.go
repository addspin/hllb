package utils

import (
	"log"
	"os"
	"strings"

	"codeberg.org/miekg/dns"
)

type ZoneRecord struct {
	A  []string
	NS []string
}

var Zone = make(map[string]ZoneRecord)

func InitZone() map[string]ZoneRecord {
	tmpZone := make(map[string]ZoneRecord)
	f, err := os.Open("./zone/test.ru")
	if err != nil {
		log.Fatal(err)
	}

	zp := dns.NewZoneParser(f, "test.ru", "test.ru")
	for rr, err := range zp.RRs() {
		if err != nil {
			log.Fatal(err)
		}
		if rr == nil { // <‑‑ protect against nil RR
			continue
		}
		name := strings.TrimSuffix(strings.ToLower(rr.Header().Name), ".")

		record := tmpZone[name] // Получаем текущую структуру
		switch rr := rr.(type) {
		case *dns.A:
			record.A = append(record.A, rr.A.String())
		case *dns.NS:
			record.NS = append(record.NS, rr.Ns)

		}
		tmpZone[name] = record // После изменений в норвой переменной, записываем обратно в структуру
	}
	Zone = tmpZone
	return Zone
}

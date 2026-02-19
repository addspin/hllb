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
	allFileZone, err := os.ReadDir("./zone")
	if err != nil {
		log.Fatal(err)
	}
	// Получаем список файлов зон
	for _, file := range allFileZone {
		if file.IsDir() { // Пропускаем поддиректории
			continue
		}
		f, err := os.Open("./zone/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		zp := dns.NewZoneParser(f, file.Name(), file.Name()) // 1) origin подстановка к записям как корневого домена 2) значение в ошибках при парсингах
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
			tmpZone[name] = record // После изменений в новой переменной, записываем обратно в структуру
		}
	}
	Zone = tmpZone
	return Zone
}

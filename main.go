package main

import (
	"context"
	"hllb/utils"
	"io"
	"log"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
)

const (
	addr = ":1053"
)

type zoneRecord struct {
	A  []string
	NS []string
}

var zoneMutex sync.RWMutex
var zone map[string]zoneRecord

func initZone() map[string]zoneRecord {
	zone := make(map[string]zoneRecord)
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
		record := zone[name] // Получаем текущую структуру
		switch rr := rr.(type) {
		case *dns.A:
			record.A = append(record.A, rr.A.String())
		case *dns.NS:
			record.NS = append(record.NS, rr.Ns)

		}
		zone[name] = record // После изменений в норвой переменной, записываем обратно в структуру
	}
	return zone
}

// func getFileHash(path string) (string, error) {
// 	f, err := os.Open(path)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer f.Close()

// 	h := sha256.New()
// 	if _, err := io.Copy(h, f); err != nil {
// 		return "", err
// 	}

// 	return hex.EncodeToString(h.Sum(nil)), nil
// }

func watchZoneFile(path string, interval time.Duration) {
	// Запоминаем начальный хеш
	lastHash, _ := utils.GetHash(path)

	ticker := time.NewTicker(interval)
	for range ticker.C {
		currentHash, err := utils.GetHash(path)
		if err != nil {
			continue
		}

		if currentHash != lastHash {
			log.Printf("Хеш изменился [%s], обновляю зону...", currentHash[:8])

			newZone := initZone()

			zoneMutex.Lock()
			zone = newZone
			zoneMutex.Unlock()

			lastHash = currentHash
		}
	}
}

func main() {
	zone = initZone()

	go watchZoneFile("./zone/test.ru", 5*time.Second)

	dns.HandleFunc(".", handleDNS)
	go func() {
		if err := dns.ListenAndServe(addr, "udp", nil); err != nil {
			log.Fatalf("DNS UDP: %v", err)
		}
	}()
	if err := dns.ListenAndServe(addr, "tcp", nil); err != nil {
		log.Fatalf("DNS TCP: %v", err)
	}

}

func handleDNS(ctx context.Context, w dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) == 0 {
		return
	}
	q := req.Question[0]
	qyeryDomainName := q.Header().Name
	qtype := dns.RRToType(q)
	log.Println("qtype:", qtype)

	resp := new(dns.Msg)
	resp.ID = req.ID
	resp.Response = true
	resp.Question = req.Question
	resp.Authoritative = true

	queryNormDomainName := strings.TrimSuffix(strings.ToLower(qyeryDomainName), ".")
	for key, value := range zone {
		log.Println("key:", key, "value:", value)
	}

	if data, ok := zone[queryNormDomainName]; ok {
		switch qtype {
		case dns.TypeA:
			for _, ipData := range data.A {
				ip, _ := netip.ParseAddr(ipData)
				replyName := qyeryDomainName
				if !strings.HasSuffix(replyName, ".") {
					replyName += "."
				}
				a := &dns.A{
					Hdr: dns.Header{Name: replyName, Class: dns.ClassINET, TTL: 3600},
					A:   rdata.A{Addr: ip},
				}
				resp.Answer = append(resp.Answer, a)
			}
		case dns.TypeNS:
			for _, nsDomain := range data.NS {
				log.Println("nsDomain:", nsDomain)
				ns := &dns.NS{
					Hdr: dns.Header{Name: nsDomain, Class: dns.ClassINET, TTL: 3600},
					NS:  rdata.NS{Ns: nsDomain},
				}
				resp.Answer = append(resp.Answer, ns)
			}
		default:
			resp.Rcode = dns.RcodeNameError
		}

		if err := resp.Pack(); err != nil {
			return
		}
		if _, err := io.Copy(w, resp); err != nil {
			log.Printf("DNS write: %v", err)
		}
	} else {
		resp.Rcode = dns.RcodeNameError
		resp.Pack()
		io.Copy(w, resp)
	}
}

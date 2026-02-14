package handles

import (
	"context"
	"hllb/utils"
	"io"
	"log"
	"net/netip"
	"strings"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
)

var inZoneExists bool
var testWildcard bool
var values []string

func HandleDNS(ctx context.Context, w dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) == 0 {
		return
	}

	q := req.Question[0]
	queryDomainName := q.Header().Name
	queryType := dns.RRToType(q)
	log.Println("queryType:", queryType)

	resp := new(dns.Msg)
	resp.ID = req.ID
	resp.Response = true
	resp.Question = req.Question
	resp.Authoritative = true

	queryNormDomainName := strings.TrimSuffix(strings.ToLower(queryDomainName), ".")

	// проверяем wildcard или нет ff.test.ru
	testQueryWildcard := strings.Split(queryNormDomainName, ".")
	rootTest := strings.Join(testQueryWildcard[1:], ".")
	// testWildcard := strings.HasPrefix(queryNormDomainName, "*")

	log.Println("TEST rootTest", rootTest)
	log.Println("TEST testQueryWildcard", testQueryWildcard)

	// Если в зоне есть запись с wildcard например *.info
	test := utils.Zone
	for t := range test {
		testWildcard = strings.HasPrefix(t, "*.")
		if testWildcard {
			v := strings.Split(t, ".")
			values = v
			if values[1] == testQueryWildcard[1] {
				data := utils.Zone["*."+rootTest]
				inZoneExists = true
				log.Println("data:", data)
				switch queryType {
				case dns.TypeA:
					log.Println("data A:", data.A)
					for _, ipData := range data.A {
						log.Println("domain wildcard in zone *.info:", ipData)
						ip, _ := netip.ParseAddr(ipData)
						replyName := queryDomainName
						if !strings.HasSuffix(replyName, ".") {
							replyName += "."
						}
						a := &dns.A{
							Hdr: dns.Header{Name: replyName, Class: dns.ClassINET, TTL: 3600},
							A:   rdata.A{Addr: ip},
						}
						resp.Answer = append(resp.Answer, a)
						resp.Rcode = dns.RcodeSuccess
						if _, err := io.Copy(w, resp); err != nil {
							log.Printf("DNS write: %v", err)
						}

					}
				case dns.TypeNS:
					for _, nsDomain := range utils.Zone["test.ru"].NS {
						log.Println("nsdomain wildcard in zone *.info:", nsDomain)
						ns := &dns.NS{
							Hdr: dns.Header{Name: nsDomain, Class: dns.ClassINET, TTL: 3600},
							NS:  rdata.NS{Ns: nsDomain},
						}
						resp.Answer = append(resp.Answer, ns)
						resp.Rcode = dns.RcodeSuccess
						if _, err := io.Copy(w, resp); err != nil {
							log.Printf("DNS write: %v", err)
						}
					}
				}
				resp.Rcode = dns.RcodeNameError
				if err := resp.Pack(); err != nil {
					return
				}
				if _, err := io.Copy(w, resp); err != nil {
					log.Printf("DNS write: %v", err)
				}
			}
		}
	}

	// Если в зоне есть значение котрое запрашивает пользователь, отдаем
	if data, ok := utils.Zone[queryNormDomainName]; ok {
		inZoneExists = true
		log.Println("data:", data)
		switch queryType {
		case dns.TypeA:
			log.Println("data A:", data.A)
			for _, ipData := range data.A {

				log.Println("domain not wildcard:", ipData)
				ip, _ := netip.ParseAddr(ipData)
				replyName := queryDomainName
				if !strings.HasSuffix(replyName, ".") {
					replyName += "."
				}
				a := &dns.A{
					Hdr: dns.Header{Name: replyName, Class: dns.ClassINET, TTL: 3600},
					A:   rdata.A{Addr: ip},
				}
				resp.Answer = append(resp.Answer, a)
				resp.Rcode = dns.RcodeSuccess
				if _, err := io.Copy(w, resp); err != nil {
					log.Printf("DNS write: %v", err)
				}
			}
		case dns.TypeNS:
			if len(testQueryWildcard) > 2 {
				for _, nsDomain := range utils.Zone["test.ru"].NS {
					ns := &dns.NS{
						Hdr: dns.Header{Name: nsDomain, Class: dns.ClassINET, TTL: 3600},
						NS:  rdata.NS{Ns: nsDomain},
					}
					resp.Answer = append(resp.Answer, ns)
					resp.Rcode = dns.RcodeSuccess
					if _, err := io.Copy(w, resp); err != nil {
						log.Printf("DNS write: %v", err)
					}
				}
			} else {
				log.Println("data.NS:", data.NS)
				for _, nsDomain := range data.NS {
					log.Println("nsDomain not wildcard:", nsDomain)
					ns := &dns.NS{
						Hdr: dns.Header{Name: nsDomain, Class: dns.ClassINET, TTL: 3600},
						NS:  rdata.NS{Ns: nsDomain},
					}
					resp.Answer = append(resp.Answer, ns)
					resp.Rcode = dns.RcodeSuccess
					if _, err := io.Copy(w, resp); err != nil {
						log.Printf("DNS write: %v", err)
					}
				}
			}
		}
		resp.Rcode = dns.RcodeNameError
		if err := resp.Pack(); err != nil {
			return
		}
		if _, err := io.Copy(w, resp); err != nil {
			log.Printf("DNS write: %v", err)
		}
	}

	// Если передается вилдкард запись, и в зоне есть запись *
	if len(testQueryWildcard) > 2 && !inZoneExists { //значит вилдкард
		// проверем вилдкард запись в файле зоны *
		if _, ok := utils.Zone["*.test.ru"]; !ok {
			resp.Rcode = dns.RcodeNameError
			if err := resp.Pack(); err != nil {
				log.Printf("Pack error: %v", err)
			}
			if _, err := io.Copy(w, resp); err != nil {
				log.Printf("DNS write: %v", err)
			}
			log.Println("В зоне нет *")
			return
		}
		// если корневой домен верный и есть в зоне, тогда
		if data, ok := utils.Zone[rootTest]; ok {
			switch queryType {
			case dns.TypeA:
				for _, ipData := range data.A {
					log.Println("ADomain Wildcard:", ipData)
					ip, _ := netip.ParseAddr(ipData)
					replyName := queryDomainName
					if !strings.HasSuffix(replyName, ".") {
						replyName += "."
					}
					a := &dns.A{
						Hdr: dns.Header{Name: replyName, Class: dns.ClassINET, TTL: 3600},
						A:   rdata.A{Addr: ip},
					}
					resp.Answer = append(resp.Answer, a)
					resp.Rcode = dns.RcodeSuccess
					if _, err := io.Copy(w, resp); err != nil {
						log.Printf("DNS write: %v", err)
					}
				}
			case dns.TypeNS:
				for _, nsDomain := range utils.Zone["test.ru"].NS {
					log.Println("nsDomain Wildcard:", nsDomain)
					ns := &dns.NS{
						Hdr: dns.Header{Name: nsDomain, Class: dns.ClassINET, TTL: 3600},
						NS:  rdata.NS{Ns: nsDomain},
					}
					resp.Answer = append(resp.Answer, ns)
					resp.Rcode = dns.RcodeSuccess
					if _, err := io.Copy(w, resp); err != nil {
						log.Printf("DNS write: %v", err)
					}
				}
			}
			if err := resp.Pack(); err != nil {
				return
			}
			if _, err := io.Copy(w, resp); err != nil {
				log.Printf("DNS write: %v", err)
			}
		} else {
			resp.Rcode = dns.RcodeNameError
			if err := resp.Pack(); err != nil {
				log.Printf("Pack error: %v", err)
			}
			if _, err := io.Copy(w, resp); err != nil {
				log.Printf("DNS write: %v", err)
			}
			return
		}
	}

	resp.Rcode = dns.RcodeNameError
	if err := resp.Pack(); err != nil {
		log.Printf("Pack error: %v", err)
	}
	if _, err := io.Copy(w, resp); err != nil {
		log.Printf("DNS write: %v", err)
	}
}

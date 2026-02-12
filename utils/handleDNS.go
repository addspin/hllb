package utils

import (
	"context"
	"io"
	"log"
	"net/netip"
	"strings"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
)

func HandleDNS(ctx context.Context, w dns.ResponseWriter, req *dns.Msg) {
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

	// проверем вилдкард или нет ff.test.ru
	testWildcard := strings.Split(queryNormDomainName, ".")
	rootTest := strings.Join(testWildcard[1:], ".")
	log.Println("TEST testWildcard", testWildcard)

	if len(testWildcard) > 2 {
		// проверем вилдкард запись в файле зоны *
		if _, ok := Zone["*.test.ru"]; !ok {
			resp.Rcode = dns.RcodeNameError
			resp.Pack()
			io.Copy(w, resp)
		}
		if rootTest == queryNormDomainName {

		}

	}

	if data, ok := Zone[queryNormDomainName]; ok {
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
				log.Println("nsDomain not wildcard:", nsDomain)
				ns := &dns.NS{
					Hdr: dns.Header{Name: nsDomain, Class: dns.ClassINET, TTL: 3600},
					NS:  rdata.NS{Ns: nsDomain},
				}
				resp.Answer = append(resp.Answer, ns)
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

	// test no exist *
	if _, ok := Zone["*.test.ru"]; !ok {
		resp.Rcode = dns.RcodeNameError
		resp.Pack()
		io.Copy(w, resp)
	}

	if data, ok := Zone["*.test.ru"]; ok {
		// nsData := Zone["test.ru"]
		// log.Println("nsData", nsData)
		switch qtype {
		case dns.TypeA:
			for _, ipData := range data.A {
				log.Println("ADomain:", ipData)
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
			for _, nsDomain := range Zone["test.ru"].NS {
				log.Println("nsDomain:", nsDomain)
				ns := &dns.NS{
					Hdr: dns.Header{Name: nsDomain, Class: dns.ClassINET, TTL: 3600},
					NS:  rdata.NS{Ns: nsDomain},
				}
				resp.Answer = append(resp.Answer, ns)
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
		resp.Pack()
		io.Copy(w, resp)
	}
}

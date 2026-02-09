package main

import (
	"context"
	"io"
	"log"
	"net/netip"
	"os"
	"strings"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
)

const (
	addr          = ":53"
	resolveDomain = "test.ru."
	resolveIP     = "10.13.1.34"
)

func initZone() map[string]dns.RR {
	zone := make(map[string]dns.RR)
	f, err := os.Open("./zone/test.ru")
	if err != nil {
		log.Fatal(err)
	}
	zp := dns.NewZoneParser(f, "test.ru", "test.ru")
	for rr, err := range zp.RRs() {
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("%s\n", rr)
		// zone[rr.Header().Name] = rr
		log.Printf("zone[%d]", rr.Header().TTL)
	}
	return zone
}

func main() {

	zone := initZone()
	for dname := range zone {
		// nameNorm := strings.TrimSuffix(strings.ToLower(resolveDomain), ".")
		domainNorm := strings.TrimSuffix(strings.ToLower(dname), ".")
		log.Printf("domainNorm: %s", domainNorm)
	}
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
	name := q.Header().Name
	qtype := dns.RRToType(q)

	nameNorm := strings.TrimSuffix(strings.ToLower(name), ".")
	domainNorm := strings.TrimSuffix(strings.ToLower(resolveDomain), ".")
	log.Printf("nameNorm: %s, domainNorm: %s", nameNorm, domainNorm)

	resp := new(dns.Msg)
	resp.ID = req.ID
	resp.Response = true
	resp.Question = req.Question
	resp.Authoritative = true

	if qtype == dns.TypeA && nameNorm == domainNorm {
		ip, _ := netip.ParseAddr(resolveIP)
		replyName := name
		if !strings.HasSuffix(replyName, ".") {
			replyName += "."
		}
		a := &dns.A{
			Hdr: dns.Header{Name: replyName, Class: dns.ClassINET, TTL: 3600},
			A:   rdata.A{Addr: ip},
		}
		resp.Answer = append(resp.Answer, a)
	} else {
		resp.Rcode = dns.RcodeNameError
	}

	if err := resp.Pack(); err != nil {
		return
	}
	if _, err := io.Copy(w, resp); err != nil {
		log.Printf("DNS write: %v", err)
	}
}

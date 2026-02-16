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

const defaultTTL = 3600

func HandleDNS(ctx context.Context, w dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) == 0 {
		return
	}

	q := req.Question[0]
	queryDomainName := q.Header().Name
	queryType := dns.RRToType(q)

	log.Printf("DNS query: %s, type: %s", queryDomainName, dns.TypeToString[queryType])

	resp := newResponse(req)
	queryNorm := normalizeDomain(queryDomainName)
	parts := strings.Split(queryNorm, ".")

	if len(parts) < 2 {
		sendErrorResponse(w, resp, dns.RcodeNameError)
		return
	}

	rootDomain := strings.Join(parts[1:], ".")
	log.Println("rootDomain", rootDomain)

	// Проверяем wildcard записи в зоне (например *.info) и отдадем значения по типу msg.info.test.ru
	if found := handleWildcardMatch(w, resp, queryDomainName, queryNorm, rootDomain, queryType); found {
		return
	}

	// Проверяем точное совпадение
	if found := handleExactMatch(w, resp, queryDomainName, queryNorm, parts, queryType); found {
		return
	}

	// Проверяем wildcard записи в зоне (например *.info) отдаем значение самой записи если в запросе info.test.ru
	if found := handleExactMatchRecord(w, resp, queryDomainName, queryNorm, parts, queryType); found {
		return
	}

	// Проверяем wildcard запись *.test.ru
	if found := handleWildcardFallback(w, resp, queryDomainName, rootDomain, queryType); found {
		return
	}

	sendErrorResponse(w, resp, dns.RcodeNameError)
}

func newResponse(req *dns.Msg) *dns.Msg {
	resp := new(dns.Msg)
	resp.ID = req.ID
	resp.Response = true
	resp.Question = req.Question
	resp.Authoritative = true
	return resp
}

func normalizeDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(domain), ".")
}

// Проверяем wildcard записи в зоне (например *.info)
func handleWildcardMatch(w dns.ResponseWriter, resp *dns.Msg, queryDomain, queryNorm, rootDomain string, qType uint16) bool {
	for zoneKey := range utils.Zone {
		log.Println("zoneKey", zoneKey)
		if !strings.HasPrefix(zoneKey, "*.") {
			continue
		}
		zoneParts := strings.Split(zoneKey, ".")
		if len(zoneParts) < 2 {
			continue
		}

		queryParts := strings.Split(queryNorm, ".") // разбиение запроса пользователя на части info test ru
		if len(queryParts) < 2 || zoneParts[1] != queryParts[1] {
			continue
		}

		wildcardKey := "*." + rootDomain // *.test.ru
		log.Println("rootDomain", rootDomain)
		wildcardData := utils.Zone[wildcardKey]

		switch qType {
		case dns.TypeA:
			addResponseARecords(resp, queryDomain, wildcardData.A)
		case dns.TypeNS:
			addResponseNSRecords(resp, wildcardData.NS)
		}

		sendResponse(w, resp, dns.RcodeSuccess)
		return true
	}
	return false
}

// Проверяем wildcard записи в зоне (например *.info) отдаем значение самой записи если в запросе info.test.ru
func handleExactMatchRecord(w dns.ResponseWriter, resp *dns.Msg, queryDomain, queryNorm string, parts []string, qType uint16) bool {
	if len(parts) > 2 {
		data, ok := utils.Zone["*."+queryNorm]
		log.Println("DATA", data)
		if !ok {
			return false
		}
		switch qType {
		case dns.TypeA:
			addResponseARecords(resp, queryDomain, data.A)
		case dns.TypeNS:
			if len(parts) > 2 {
				// Для поддоменов используем NS от корневого домена
				if rootData, ok := utils.Zone["test.ru"]; ok {
					addResponseNSRecords(resp, rootData.NS)
				}
			} else {
				addResponseNSRecords(resp, data.NS)
			}
		}

		sendResponse(w, resp, dns.RcodeSuccess)
		return true
	}
	return false
}

// Проверяем точное совпадение
func handleExactMatch(w dns.ResponseWriter, resp *dns.Msg, queryDomain, queryNorm string, parts []string, qType uint16) bool {
	data, ok := utils.Zone[queryNorm]
	if !ok {
		return false
	}

	switch qType {
	case dns.TypeA:
		addResponseARecords(resp, queryDomain, data.A)
	case dns.TypeNS:
		if len(parts) > 2 {
			// Для поддоменов используем NS от корневого домена
			if rootData, ok := utils.Zone["test.ru"]; ok {
				addResponseNSRecords(resp, rootData.NS)
			}
		} else {
			addResponseNSRecords(resp, data.NS)
		}
	}

	sendResponse(w, resp, dns.RcodeSuccess)
	return true
}

// Проверяем wildcard запись *.test.ru
func handleWildcardFallback(w dns.ResponseWriter, resp *dns.Msg, queryDomain, rootDomain string, qType uint16) bool {
	parts := strings.Split(queryDomain, ".")
	if len(parts) <= 2 {
		return false
	}

	// Проверяем наличие wildcard записи *.test.ru
	if _, ok := utils.Zone["*.test.ru"]; !ok {
		sendErrorResponse(w, resp, dns.RcodeNameError)
		return true
	}

	rootData, ok := utils.Zone[rootDomain]
	if !ok {
		sendErrorResponse(w, resp, dns.RcodeNameError)
		return true
	}

	switch qType {
	case dns.TypeA:
		addResponseARecords(resp, queryDomain, rootData.A)
	case dns.TypeNS:
		if rootNS, ok := utils.Zone["test.ru"]; ok {
			addResponseNSRecords(resp, rootNS.NS)
		}
	}

	sendResponse(w, resp, dns.RcodeSuccess)
	return true
}

func addResponseARecords(resp *dns.Msg, responseDomain string, ips []string) {
	responseDomain = addTrailingDot(responseDomain)

	for _, ipStr := range ips {
		ip, err := netip.ParseAddr(ipStr)
		if err != nil {
			log.Printf("Failed to parse IP %s: %v", ipStr, err)
			continue
		}

		a := &dns.A{
			Hdr: dns.Header{Name: responseDomain, Class: dns.ClassINET, TTL: defaultTTL},
			A:   rdata.A{Addr: ip},
		}
		resp.Answer = append(resp.Answer, a)
	}
}

func addResponseNSRecords(resp *dns.Msg, responseNSDomains []string) {
	for _, nsDomain := range responseNSDomains {
		ns := &dns.NS{
			Hdr: dns.Header{Name: nsDomain, Class: dns.ClassINET, TTL: defaultTTL},
			NS:  rdata.NS{Ns: nsDomain},
		}
		resp.Answer = append(resp.Answer, ns)
	}
}

func addTrailingDot(domain string) string {
	if !strings.HasSuffix(domain, ".") {
		return domain + "."
	}
	return domain
}

func sendResponse(w dns.ResponseWriter, resp *dns.Msg, rcode uint16) {
	resp.Rcode = rcode
	if err := resp.Pack(); err != nil {
		log.Printf("Pack error: %v", err)
		return
	}
	if _, err := io.Copy(w, resp); err != nil {
		log.Printf("DNS write error: %v", err)
	}
}

func sendErrorResponse(w dns.ResponseWriter, resp *dns.Msg, rcode uint16) {
	resp.Rcode = rcode
	resp.Answer = nil // Очищаем ответы для ошибок
	if err := resp.Pack(); err != nil {
		log.Printf("Pack error: %v", err)
		return
	}
	if _, err := io.Copy(w, resp); err != nil {
		log.Printf("DNS write error: %v", err)
	}
}

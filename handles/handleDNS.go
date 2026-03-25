package handles

import (
	"context"
	"hllb/algorithm"
	"hllb/checks"
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
	cfg := utils.GetConfig()

	q := req.Question[0]
	queryDomainName := q.Header().Name
	queryType := dns.RRToType(q)

	// log.Printf("DNS query: %s, type: %s", queryDomainName, dns.TypeToString[queryType])

	resp := newResponse(req)
	queryNorm := normalizeDomain(queryDomainName)
	parts := strings.Split(queryNorm, ".")

	if len(parts) < 2 {
		sendErrorResponse(w, resp, dns.RcodeNameError)
		return
	}

	rootDomain := strings.Join(parts[1:], ".")

	// Проверяем точное совпадение
	if found := handleExactMatch(w, resp, queryDomainName, queryNorm, parts, queryType, cfg, rootDomain); found {
		return
	}

	// Проверяем wildcard записи в зоне (например *.info)
	if found := handleWildcardMatch(w, resp, queryDomainName, queryNorm, rootDomain, queryType, cfg); found {
		return
	}

	// Проверяем wildcard запись *.rootDomain
	if found := handleWildcardFallback(w, resp, queryDomainName, rootDomain, queryType, cfg); found {
		return
	}

	// Если запись не найдена в зонах и включён forward — пересылаем на внешний DNS
	if cfg.App.Forward {
		forwardHandlerDNS(ctx, w, req)
		return
	}

	sendErrorResponse(w, resp, dns.RcodeNameError)
}

func newResponse(req *dns.Msg) *dns.Msg {
	resp := new(dns.Msg)
	resp.ID = req.ID
	resp.Response = true
	resp.Question = req.Question
	return resp
}

func normalizeDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(domain), ".")
}

// Проверяем wildcard записи в зоне (например *.info)
func handleWildcardMatch(w dns.ResponseWriter, resp *dns.Msg, queryDomain, queryNorm, rootDomain string, qType uint16, cfg *utils.Config) bool {
	snapshot := utils.GetZoneSnapshot()
	for zoneKey := range snapshot {
		// log.Println("zoneKey", zoneKey)
		if !strings.HasPrefix(zoneKey, "*.") {
			continue
		}
		zoneParts := strings.Split(zoneKey, ".") // разбиение запроса пользователя на части info test ru
		if len(zoneParts) < 2 {
			continue
		}

		wildRecord := strings.TrimPrefix(zoneKey, "*.") // находим запись в зоне с началом *. - получим info.test.ru
		if len(wildRecord) != 0 {
			if strings.Contains(queryNorm, wildRecord) { // проверяем если вхождение info.test.ru в запросе пользователя например dd.asdf.info.test.ru
				wildcardKey := "*." + wildRecord // если есть получаем *.info.test.ru и извлекаем ip из А записи в зоне
				wildcardData := snapshot[wildcardKey]

				switch qType {
				case dns.TypeA:
					addResponseARecords(w, resp, queryDomain, wildcardData.A, cfg)
				case dns.TypeNS:
					if rootData, ok := snapshot[rootDomain]; ok {
						addResponseNSRecords(resp, rootData.NS)
					}
				}

				sendResponse(w, resp, dns.RcodeSuccess)
				return true
			}
		}
	}
	return false
}

// Проверяем точное совпадение
func handleExactMatch(w dns.ResponseWriter, resp *dns.Msg, queryDomain, queryNorm string, parts []string, qType uint16, cfg *utils.Config, rootDomain string) bool {
	data, ok := utils.GetZone(queryNorm)
	if !ok {
		return false
	}

	switch qType {
	case dns.TypeA:
		addResponseARecords(w, resp, queryDomain, data.A, cfg)
	case dns.TypeNS:
		if rootData, ok := utils.GetZone(rootDomain); ok {
			addResponseNSRecords(resp, rootData.NS)
		}
	}

	sendResponse(w, resp, dns.RcodeSuccess)
	return true
}

// Проверяем wildcard запись *.rootDomain
func handleWildcardFallback(w dns.ResponseWriter, resp *dns.Msg, queryDomain, rootDomain string, qType uint16, cfg *utils.Config) bool {
	parts := strings.Split(queryDomain, ".")
	if len(parts) <= 2 {
		return false
	}

	if _, ok := utils.GetZone("*." + rootDomain); !ok {
		return false
	}

	rootData, ok := utils.GetZone(rootDomain)
	if !ok {
		return false
	}
	switch qType {
	case dns.TypeA:
		addResponseARecords(w, resp, queryDomain, rootData.A, cfg)
	case dns.TypeNS:
		if rootNS, ok := utils.GetZone(rootDomain); ok {
			addResponseNSRecords(resp, rootNS.NS)
		}
	}

	sendResponse(w, resp, dns.RcodeSuccess)
	return true
}

// isInCheckPool проверяет, есть ли хотя бы один IP записи в пуле check.yaml
func isInCheckPool(ips []string) bool {
	for _, ip := range ips {
		if checks.IsInPool(ip) {
			return true
		}
	}
	return false
}

// Добавляем ответы с A записью
func addResponseARecords(w dns.ResponseWriter, resp *dns.Msg, responseDomain string, ips []string, cfg *utils.Config) {
	responseDomain = addTrailingDot(responseDomain)

	// RR только для записей, чьи IP находятся в пуле check.yaml
	if cfg.App.ActiveCheck && isInCheckPool(ips) {
		ip, err := algorithm.RR()
		if err != nil {
			sendErrorResponse(w, resp, dns.RcodeNameError)
			return
		}

		a := &dns.A{
			Hdr: dns.Header{Name: responseDomain, Class: dns.ClassINET, TTL: defaultTTL},
			A:   rdata.A{Addr: ip},
		}
		resp.Answer = append(resp.Answer, a)
		return
	}

	// Для остальных записей — просто возвращаем IP из зоны
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
	resp.Authoritative = true
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

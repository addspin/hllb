package handles

import (
	"context"
	"log"

	"codeberg.org/miekg/dns"
)

var forwardAddr string

func InitForward(host, port string) {
	forwardAddr = host + ":" + port
}

func forwardHandlerDNS(ctx context.Context, w dns.ResponseWriter, req *dns.Msg) {

	if len(req.Question) == 0 {
		return
	}

	// Если стоит флаг forward, то перенаправляем запросы на внешний dns сервер
	c := new(dns.Client)

	resp, _, err := c.Exchange(ctx, req, "udp", forwardAddr)
	if err != nil {
		log.Printf("Forward error: %v", err)
		errResp := newResponse(req)
		sendErrorResponse(w, errResp, dns.RcodeServerFailure)
		return
	}
	sendResponse(w, resp, dns.RcodeSuccess)
}

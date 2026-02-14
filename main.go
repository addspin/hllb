package main

import (
	"hllb/handles"
	"hllb/utils"
	"log"
	"time"

	"codeberg.org/miekg/dns"
)

const (
	addr = ":1053"
)

func main() {
	s := utils.InitZone()
	log.Println("НАЧАЛО", s)
	go utils.WatchZoneFile("./zone/test.ru", 5*time.Second)

	dns.HandleFunc(".", handles.HandleDNS)
	go func() {
		if err := dns.ListenAndServe(addr, "udp", nil); err != nil {
			log.Fatalf("DNS UDP: %v", err)
		}
	}()
	if err := dns.ListenAndServe(addr, "tcp", nil); err != nil {
		log.Fatalf("DNS TCP: %v", err)
	}

}

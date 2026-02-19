package main

import (
	"hllb/checks"
	"hllb/handles"
	"hllb/utils"
	"log"
	"os"
	"time"

	"codeberg.org/miekg/dns"
)

func main() {
	s := utils.InitZone()
	log.Println("НАЧАЛО", s)
	cfg, err := utils.ReadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	port := ":" + cfg.App.Port
	checkZoneInterval := cfg.App.CheckZoneInterval

	allFileZone, err := os.ReadDir("./zone")
	if err != nil {
		log.Fatal(err)
	}
	checks.TCPCheck()
	// Получаем список файлов зон
	for _, file := range allFileZone {
		if file.IsDir() { // Пропускаем поддиректории
			continue
		}
		go utils.WatchZoneFile("./zone/"+file.Name(), time.Duration(checkZoneInterval)*time.Second)
	}
	dns.HandleFunc(".", handles.HandleDNS)
	go func() {
		if err := dns.ListenAndServe(port, "udp", nil); err != nil {
			log.Fatalf("DNS UDP: %v", err)
		}
	}()
	if err := dns.ListenAndServe(port, "tcp", nil); err != nil {
		log.Fatalf("DNS TCP: %v", err)
	}

}

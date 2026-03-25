package main

import (
	"hllb/checks"
	"hllb/handles"
	"hllb/utils"
	"log"
	_ "net/http/pprof"
	"os"

	"codeberg.org/miekg/dns"
)

func main() {
	// Включить для профилирования
	// go func() {
	// 	log.Println("pprof: http://localhost:6060/debug/pprof/")
	// 	log.Println(http.ListenAndServe(":6060", nil))
	// }()

	utils.EnsureRequiredFiles()
	utils.InitZone()

	cfg, err := utils.InitConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	port := ":" + cfg.App.Port
	checkZoneInterval := cfg.App.CheckZoneInterval
	checkZoneIntervalType := cfg.App.CheckZoneIntervalType
	repeatCheckTime := cfg.App.RepeatCheckInterval
	repeatCheckTimeType := cfg.App.RepeatCheckIntervalType
	repeatCheckFileTime := cfg.App.RepeatCheckFileInterval
	repeatCheckFileTimeType := cfg.App.RepeatCheckFileIntervalType
	forwardDNS := cfg.App.ForwardDNS
	forwardDNSPort := cfg.App.ForwardDNSPort
	handles.InitForward(forwardDNS, forwardDNSPort)

	allFileZone, err := os.ReadDir("./zone")
	if err != nil {
		log.Fatal(err)
	}
	if cfg.App.ActiveCheck {
		ready := make(chan struct{})
		go utils.WatchCheckFile("./check.yaml", utils.SelectTime(repeatCheckFileTimeType, repeatCheckFileTime), ready)
		<-ready
		serverInterval := utils.SelectTime(repeatCheckTimeType, repeatCheckTime)
		checkTCP := checks.StatusCodeTcp{}
		go checkTCP.TCPCheck(serverInterval)
	}

	for _, file := range allFileZone {
		if file.IsDir() {
			continue
		}
		go utils.WatchZoneFile("./zone/"+file.Name(), utils.SelectTime(checkZoneIntervalType, checkZoneInterval))
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

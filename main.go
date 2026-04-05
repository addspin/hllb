package main

import (
	"hllb/checks"
	"hllb/handles"
	"hllb/metrics"
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
	pathLogs := cfg.App.PathLog
	lvlLog := cfg.App.LvlLog
	active := cfg.App.ActiveLog
	utils.InitLogs(pathLogs, lvlLog, active)

	metrics.Init()
	if cfg.App.MetricsPort != "" {
		go metrics.ServeHTTP(":" + cfg.App.MetricsPort)
		utils.LogInfo("Metrics started on :%s/metrics", cfg.App.MetricsPort)
	}

	port := cfg.App.Port

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
		if err := dns.ListenAndServe(":"+port, "udp", nil); err != nil {
			utils.LogError("DNS UDP: %v", err)
			log.Fatalf("DNS UDP: %v", err)
		}
	}()
	utils.LogInfo("HLLB started UDP:%s", port)
	utils.LogInfo("HLLB started TCP:%s", port)
	if err := dns.ListenAndServe(":"+port, "tcp", nil); err != nil {
		utils.LogError("DNS TCP: %v", err)
		log.Fatalf("DNS TCP: %v", err)
	}
}

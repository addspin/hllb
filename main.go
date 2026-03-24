package main

import (
	"hllb/checks"
	"hllb/handles"
	"hllb/utils"
	"log"
	"os"

	"codeberg.org/miekg/dns"
)

func main() {
	utils.InitZone()
	// log.Println("НАЧАЛО", s)
	cfg, err := utils.ReadConfig("config.yaml")
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

	allFileZone, err := os.ReadDir("./zone")
	if err != nil {
		log.Fatal(err)
	}
	if cfg.App.ActiveCheck {
		ready := make(chan struct{})
		// Получаем конфигурацию check файла
		go utils.WatchCheckFile("./check.yaml", utils.SelectTime(repeatCheckFileTimeType, repeatCheckFileTime), ready)
		<-ready // если конфиг получен
		// Проверяем досупность серверов
		serverInterval := utils.SelectTime(repeatCheckTimeType, repeatCheckTime)
		checkTCP := checks.StatusCodeTcp{}
		go checkTCP.TCPCheck(serverInterval)
	}

	// Получаем список файлов зон
	for _, file := range allFileZone {
		if file.IsDir() { // Пропускаем поддиректории
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

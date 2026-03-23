package checks

import (
	"hllb/utils"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type StatusCodeTcp struct {
	MutexTcp sync.Mutex
}

var ValidPoolHost []string

func (s *StatusCodeTcp) TCPCheck(timeTicker time.Duration) {
	// Выполняем проверку сразу при запуске
	s.checkPort()

	// Затем запускаем периодическую проверку
	ticker := time.NewTicker(timeTicker)
	defer ticker.Stop()

	for range ticker.C {
		s.checkPort()
	}
}

func (s *StatusCodeTcp) checkPort() {
	s.MutexTcp.Lock()
	defer s.MutexTcp.Unlock()

	c := utils.CheckFile
	if len(c.HostCheck) == 0 {
		log.Println("checkItems: No server for check list")
	}
	port := c.PortCheck
	newPool := make([]string, 0, len(c.HostCheck))

	for _, server := range c.HostCheck {
		log.Println("checkItems:", server)
		conn, err := net.DialTimeout("tcp", server+":"+strconv.Itoa(port), 1*time.Second)
		if err != nil {
			log.Println("Connection failed")
		} else {
			newPool = append(newPool, server)
			log.Println("Connection established")
			connErr := conn.Close()
			if connErr != nil {
				log.Printf("Error close connections %s", connErr)
			}
		}
	}
	ValidPoolHost = newPool
	log.Printf("ValidPoolHost %s", ValidPoolHost)
	
}

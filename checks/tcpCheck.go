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
	//ExitCodeTcp bool
	//Server      string
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
	log.Println("TEST", c)
	if len(c.HostCheck) == 0 {
		log.Println("checkItems: No server for check list")
	}
	port := c.PortCheck

	for _, server := range c.HostCheck {
		log.Println("checkItems:", server)
		conn, err := net.Dial("tcp", server+":"+strconv.Itoa(port))
		if err != nil {
			//s.ExitCodeTcp = false // port is not available
			log.Println("Connection failed")
		} else {
			ValidPoolHost = append(ValidPoolHost, server)
			//s.ExitCodeTcp = true // port is available
			log.Println("Connection established")
			connErr := conn.Close()
			if connErr != nil {
				log.Fatal(connErr)
			}
		}
	}
}

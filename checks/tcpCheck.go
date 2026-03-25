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

var (
	validPoolHost []string
	validPoolSet  map[string]struct{}
	poolMutex     sync.RWMutex
)

// GetValidPoolHost возвращает копию слайса живых хостов (для Algorithm balance).
func GetValidPoolHost() []string {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	cp := make([]string, len(validPoolHost))
	copy(cp, validPoolHost)
	return cp
}

// IsInPool проверяет наличие IP в пуле
func IsInPool(ip string) bool {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	_, ok := validPoolSet[ip]
	return ok
}

func (s *StatusCodeTcp) TCPCheck(timeTicker time.Duration) {
	s.checkPort()

	ticker := time.NewTicker(timeTicker)
	defer ticker.Stop()

	for range ticker.C {
		s.checkPort()
	}
}

func (s *StatusCodeTcp) checkPort() {
	s.MutexTcp.Lock()
	defer s.MutexTcp.Unlock()

	c := utils.GetCheckFile()
	if len(c.HostCheck) == 0 {
		log.Println("checkItems: No server for check list")
	}
	port := c.PortCheck
	newPool := make([]string, 0, len(c.HostCheck))
	newSet := make(map[string]struct{}, len(c.HostCheck))

	for _, server := range c.HostCheck {
		log.Println("checkItems:", server)
		conn, err := net.DialTimeout("tcp", server+":"+strconv.Itoa(port), 1*time.Second)
		if err != nil {
			log.Println("Connection failed")
		} else {
			newPool = append(newPool, server)
			newSet[server] = struct{}{}
			log.Println("Connection established")
			connErr := conn.Close()
			if connErr != nil {
				log.Printf("Error close connections %s", connErr)
			}
		}
	}
	poolMutex.Lock()
	validPoolHost = newPool // для Algorithm
	validPoolSet = newSet   // для сравнения данных чекера в IsInPool с ip
	poolMutex.Unlock()
	log.Printf("ValidPoolHost %s", newPool)
}

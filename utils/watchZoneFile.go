package utils

import (
	"log"
	"sync"
	"time"
)

var zoneMutex sync.RWMutex

func WatchZoneFile(path string, interval time.Duration) {
	// Запоминаем начальный хеш
	lastHash, _ := GetFileHash(path)

	ticker := time.NewTicker(interval)
	for range ticker.C {
		currentHash, err := GetFileHash(path)
		if err != nil {
			continue
		}

		if currentHash != lastHash {
			log.Printf("Хеш изменился [%s], обновляю зону %s", currentHash[:8], path)

			newZone := InitZone()

			zoneMutex.Lock()
			Zone = newZone
			zoneMutex.Unlock()

			lastHash = currentHash
		}
	}
}

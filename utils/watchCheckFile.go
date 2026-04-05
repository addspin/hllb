package utils

import (
	"log"
	"time"
)

func WatchCheckFile(path string, interval time.Duration, ready chan<- struct{}) {
	lastHash, _ := GetFileHash(path)
	err := ReadCheckConfig(path)
	if err != nil {
		log.Fatalf("Error read check file config %s", err)
	}
	close(ready)

	ticker := time.NewTicker(interval)
	for range ticker.C {
		currentHash, err := GetFileHash(path)
		if err != nil {
			continue
		}

		if currentHash != lastHash {
			LogInfo("Hash change [%s], update check file %s", currentHash[:8], path)

			err := ReadCheckConfig(path)
			if err != nil {
				LogError("Error read check file %s", err)
			}

			lastHash = currentHash
		}
	}
}

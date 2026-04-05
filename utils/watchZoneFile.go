package utils

import (
	"time"
)

func WatchZoneFile(path string, interval time.Duration) {
	lastHash, _ := GetFileHash(path)

	ticker := time.NewTicker(interval)
	for range ticker.C {
		currentHash, err := GetFileHash(path)
		if err != nil {
			continue
		}

		if currentHash != lastHash {
			LogInfo("Hash changed [%s], reloading zone %s", currentHash[:8], path)
			InitZone()
			lastHash = currentHash
		}
	}
}

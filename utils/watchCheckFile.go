package utils

import (
	"log"
	"time"
)

func WatchCheckFile(path string, interval time.Duration) {
	// Запоминаем начальный хеш
	lastHash, _ := GetFileHash(path)
	err := ReadCheckConfig(path)
	if err != nil {
		log.Fatalf("Ошибка чтения %s", err)
	}

	ticker := time.NewTicker(interval)
	for range ticker.C {
		currentHash, err := GetFileHash(path)
		if err != nil {
			continue
		}

		if currentHash != lastHash {
			log.Printf("Хеш изменился [%s], обновляю check файл %s", currentHash[:8], path)

			err := ReadCheckConfig(path)
			if err != nil {
				log.Fatalf("Ошибка чтения %s", err)
			}

			lastHash = currentHash
		}
	}
}

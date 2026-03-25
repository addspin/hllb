package utils

import (
	"log"
	"os"
)

const defaultConfig = `app:
  port: 53
  checkZoneInterval: 5
  checkZoneIntervalType: seconds
  activeCheck: false
  algorithmCheck: RR
  repeatCheckInterval: 3
  repeatCheckIntervalType: seconds
  repeatCheckFileInterval: 3
  repeatCheckFileIntervalType: seconds
  forward: true
  forwardDNS: 8.8.8.8
  forwardDNSPort: 53
`

const defaultCheck = `# hostCheck - ip for health check, must match A records in zone (e.g. lb.example.com)
hostCheck:
  - 127.0.0.1
portCheck: 22
`

const defaultZone = `$TTL 3600
@       IN      SOA     ns1.example.com. admin.example.com. (
                        2026010101
                        7200
                        3600
                        1209600
                        86400
                        )

@       IN      NS      ns1.example.com.
@       IN      A       127.0.0.1
`

// EnsureRequiredFiles проверяет наличие необходимых файлов и папок при старте.
func EnsureRequiredFiles() {
	ensureFile("config.yaml", defaultConfig)
	ensureFile("check.yaml", defaultCheck)
	ensureZoneDir()
}

func ensureFile(path, content string) {
	if _, err := os.Stat(path); err == nil {
		return
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		log.Fatalf("Failed to create %s: %v", path, err)
	}
	log.Printf("Created default %s", path)
}

func ensureZoneDir() {
	if _, err := os.Stat("./zone"); err == nil {
		return
	}

	if err := os.Mkdir("./zone", 0755); err != nil {
		log.Fatalf("Failed to create zone directory: %v", err)
	}
	log.Println("Created ./zone/")

	ensureFile("./zone/example.com", defaultZone)
}

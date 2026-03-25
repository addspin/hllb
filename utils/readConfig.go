package utils

import (
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Port                        string `yaml:"port"`
		CheckZoneInterval           int    `yaml:"checkZoneInterval"`
		CheckZoneIntervalType       string `yaml:"checkZoneIntervalType"`
		ActiveCheck                 bool   `yaml:"activeCheck"`
		AlgorithmCheck              string `yaml:"algorithmCheck"`
		RepeatCheckInterval         int    `yaml:"repeatCheckInterval"`
		RepeatCheckIntervalType     string `yaml:"repeatCheckIntervalType"`
		RepeatCheckFileInterval     int    `yaml:"repeatCheckFileInterval"`
		RepeatCheckFileIntervalType string `yaml:"repeatCheckFileIntervalType"`
		Forward                     bool   `yaml:"forward"`
		ForwardDNS                  string `yaml:"forwardDNS"`
		ForwardDNSPort              string `yaml:"forwardDNSPort"`
	} `yaml:"app"`
}

var (
	cachedConfig *Config
	configMutex  sync.RWMutex
)

// InitConfig загружает конфиг при старте и сохраняет в кэш.
func InitConfig(path string) (*Config, error) {
	cfg, err := readConfigFromDisk(path)
	if err != nil {
		return nil, err
	}
	configMutex.Lock()
	cachedConfig = cfg
	configMutex.Unlock()
	return cfg, nil
}

// GetConfig возвращает кэшированный конфиг без чтения с диска.
func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return cachedConfig
}

// ReloadConfig перечитывает конфиг с диска и обновляет кэш.
func ReloadConfig(path string) error {
	cfg, err := readConfigFromDisk(path)
	if err != nil {
		return err
	}
	configMutex.Lock()
	cachedConfig = cfg
	configMutex.Unlock()
	log.Printf("Config reloaded from %s", path)
	return nil
}

// ReadConfig оставлен для обратной совместимости, но не кэширует.
func ReadConfig(path string) (*Config, error) {
	return readConfigFromDisk(path)
}

func readConfigFromDisk(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

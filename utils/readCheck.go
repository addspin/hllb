package utils

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type CheckConfig struct {
	HostCheck []string `yaml:"hostCheck"`
	PortCheck int      `yaml:"portCheck"`
}

var (
	checkFile      CheckConfig
	checkFileMutex sync.RWMutex
)

func GetCheckFile() CheckConfig {
	checkFileMutex.RLock()
	defer checkFileMutex.RUnlock()
	return checkFile
}

func ReadCheckConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg CheckConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	checkFileMutex.Lock()
	checkFile = cfg
	checkFileMutex.Unlock()
	return nil
}

package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

type CheckConfig struct {
	HostCheck []string `yaml:"hostCheck"`
	PortCheck int      `yaml:"portCheck"`
}

func ReadCheckConfig(path string) (*CheckConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg CheckConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

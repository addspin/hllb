package utils

import (
	"os"

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
	} `yaml:"app"`
}

func ReadConfig(path string) (*Config, error) {
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

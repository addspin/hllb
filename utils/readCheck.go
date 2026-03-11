package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

type CheckConfig struct {
	HostCheck []string `yaml:"hostCheck"`
	PortCheck int      `yaml:"portCheck"`
}

var CheckFile CheckConfig

func ReadCheckConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &CheckFile); err != nil {
		return err
	}
	return nil
}

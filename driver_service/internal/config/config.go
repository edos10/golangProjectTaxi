package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Http   Http   `yaml:"http"`
	SemVer string `yaml:"semver"`
}

type Http struct {
	Port int `yaml:"port"`
}

func NewConfig(filePath string) (*Config, error) {
	curDir, _ := os.Getwd()
	filePath = curDir + filePath
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	fmt.Println(cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

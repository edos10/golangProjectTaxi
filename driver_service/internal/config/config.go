package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Http   Http   `yaml:"http"`
	SemVer string `yaml:"semver"`
}

type Http struct {
	Port int `yaml:"port"`
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return nil, err
	}

	var cfg Config

	portStr := os.Getenv("HTTP_PORT")
	if portStr == "" {
		return nil, fmt.Errorf("HTTP_PORT is not set in .env")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("HTTP_PORT must be a valid integer")
	}

	cfg.Http.Port = port
	cfg.SemVer = os.Getenv("SEMVER")

	return &cfg, nil
}

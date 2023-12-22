package app

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/location/internal/httpserver"
	"github.com/location/internal/persistency/database"
)

const (
	AppName                     = "auth"
	DefaultServeAddress         = "localhost:9626"
	DefaultShutdownTimeout      = 20 * time.Second
	DefaultBasePath             = "/auth/v1"
	DefaultAccessTokenCookie    = "access_token"
	DefaultRefreshTokenCookie   = "refresh_token"
	DefaultSigningKey           = "qwerty"
	DefaultAccessTokenDuration  = 1 * time.Minute
	DefaultRefreshTokenDuration = 1 * time.Hour
	DefaultDSN                  = "dsn://"
	DefaultMigrationsDir        = "file://migrations/auth"
)

type Config struct {
	HTTP     httpserver.HttpConfig `yaml:"http"`
	Database database.DatabaseConfig     `yaml:"database"`
}

func NewConfig(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	cnf := Config{
		HTTP: httpserver.HttpConfig{
			ServeAddress:       DefaultServeAddress,
			BasePath:           DefaultBasePath,
		},
		Database: database.DatabaseConfig{
			DSN:           DefaultDSN,
			MigrationsDir: DefaultMigrationsDir,
		},
	}

	if err := yaml.Unmarshal(data, &cnf); err != nil {
		return nil, err
	}

	return &cnf, nil
}
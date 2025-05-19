package config

import (
	"errors"
	"os"
)

const (
	serverAddressEnvName = "GOPHKEEPER_SERVER"
)

type Config struct {
	ServerAddress string
}

func NewConfig() (*Config, error) {
	serverAddress := os.Getenv(serverAddressEnvName)
	if serverAddress == "" {
		return nil, errors.New("missing server address")
	}

	return &Config{
		ServerAddress: serverAddress,
	}, nil
}

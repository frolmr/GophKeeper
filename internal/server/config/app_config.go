package config

import (
	"errors"
	"flag"
	"os"
)

type AppConfig struct {
	RunAddress  string
	DatabaseURI string
	TLSCertFile string
	TLSKeyFile  string
}

const (
	runAddressEnvName    = "RUN_ADDRESS"
	dadatabaseURIEnvName = "DATABASE_URI"
)

var (
	ErrMissingAddress  = errors.New("missing run address")
	ErrMissingDDURI    = errors.New("missing database URI")
	ErrMissingCertFile = errors.New("missing TLS certificate file")
	ErrMissingKeyFile  = errors.New("missing TLS key file")
)

func NewAppConfig() (*AppConfig, error) {
	var (
		runAddress  string
		databaseURI string
		certFile    string
		keyFile     string
		errs        []error
	)

	flag.StringVar(&runAddress, "a", runAddress, "sets host and port to run")
	flag.StringVar(&databaseURI, "d", databaseURI, "set database URI to use")
	flag.StringVar(&certFile, "c", certFile, "sets TLS certificate file")
	flag.StringVar(&keyFile, "k", keyFile, "set TLS key file")
	flag.Parse()

	if runAddressEnv := os.Getenv(runAddressEnvName); runAddressEnv != "" && runAddress == "" {
		runAddress = runAddressEnv
	}

	if databaseURIEnv := os.Getenv(dadatabaseURIEnvName); databaseURIEnv != "" && databaseURI == "" {
		databaseURI = databaseURIEnv
	}

	if runAddress == "" {
		errs = append(errs, ErrMissingAddress)
	}

	if databaseURI == "" {
		errs = append(errs, ErrMissingDDURI)
	}

	if certFile == "" {
		errs = append(errs, ErrMissingCertFile)
	}

	if keyFile == "" {
		errs = append(errs, ErrMissingKeyFile)
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return &AppConfig{
		RunAddress:  runAddress,
		DatabaseURI: databaseURI,
		TLSCertFile: certFile,
		TLSKeyFile:  keyFile,
	}, nil
}

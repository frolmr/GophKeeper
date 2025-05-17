package config

type Config struct {
	AppName string
}

func NewConfig(appName string) *Config {
	return &Config{
		AppName: appName,
	}
}

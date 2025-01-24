package config

import (
	viseconfig "git.grassecon.net/grassrootseconomics/visedriver/config"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/config"
	"git.grassecon.net/grassrootseconomics/visedriver/env"
)

var (
	JetstreamURL        string
	JetstreamClientName string
	Apply               = config.Apply
	GetConns            = config.GetConns
)

const (
	defaultJetstreamURL        string = "localhost:4222"
	defaultJetstreamClientName string = "omnom"
)

func LoadConfig() error {
	err := config.LoadConfig()
	if err != nil {
		return err
	}
	JetstreamURL = env.GetEnv("NATS_JETSTREAM_URL", defaultJetstreamURL)
	JetstreamClientName = env.GetEnv("NATS_JETSTREAM_CLIENT_NAME", defaultJetstreamClientName)
	return nil
}

func Language() string {
	return viseconfig.DefaultLanguage
}

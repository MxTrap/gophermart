package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	HTTPAdress     string `env:"RUN_ADDRESS"`
	DatabaseDSN    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() (*Config, error) {
	httpAddr := flag.String("a", "", "host and port http")
	databaseDSN := flag.String("d", "", "database DSN")
	accrualAddr := flag.String("r", "", "address of the accrual calculation system")
	flag.Parse()

	cfg := &Config{
		HTTPAdress:     *httpAddr,
		DatabaseDSN:    *databaseDSN,
		AccrualAddress: *accrualAddr,
	}

	err := env.Parse(cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

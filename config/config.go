package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env       string         `yaml:"env" env-default:"local"`
	HTTP      HTTPConfig     `yaml:"http"`
	Postgres  PostgresConfig `yaml:"postgres"`
	Token     Token          `yaml:"token"`
	JWTSecret string         `yaml:"jwt_secret"`
}

type HTTPConfig struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func (cfg HTTPConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

type PostgresConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	Database      string `yaml:"database"`
	MigrationPath string `yaml:"migration_path"`
}

func (cfg PostgresConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)
}

type Token struct {
	MaxCount int8 `yaml:"max_count"`
	TTL      struct {
		Access  time.Duration `yaml:"access" env-default:"1h"`
		Refresh time.Duration `yaml:"refresh" env-default:"60h"`
	} `yaml:"ttl"`
}

func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}

func NewConfig() *Config {

	configPath := fetchConfigPath()

	if configPath == "" {
		panic("Config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("File is not exist")
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("Failed to read config: " + err.Error())
	}
	return &cfg
}

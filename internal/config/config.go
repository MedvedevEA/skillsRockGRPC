package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

type Config struct {
	Env       string    `yaml:"env" env:"AUTH_ENV" env-default:"local"`
	Token     Token     `yaml:"token" env:"AUTH_TOKEN" env-required:"true"`
	Api       Api       `yaml:"api"`
	Store     Store     `yaml:"store"`
	Scheduler Scheduler `yaml:"scheduler"`
}
type Token struct {
	PrivateKeyPath  string        `yaml:"privateKeyPath" env:"AUTH_TOKEN_PRIVATE_KEY_PATH" env-required:"true"`
	AccessLifetime  time.Duration `yaml:"accessLifetime" env:"AUTH_TOKEN_ACCEESS_LIFETIME" env-default:"3600s"`
	RefreshLifetime time.Duration `yaml:"refreshLifetime" env:"AUTH_TOKEN_REFRESH_LIFETIME" env-default:"2592000s"`
}

type Api struct {
	Addr         string        `yaml:"addr" env:"AUTH_API_ADDR" env-required:"true"`
	WriteTimeout time.Duration `yaml:"writeTimeout" env:"AUTH_API_WRITE_TIMEOUT" env-required:"true"`
	Name         string        `yaml:"name" env:"AUTH_API_NAME" env-required:"true"`
}

type Store struct {
	Host                string        `yaml:"host" env:"AUTH_STORE_HOST" env-required:"true"`
	Port                int           `yaml:"port" env:"AUTH_STORE_PORT" env-required:"true"`
	Name                string        `yaml:"name" env:"AUTH_STORE_NAME" env-required:"true"`
	User                string        `yaml:"user" env:"AUTH_STORE_USER" env-required:"true"`
	Password            string        `yaml:"password" env:"AUTH_STORE_PASSWORD" env-required:"true"`
	SSLMode             string        `yaml:"sslMode" env:"AUTH_STORE_SSL_MODE" env-default:"disable"`
	PoolMaxConns        int           `yaml:"poolMaxConns" env:"AUTH_STORE_POOL_MAX_CONNS" env-default:"5"`
	PoolMaxConnLifetime time.Duration `yaml:"poolMaxConnLifeTime" env:"AUTH_STORE_POOL_MAX_CONN_LIFETIME" env-default:"180s"`
	PoolMaxConnIdleTime time.Duration `yaml:"poolMaxConnIidleTime" env:"AUTH_STORE_POOL_MAX_CONN_IDLE_TIME" env-default:"100s"`
}
type Scheduler struct {
	TimeoutRemoveRefreshTokens time.Duration `yaml:"timeoutRemoveRefreshTokens" env:"AUTH_SCHEDULER_TIMEOUT_REMOVE_REFRESH_TOKENS" env-default:"86400s"`
}

func MustLoad() *Config {
	const op = "config.MustLoad"

	configPath := ""
	cfg := new(Config)

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()
	configPath = "./../../config/local.yml"
	if configPath != "" {
		log.Printf("%s: the value of the 'config' flag: %s\n", op, configPath)
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			log.Fatal(errors.Wrap(err, op))
		}
		log.Printf("%s: configuration file %+v", op, cfg)
		return cfg
	}
	log.Printf("%s: the 'config' flag is not set\n", op)

	configPath = os.Getenv("AUTH_CONFIG_PATH")
	if configPath != "" {
		log.Printf("%s: the value of the environment variable: %s\n", op, configPath)
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			log.Fatal(errors.Wrap(err, op))
		}
		log.Printf("%s: configuration file %+v", op, cfg)
		return cfg
	}
	log.Printf("%s: environment variable 'AUTH_CONFIG_PATH' is not set\n", op)

	log.Printf("%s: the parameter file is not defined. Loading the application configuration from the environment variables\n", op)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("%s: %v", op, err)
	}
	log.Printf("%s: configuration file %+v", op, cfg)
	return cfg
}

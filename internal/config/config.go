package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env       string    `yaml:"env" env:"AUTH_ENV" env-default:"local"`
	Token     Token     `yaml:"token" env:"AUTH_TOKEN" env-required:"true"`
	Grpc      Grpc      `yaml:"grpc"`
	Http      Http      `yaml:"http"`
	Store     Store     `yaml:"store"`
	Scheduler Scheduler `yaml:"scheduler"`
}
type Token struct {
	PrivateKeyPath  string        `yaml:"privateKeyPath" env:"AUTH_TOKEN_PRIVATE_KEY_PATH" env-required:"true"`
	AccessLifetime  time.Duration `yaml:"accessLifetime" env:"AUTH_TOKEN_ACCEESS_LIFETIME" env-default:"3600s"`
	RefreshLifetime time.Duration `yaml:"refreshLifetime" env:"AUTH_TOKEN_REFRESH_LIFETIME" env-default:"2592000s"`
}

type Grpc struct {
	Addr         string        `yaml:"addr" env:"AUTH_GRPC_ADDR" env-required:"true"`
	WriteTimeout time.Duration `yaml:"writeTimeout" env:"AUTH_GRPC_WRITE_TIMEOUT" env-required:"true"`
	Name         string        `yaml:"name" env:"AUTH_GRPC_NAME" env-required:"true"`
}
type Http struct {
	Addr string `yaml:"addr" env:"AUTH_HTTP_ADDR" env-required:"true"`
	Name string `yaml:"name" env:"AUTH_HTTP_NAME" env-required:"true"`
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

	configPath := ""
	cfg := new(Config)

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()
	//configPath = "./../../config/local.yml"//for debug
	if configPath != "" {
		log.Printf("CONFIG: the value of the 'config' flag: %s\n", configPath)
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			log.Fatalf("CONFIG: %v\n", err)
		}
		log.Printf("CONFIG: configuration file %+v\n", cfg)
		return cfg
	}
	log.Printf("CONFIG: the 'config' flag is not set\n")

	configPath = os.Getenv("AUTH_CONFIG_PATH")
	if configPath != "" {
		log.Printf("CONFIG: the value of the environment variable: %s\n", configPath)
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			log.Fatalf("CONFIG: %v\n", err)
		}
		log.Printf("CONFIG: configuration file %+v\n", cfg)
		return cfg
	}
	log.Printf("CONFIG: environment variable 'AUTH_CONFIG_PATH' is not set\n")

	log.Printf("CONFIG: the parameter file is not defined. Loading the application configuration from the environment variables\n")
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("CONFIG: %v\n", err)
	}
	log.Printf("CONFIG: configuration file %+v\n", cfg)
	return cfg
}

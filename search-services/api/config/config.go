package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel     string        `env:"LOG_LEVEL" env-default:"INFO" yaml:"log_level"`
	WordsAddress string        `env:"WORDS_ADDRESS" env-default:"localhost:8080" yaml:"words_address"`
	HTTPServer   HTTPServerCfg `yaml:"http_server"`
}

type HTTPServerCfg struct {
	Address string        `env:"HTTP_SERVER_ADDRESS" env-default:":8080" yaml:"address"`
	Timeout time.Duration `env:"HTTP_SERVER_TIMEOUT" env-default:"10s" yaml:"timeout"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Printf("cannot read config %q: %s", configPath, err)
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("cannot read env %s", err)
		}
	}
	return cfg
}

package tfl

import (
	"fmt"
	"os"
)

type Config struct {
	TFLAppKey string
	Port      string
}

func (c Config) Valid() error {
	if c.TFLAppKey != "" {
		return nil
	}
	return fmt.Errorf("TFL_APP_KEY must be set")
}

func Init() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	cfg := Config{
		TFLAppKey: os.Getenv("TFL_APP_KEY"),
		Port:      port,
	}

	return cfg, cfg.Valid()
}

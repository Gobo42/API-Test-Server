package main

import (
	"encoding/json"
	"os"
)

type RouteConfig struct {
	URI     string            `json:"uri"`
	Method  string            `json:"method"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type DefaultConfig struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Config struct {
	Port    int           `json:"port"`
	Default DefaultConfig `json:"default"`
	Routes  []RouteConfig `json:"routes"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.Default.Status == 0 {
		cfg.Default.Status = 200
	}
	for i := range cfg.Routes {
		if cfg.Routes[i].Method == "" {
			cfg.Routes[i].Method = "*"
		}
		if cfg.Routes[i].Status == 0 {
			cfg.Routes[i].Status = 200
		}
	}
	return &cfg, nil
}

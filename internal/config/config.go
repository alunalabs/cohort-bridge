package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Type              string   `yaml:"type"`
		Host              string   `yaml:"host"`
		Port              int      `yaml:"port"`
		User              string   `yaml:"user"`
		Password          string   `yaml:"password"`
		DBName            string   `yaml:"dbname"`
		Table             string   `yaml:"table"`
		Filename          string   `yaml:"filename"`
		Fields            []string `yaml:"fields"`
		RandomBitsPercent float64  `yaml:"random_bits_percent"`
	} `yaml:"database"`
	Peer struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"peer"`
	ListenPort int    `yaml:"listen_port"`
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

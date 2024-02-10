package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const configFileName = "./config/config.yaml"

type Config struct {
	WebServer struct {
		Host               string   `yaml:"host"`
		Port               string   `yaml:"port"`
		PortTLS            string   `yaml:"portTLS"`
		Timeout            int      `yaml:"timeout"`
		CertCrt            string   `yaml:"certCrt"`
		CertKey            string   `yaml:"certKey"`
		AllowServiceTokens []string `yaml:"allowServiceTokens"`
	} `yaml:"webserver"`
	Service struct {
		KeyLimit int `yaml:"keyLimit"`
	} `yaml:"service"`
	DB struct {
		Host         string `yaml:"host"`
		Port         string `yaml:"port"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		Database     string `yaml:"database"`
		MinOpenConns int32  `yaml:"minOpenConns"`
		MaxOpenConns int32  `yaml:"maxOpenConns"`
	} `yaml:"db"`
	APIService struct {
		Address string `yaml:"address"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"apiservice"`
	Logger struct {
		Level string `yaml:"level"`
		Path  string `yaml:"path"`
		File  string `yaml:"file"`
	} `yaml:"logger"`
}

func GetConfig() (*Config, error) {
	filename := os.Getenv("CONFIG_FILE")
	if filename == "" {
		filename = configFileName
	}

	return readFromFile(filename)
}

func readFromFile(filename string) (*Config, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error open config file: %s", err.Error())
	}

	var cfg Config
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

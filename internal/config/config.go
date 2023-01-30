package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		App App `yaml:"app"`
	}

	App struct {
		Name        string `yaml:"name" env:"APP_NAME"`
		Host        string `yaml:"host" env:"HOST"`
		Port        string `yaml:"port" env:"PORT"`
		Environment string `yaml:"environment" env:"ENVIRONMENT"`
	}
)

func New(configPath string) (*Config, error) {
	var (
		cfg  *Config
		err  error
		once sync.Once
	)

	once.Do(func() {
		cfg, err = parse(configPath)
	})

	return cfg, err
}

func parse(configPath string) (config *Config, err error) {
	filename, err := filepath.Abs(configPath)
	if err != nil {
		return
	}

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	cfg := Config{}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return
	}

	if err = cleanenv.ReadConfig(filename, &cfg); err != nil {
		return
	}

	config = &cfg
	return
}

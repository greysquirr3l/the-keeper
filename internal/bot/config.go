// internal/bot/config.go
package bot

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Discord struct {
		Token        string `yaml:"token"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectURL  string `yaml:"redirect_url"`
	} `yaml:"discord"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
}

func LoadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var cfg Config
	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

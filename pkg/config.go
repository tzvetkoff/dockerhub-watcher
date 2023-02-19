package pkg

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/tzvetkoff-go/errors"
)

// Config ...
type Config struct {
	Docker   DockerConfig           `yaml:"docker"`
	Periodic PeriodicConfig         `yaml:"periodic"`
	Console  ConsoleAnnouncerConfig `yaml:"console"`
	Slack    SlackAnnouncerConfig   `yaml:"slack"`
}

// DockerConfig ...
type DockerConfig struct {
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	DB       string   `yaml:"db"`
	Repos    []string `yaml:"repos"`
}

// PeriodicConfig ...
type PeriodicConfig struct {
	Period int `yaml:"period"`
}

// ConsoleAnnouncerConfig ...
type ConsoleAnnouncerConfig struct {
	Enabled bool `yaml:"enabled"`
	Color   bool `yaml:"color"`
}

// SlackAnnouncerConfig ...
type SlackAnnouncerConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
}

// LoadConfig ...
func LoadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Propagate(err, "could not read config file %s", path)
	}

	result := &Config{}
	err = yaml.Unmarshal(b, &result)
	if err != nil {
		return nil, errors.Propagate(err, "could not unmarshal config file %s", path)
	}

	return result, nil
}

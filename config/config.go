package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFile = "config.yml"

// default values
const (
	defaultUser             = "root"
	defaultContainerName    = "elixir"         // Replace with your container name
	defaultRestartPolicy    = "unless-stopped" // Set to "always", "unless-stopped", or "no"
	defaultEnvFilePath      = "/opt/elixir/validator.env"
	defaultServiceName      = "elixir-updater"
	defaultHost             = "http://localhost"
	defaultPort             = "17690"
	defaultDockerAPIVersion = "1.42"
)

// Config represents app configuration
type Config struct {
	TGBotToken    string `yaml:"tg_bot_token"`
	TGForceChatID int64  `yaml:"tg_force_chat_id"`

	User             string `yaml:"user"`
	ContainerName    string `yaml:"container_name"`
	RestartPolicy    string `yaml:"restart_policy"`
	EnvFilePath      string `yaml:"env_file_path"`
	ServiceName      string `yaml:"service_name"`
	Host             string `yaml:"host"`
	Port             string `yaml:"port"`
	DockerAPIVersion string `yaml:"docker_api_version"`
}

// SetDefaults to the config
func (c *Config) SetDefaults() {
	c.TGBotToken = strings.TrimSpace(c.TGBotToken)
	c.User = strings.TrimSpace(c.User)
	c.ContainerName = strings.TrimSpace(c.ContainerName)
	c.RestartPolicy = strings.TrimSpace(c.RestartPolicy)
	c.EnvFilePath = strings.TrimSpace(c.EnvFilePath)
	c.ServiceName = strings.TrimSpace(c.ServiceName)
	c.Host = strings.TrimSpace(c.Host)
	c.Port = strings.TrimSpace(c.Port)
	c.DockerAPIVersion = strings.TrimSpace(c.DockerAPIVersion)

	if c.User == "" {
		c.User = defaultUser
	}
	if c.ContainerName == "" {
		c.ContainerName = defaultContainerName
	}
	if c.RestartPolicy == "" {
		c.RestartPolicy = defaultRestartPolicy
	}
	if c.EnvFilePath == "" {
		c.EnvFilePath = defaultEnvFilePath
	}
	if c.ServiceName == "" {
		c.ServiceName = defaultServiceName
	}
	if c.Host == "" {
		c.Host = defaultHost
	}
	if c.Port == "" {
		c.Port = defaultPort
	}
	if c.DockerAPIVersion == "" {
		c.DockerAPIVersion = defaultDockerAPIVersion
	}
}

// New initializes new app configuration
func New() (Config, error) {
	var cfg Config
	b, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}

	if err = yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	cfg.SetDefaults()
	return cfg, nil
}

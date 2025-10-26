package config

import (
	"fmt"
	"net/netip"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

type Module struct {
	Username        string        `yaml:"username"`
	UsernameFile    string        `yaml:"username_file"`
	Password        string        `yaml:"password"`
	PasswordFile    string        `yaml:"password_file"`
	Secret          []byte        `yaml:"-"`
	SecretFile      string        `yaml:"secret_file"`
	Timeout         time.Duration `yaml:"-"`
	Retry           time.Duration `yaml:"-"`
	MaxPacketErrors int           `yaml:"max_packet_errors"`
	NasID           string        `yaml:"nas_id"`
	NasIP           netip.Addr    `yaml:"nas_ip"`
}

func (m *Module) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tempModule Module
	temp := struct {
		Secret      string `yaml:"secret"`
		Timeout     int    `yaml:"timeout"`
		Retry       int    `yaml:"retry"`
		*tempModule `yaml:",inline"`
	}{
		Timeout: 5,
		// Default to no retries
		Retry:      0,
		tempModule: (*tempModule)(m),
	}
	m.MaxPacketErrors = 10

	err := unmarshal(&temp)
	if err != nil {
		return err
	}

	m.Timeout = time.Second * time.Duration(temp.Timeout)
	m.Retry = time.Second * time.Duration(temp.Retry)
	m.Secret = []byte(temp.Secret)

	return nil
}

func LoadFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	c := &Config{}
	if err = d.Decode(c); err != nil {
		return nil, err
	}

	for name, module := range c.Modules {
		const envPrefix = "RADIUS_EXPORTER_MODULE_"
		if module.Username == "" && module.UsernameFile == "" {
			envUsername := os.Getenv(envPrefix + name + "_USERNAME")
			if envUsername == "" {
				return nil, fmt.Errorf("%s: username not found in config, env or file", name)
			}
		}
		if module.Password == "" && module.PasswordFile == "" {
			envPassword := os.Getenv(envPrefix + name + "_PASSWORD")
			if envPassword == "" {
				return nil, fmt.Errorf("%s: password not found in config, env or file", name)
			}
		}
		if len(module.Secret) == 0 && module.SecretFile == "" {
			envSecret := os.Getenv(envPrefix + name + "_SECRET")
			if envSecret == "" {
				return nil, fmt.Errorf("%s: secret not found in config, env or file", name)
			}
		}
	}

	return c, nil
}

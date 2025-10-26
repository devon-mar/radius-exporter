package config

import (
	"fmt"
	"net/netip"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

type Module struct {
	Username        string     `yaml:"username"`
	UsernameFile    string     `yaml:"username_file"`
	Password        string     `yaml:"password"`
	PasswordFile    string     `yaml:"password_file"`
	Secret          string     `yaml:"secret"`
	SecretFile      string     `yaml:"secret_file"`
	TimeoutSeconds  uint       `yaml:"timeout"`
	RetrySeconds    uint       `yaml:"retry"`
	MaxPacketErrors int        `yaml:"max_packet_errors"`
	NasID           string     `yaml:"nas_id"`
	NasIP           netip.Addr `yaml:"nas_ip"`
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

		// Set defaults
		if module.TimeoutSeconds == 0 {
			module.TimeoutSeconds = 5
		}
		if module.MaxPacketErrors == 0 {
			module.MaxPacketErrors = 10
		}

		if module.Username == "" && module.UsernameFile == "" {
			module.Username = os.Getenv(envPrefix + name + "_USERNAME")
			if module.Username == "" {
				return nil, fmt.Errorf("%s: username not found in config, env or file", name)
			}
		}
		if module.Password == "" && module.PasswordFile == "" {
			module.Password = os.Getenv(envPrefix + name + "_PASSWORD")
			if module.Password == "" {
				return nil, fmt.Errorf("%s: password not found in config, env or file", name)
			}
		}
		if len(module.Secret) == 0 && module.SecretFile == "" {
			module.Secret = os.Getenv(envPrefix + name + "_SECRET")
			if module.Secret == "" {
				return nil, fmt.Errorf("%s: secret not found in config, env or file", name)
			}
		}
		c.Modules[name] = module
	}

	return c, nil
}

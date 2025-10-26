package config

import (
	"errors"
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

	if temp.Username == "" && temp.UsernameFile == "" {
		return errors.New("username or username_file must be specified")
	}
	if temp.Password == "" && temp.PasswordFile == "" {
		return errors.New("password or password_file must be specified")
	}
	if temp.Secret == "" && temp.SecretFile == "" {
		return errors.New("secret or secret_file must be specified")
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

	return c, nil
}

package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

type Module struct {
	Username        string
	Password        string
	Secret          []byte
	Timeout         time.Duration
	Retry           time.Duration
	MaxPacketErrors int
	NasID           string
	NasIP           net.IP
}

func (m *Module) UnmarshalYAML(unmarshal func(interface{}) error) error {
	temp := struct {
		Username        string `yaml:"username"`
		Password        string `yaml:"password"`
		Secret          string `yaml:"secret"`
		Timeout         int    `yaml:"timeout"`
		Retry           int    `yaml:"retry"`
		MaxPacketErrors int    `yaml:"max_packet_errors"`
		NasID           string `yaml:"nas_id"`
		NasIP           string `yaml:"nas_ip"`
	}{
		Timeout: 5,
		// Default to no retries
		Retry:           0,
		MaxPacketErrors: 10,
	}

	err := unmarshal(&temp)
	if err != nil {
		return err
	}

	if temp.Username == "" {
		return errors.New("username must not be empty")
	}
	if temp.Password == "" {
		return errors.New("password must not be empty")
	}
	if temp.Secret == "" {
		return errors.New("secret must not be empty")
	}

	m.Username = temp.Username
	m.Password = temp.Password
	m.Timeout = time.Second * time.Duration(temp.Timeout)
	m.Retry = time.Second * time.Duration(temp.Retry)
	m.MaxPacketErrors = temp.MaxPacketErrors
	m.Secret = []byte(temp.Secret)
	m.NasID = temp.NasID
	if temp.NasIP != "" {
		m.NasIP = net.ParseIP(temp.NasIP)
		if m.NasIP == nil {
			return fmt.Errorf("ip '%s' is not a valid IP", temp.NasIP)
		}
	}

	return nil
}

func LoadFromFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	err = yaml.UnmarshalStrict(b, c)
	if err != nil {
		slog.Error("Error unmarshalling yaml.", "err", err)
		return nil, err
	}

	slog.Info("Loaded config successfully.")
	return c, nil
}

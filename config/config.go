package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v2"
)

var (
	DefaultModule = rawModule{Timeout: 5}
)

type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

type Module struct {
	Username string
	Password string
	Secret   []byte
	Timeout  time.Duration
	NasID    string
	NasIP    net.IP
}

// A temporary Module to unmarshal into.
type rawModule struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Secret   string `yaml:"secret"`
	Timeout  int    `yaml:"timeout"`
	NasID    string `yaml:"nas_id"`
	NasIP    string `yaml:"nas_ip"`
}

func (m *Module) UnmarshalYAML(unmarshal func(interface{}) error) error {
	temp := DefaultModule

	err := unmarshal(&temp)

	if err != nil {
		return err
	}

	if temp.Username == "" {
		return errors.New("Username must not be empty")
	}
	if temp.Password == "" {
		return errors.New("Password must not be empty")
	}
	if temp.Secret == "" {
		return errors.New("Secret must not be empty")
	}

	m.Username = temp.Username
	m.Password = temp.Password
	m.Timeout = time.Second * time.Duration(temp.Timeout)
	m.Secret = []byte(temp.Secret)
	m.NasID = temp.NasID
	if temp.NasIP != "" {
		m.NasIP = net.ParseIP(temp.NasIP)
		if m.NasIP == nil {
			return fmt.Errorf("'%s' is not a valid IP.", temp.NasIP)
		}
	}

	return nil
}

func LoadFromFile(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	err = yaml.UnmarshalStrict(b, c)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error unmarshalling yaml.")
		return nil, err
	}

	log.Info("Loaded config successfully.")
	return c, nil
}

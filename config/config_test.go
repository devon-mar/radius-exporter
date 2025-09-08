package config

import (
	"fmt"
	"net/netip"
	"os"
	"testing"
	"time"
)

func TestValidConfig(t *testing.T) {
	os.Setenv("TEST_PASSWORD", "password")
	os.Setenv("TEST_SECRET", "s3cr3t")
	config, err := LoadFromFile("testdata/valid.yml")
	if err != nil {
		t.Errorf("Expected no error from LoadFromFile() got %s", err)
	}
	m1, ok := config.Modules["m1"]
	if !ok {
		t.Errorf("Config does not contain module 'm1'")
	}
	if m1.Username != "user" {
		t.Errorf("Username: have %s, want %s", m1.Username, "user")
	}
	if m1.Password != "password" {
		t.Errorf("Password: have %s, want %s", m1.Password, "password")
	}
	if have := (string)(m1.Secret); have != "s3cr3t" {
		t.Errorf("Secret: have %s, want %s", have, "s3cr3t")
	}
	if m1.NasID != "nas_id" {
		t.Errorf("NasID: have %s, want %s", m1.NasID, "nas_id")
	}
	if expected := netip.MustParseAddr("192.0.2.1"); m1.NasIP != expected {
		t.Errorf("NasIP: have %s, want %s", m1.NasIP, expected)
	}
	if want := time.Duration(5) * time.Second; m1.Timeout != want {
		t.Errorf("Timeout: have %s, want %s", m1.Timeout, want)
	}

	m2, ok := config.Modules["m2"]
	if !ok {
		t.Errorf("Config does not contain module 'm2'")
	}
	if want := time.Duration(7) * time.Second; m2.Timeout != want {
		t.Errorf("Timeout: have %s, want %s", m2.Timeout, want)
	}

	m3, ok := config.Modules["m3"]
	if !ok {
		t.Errorf("Config does not contain module 'm3'")
	}
	if m3.Password != "password" {
		t.Errorf("Password: have %s, want %s", m3.Password, "password")
	}
	if have := (string)(m3.Secret); have != "s3cr3t" {
		t.Errorf("Secret: have %s, want %s", have, "s3cr3t")
	}

	if c := len(config.Modules); c != 3 {
		t.Errorf("Module count: have %d, want %d", c, 3)
	}
}

func TestInvalid(t *testing.T) {
	os.Setenv("SET_SECRET", "s3cr3t")
	os.Setenv("SET_PASSWORD", "password")
	for i := 0; i < 6; i++ {
		file := fmt.Sprintf("testdata/invalid%d.yml", i)

		t.Run(file, func(t *testing.T) {
			_, err := LoadFromFile(file)
			if err == nil {
				t.Errorf("Expected error from config %s", file)
			}
		})
	}
}

package config

import (
	"net/netip"
	"path"
	"reflect"
	"testing"
)

func TestValidConfig(t *testing.T) {
	config, err := LoadFromFile("testdata/valid.yml")
	if err != nil {
		t.Errorf("Expected no error from LoadFromFile() got %s", err)
	}

	expected := Config{
		Modules: map[string]Module{
			"m1": {
				Username:        "user",
				Password:        "password",
				Secret:          "s3cr3t",
				TimeoutSeconds:  5,
				MaxPacketErrors: 10,
				NasID:           "nas_id",
				NasIP:           netip.MustParseAddr("192.0.2.1"),
			},
			"m2": {
				Username:        "user",
				Password:        "password",
				Secret:          "s3cr3t",
				TimeoutSeconds:  7,
				MaxPacketErrors: 10,
				NasID:           "nas_id",
				NasIP:           netip.MustParseAddr("192.0.2.1"),
			},
		},
	}
	if !reflect.DeepEqual(&expected, config) {
		t.Errorf("expected config %#v, got %#v", expected, config)
	}
}

func TestConfigEnvSecrets(t *testing.T) {
	const username = "testuser"
	const password = "password"
	const secret = "secret123"
	t.Setenv("RADIUS_EXPORTER_MODULE_m1_USERNAME", username)
	t.Setenv("RADIUS_EXPORTER_MODULE_m1_PASSWORD", password)
	t.Setenv("RADIUS_EXPORTER_MODULE_m1_SECRET", secret)
	config, err := LoadFromFile("testdata/env.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m1 := config.Modules["m1"]

	if m1.Username != username {
		t.Errorf("expected username %q, got %q", username, m1.Username)
	}
	if m1.Password != password {
		t.Errorf("expected password %q, got %q", password, m1.Password)
	}
	if string(m1.Secret) != secret {
		t.Errorf("expected password %q, got %q", secret, m1.Secret)
	}
}

func TestInvalid(t *testing.T) {
	tests := []string{
		"invalid0.yml",
		"invalid1.yml",
		"invalid2.yml",
		"invalid3.yml",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			file := path.Join("testdata", test)

			_, err := LoadFromFile(file)
			if err == nil {
				t.Errorf("Expected error from config %s", file)
			}
		})
	}
}

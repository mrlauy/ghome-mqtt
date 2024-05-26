package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	_ = os.Setenv("CONFIG_FILE", "../config.example.yaml")
	cfg, err := ReadConfig()
	require.NoError(t, err)

	// Expected configuration
	expectedConfig := &Config{
		Server: ServerConfig{Port: 9000, Host: "localhost"},
		Log:    Log{Level: "DEBUG"},
		Auth: AuthConfig{
			Client: struct {
				Id     string `yaml:"id" env:"CLIENT_ID" env-default:"000000"`
				Secret string `yaml:"secret" env:"CLIENT_SECRET" env-default:"999999"`
				Domain string `yaml:"domain" env:"CLIENT_DOMAIN" env-default:"https://oauth-redirect.googleusercontent.com/r/project/project-id"`
			}{
				Id:     "000000",
				Secret: "999999",
				Domain: "https://oauth-redirect.googleusercontent.com/r/project/project-id",
			},
			Credientials: ".credentials",
			TokenStore:   ".tokenstore",
		},
		Mqtt: MqttConfig{
			Host:     "192.168.1.10",
			Port:     1883,
			Username: "",
			Password: "",
			Tls:      false,
		},
		Devices: map[string]DeviceConfig{
			"plug": {
				Name:            "plug",
				Topic:           "zigbee2mqtt/plug/set",
				Subscription:    "zigbee2mqtt/plug",
				Type:            "action.devices.types.OUTLET",
				WillReportState: false,
				Attributes:      SyncAttributes{},
				Traits:          []string{"action.devices.commands.OnOff"},
			},
		},
		ExecutionTemplates: map[string]string{"action.devices.commands.OnOff": "{\"state\":\"%s\"}"},
	}

	t.Logf("config: %v", cfg)

	require.NoError(t, err)
	assert.Equal(t, expectedConfig.Server, cfg.Server)
	assert.Equal(t, expectedConfig.Log.Level, cfg.Log.Level)
	assert.Equal(t, expectedConfig.Auth, cfg.Auth)
	assert.Equal(t, expectedConfig.Mqtt, cfg.Mqtt)
	assert.Equal(t, expectedConfig.Devices, cfg.Devices)
	assert.Equal(t, expectedConfig.ExecutionTemplates, cfg.ExecutionTemplates)
}

func TestParseConfigTemplates(t *testing.T) {
	yamlContent := `
templates:
  action.command.one: template1
  action.command.two: '{"state":"%s"}'
`

	// Expected configuration
	expectedConfig := &Config{
		ExecutionTemplates: map[string]string{
			"action.command.one": "template1",
			"action.command.two": `{"state":"%s"}`,
		},
	}

	cleanUp := createTempConfig(t, yamlContent)
	defer cleanUp()

	config, err := ReadConfig()

	require.NoError(t, err)
	assert.Equal(t, expectedConfig.ExecutionTemplates, config.ExecutionTemplates)
}

func createTempConfig(t *testing.T, yamlContent string) func() {
	tempFile, err := os.CreateTemp("", "*test-config.yaml")
	if err != nil {
		t.Fatalf("error creating temporary YAML file: %v", err)
	}
	defer tempFile.Close()
	cleanUp := func() {
		_ = os.Remove(tempFile.Name())
	}

	_, err = tempFile.WriteString(yamlContent)
	if err != nil {
		t.Fatalf("error writing temporary YAML file: %v", err)
	}

	err = os.Setenv("CONFIG_FILE", tempFile.Name())
	if err != nil {
		t.Fatalf("Error setting config variable %s: %v", tempFile.Name(), err)
	}

	return cleanUp
}

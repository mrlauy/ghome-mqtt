package config

import (
	"os"
	"testing"

	"github.com/mrlauy/ghome-mqtt/fullfillment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	yamlContent := `
templates:
  device1:
    action.command.one: template1
  device2:
    action.command.two: '{"state":"%s"}'
`

	// Expected configuration
	expectedConfig := &Config{
		ExecutionTemplates: fullfillment.ExecutionTemplates{
			"device1": {
				"action.command.one": "template1",
			},
			"device2": {
				"action.command.two": `{"state":"%s"}`,
			},
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
		os.Remove(tempFile.Name())
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

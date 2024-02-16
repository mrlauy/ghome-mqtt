//go:build all || it
// +build all it

package main

import (
	"github.com/mrlauy/ghome-mqtt/config"
	mqtt2 "github.com/mrlauy/ghome-mqtt/mqtt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendMessageLive(t *testing.T) {
	mqtt, err := mqtt2.NewMqtt(config.MqttConfig{
		Host: "10.0.0.21",
		Port: 1883,
	})

	require.NoError(t, err)
	mqtt.SendMessage("device/speaker/set", `{"state":"on"}`)
}

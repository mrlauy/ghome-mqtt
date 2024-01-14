package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendMessageLive(t *testing.T) {
	mqtt, err := NewMqtt(MqttConfig{
		Host: "10.0.0.21",
		Port: 1883,
	})

	require.NoError(t, err)
	mqtt.SendMessage("device/speaker/set", `{"state":"on"}`)
}

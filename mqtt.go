package main

import (
	"fmt"
	log "log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	client mqtt.Client
}

func NewMqtt(cfg MqttConfig) (*Mqtt, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Host, cfg.Port))
	opts.SetClientID("ghome-client")
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)

	if cfg.Tls {
		return nil, fmt.Errorf("TLS not yet supported")
	}

	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &Mqtt{
		client: client,
	}, nil
}

func (m *Mqtt) SendMessage(topic string, message string) {
	log.Info("send mqtt message", "topic", topic, "message", message)
	token := m.client.Publish(topic, 0, false, message)
	token.Wait()
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Info("received message", "topic", msg.Topic(), "message", msg.Payload())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Info("mqtt client connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Info("mqtt client connect lost", err)
}

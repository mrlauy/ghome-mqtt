package mqtt

import (
	"encoding/json"
	"fmt"
	"github.com/mrlauy/ghome-mqtt/config"
	log "log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var publishTimeout = 1 * time.Second

type Mqtt struct {
	client mqtt.Client
}

func NewMqtt(cfg config.MqttConfig) (*Mqtt, error) {
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

func (m *Mqtt) RegisterStateChangeListener(device string, topic string, callback func(string, map[string]interface{})) error {
	log.Info("subscribe to topic", "device", device, "topic", topic)
	callbackHandler := func(client mqtt.Client, msg mqtt.Message) {
		log.Info("message published", "device", device, "topic", msg.Topic(), "message", msg.Payload())

		payload := make(map[string]interface{})
		err := json.Unmarshal(msg.Payload(), &payload)
		if err != nil {
			log.Error("fail to marshal incoming message", "device", device, "topic", topic, err)
			return
		}

		callback(device, payload)
	}

	if token := m.client.Subscribe(topic, 0, callbackHandler); token.Wait() && token.Error() != nil {
		log.Error("failed to subscribe", "device", device, "topic", topic, token.Error())
		return token.Error()
	}
	return nil
}

func (m *Mqtt) SendMessage(topic string, message string) {
	log.Info("send mqtt message", "topic", topic, "message", message)
	token := m.client.Publish(topic, 0, false, message)
	if !token.WaitTimeout(publishTimeout) {
		log.Error("failed to publish message due to timeout", "topic", topic, "message", message, "token-error", token.Error())
	}
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Info("received message", "topic", msg.Topic(), "message", msg.Payload())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Debug("mqtt client connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Debug("mqtt client connect lost", err)
}

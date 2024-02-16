package mqtt

import (
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestSendMessage(t *testing.T) {
	mqtt := Mqtt{
		client: &mqttClientMock{},
	}

	mqtt.SendMessage("device/test", `{"state":"on"}`)
}

type mqttClientMock struct {
}

func (m *mqttClientMock) IsConnected() bool       { return true }
func (m *mqttClientMock) IsConnectionOpen() bool  { return true }
func (m *mqttClientMock) Connect() mqtt.Token     { return &mqtt.DummyToken{} }
func (m *mqttClientMock) Disconnect(quiesce uint) {}
func (m *mqttClientMock) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientMock) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientMock) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientMock) Unsubscribe(topics ...string) mqtt.Token             { return &mqtt.DummyToken{} }
func (m *mqttClientMock) AddRoute(topic string, callback mqtt.MessageHandler) {}
func (m *mqttClientMock) OptionsReader() mqtt.ClientOptionsReader             { return mqtt.ClientOptionsReader{} }

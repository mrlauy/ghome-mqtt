package main

import (
	"fmt"
	log "log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/mrlauy/ghome-mqtt/fullfillment"
)

type Config struct {
	Server             ServerConfig                    `yaml:"server"`
	Auth               AuthConfig                      `yaml:"auth"`
	Mqtt               MqttConfig                      `yaml:"mqtt"`
	Devices            DevicesConfig                   `yaml:"devices"`
	ExecutionTemplates fullfillment.ExecutionTemplates `yaml:"templates"`
}

type ServerConfig struct {
	Port int    `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
	Host string `yaml:"host" env:"host" env-default:"localhost"`
}

type AuthConfig struct {
	Client struct {
		Id     string `yaml:"id" env:"CLIENT_ID" env-default:"000000"`
		Secret string `yaml:"secret" env:"CLIENT_SECRET" env-default:"999999"`
	} `yaml:"client"`
}

type MqttConfig struct {
	Host     string `yaml:"host" env:"MQTT_BROKER_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"MQTT_BROKER_PORT" env-default:"1883"`
	Username string `yaml:"username" env:"MQTT_BROKER_USERNAME"`
	Password string `yaml:"password" env:"MQTT_BROKER_PASSWORD"`
	Tls      bool   `yaml:"tls" env:"MQTT_BROKER_TLS" env-default:"false"`
}

type DevicesConfig struct {
	File string `yaml:"file" env:"DEVICES_FILE" env-default:"devices.json"`
}

func ReadConfig() (*Config, error) {
	filename := getenv("CONFIG_FILE", "config.yaml")
	var cfg Config
	err := cleanenv.ReadConfig(filename, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", filename, err)
	}

	log.Info("read config", "config", cfg)
	return &cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

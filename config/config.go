package config

import (
	"fmt"
	log "log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server             ServerConfig            `yaml:"server"`
	Auth               AuthConfig              `yaml:"auth"`
	Mqtt               MqttConfig              `yaml:"mqtt"`
	Devices            map[string]DeviceConfig `yaml:"devices"`
	ExecutionTemplates map[string]string       `yaml:"templates"`
}

type ServerConfig struct {
	Port int    `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
}

type AuthConfig struct {
	Client struct {
		Id     string `yaml:"id" env:"CLIENT_ID" env-default:"000000"`
		Secret string `yaml:"secret" env:"CLIENT_SECRET" env-default:"999999"`
		Domain string `yaml:"domain" env:"CLIENT_DOMAIN" env-default:"https://oauth-redirect.googleusercontent.com/r/project/project-id"`
	} `yaml:"client"`
	Credientials string `yaml:"credentials" env:"CREDENTIALS" env-default:".credentials"`
	TokenStore   string `yaml:"tokenStore" env:"TOKEN_STORE" env-default:".tokenstore"`
}

type MqttConfig struct {
	Host     string `yaml:"host" env:"MQTT_BROKER_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"MQTT_BROKER_PORT" env-default:"1883"`
	Username string `yaml:"username" env:"MQTT_BROKER_USERNAME"`
	Password string `yaml:"password" env:"MQTT_BROKER_PASSWORD"`
	Tls      bool   `yaml:"tls" env:"MQTT_BROKER_TLS" env-default:"false"`
}

type DeviceConfig struct {
	Name            string         `yaml:"name"`
	Topic           string         `yaml:"topic"`
	Type            string         `yaml:"type"`
	WillReportState bool           `yaml:"willReportState"`
	Attributes      SyncAttributes `yaml:"attributes"`
	Traits          []string       `yaml:"traits"`
}

type SyncAttributes struct {
	// action.devices.traits.ColorSetting
	ColorModel              string                    `yaml:"colorModel" json:"colorModel,omitempty"`
	ColorTemperatureRange   SyncColorTemperatureRange `yaml:"colorTemperatureRange" json:"colorTemperatureRange,omitempty"`
	CommandOnlyColorSetting bool                      `yaml:"commandOnlyColorSetting" json:"commandOnlyColorSetting,omitempty"`
	// action.devices.traits.OnOff
	CommandOnlyOnOff bool `yaml:"commandOnlyOnOff" json:"commandOnlyOnOff,omitempty"`
	QueryOnlyOnOff   bool `yaml:"queryOnlyOnOff" json:"queryOnlyOnOff,omitempty"`
	// action.devices.traits.TransportControl
	TransportControlSupportedCommands []string `yaml:"transportControlSupportedCommands" json:"transportControlSupportedCommands,omitempty"`
	// action.devices.traits.Volume
	VolumeMaxLevel         bool `yaml:"volumeMaxLevel" json:"volumeMaxLevel,omitempty"`
	VolumeCanMuteAndUnmute bool `yaml:"volumeCanMuteAndUnmute" json:"volumeCanMuteAndUnmute,omitempty"`
}

type SyncColorTemperatureRange struct {
	TemperatureMinK int `yaml:"temperatureMinK" json:"temperatureMinK,omitempty"`
	TemperatureMaxK int `yaml:"temperatureMaxK" json:"temperatureMaxK,omitempty"`
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

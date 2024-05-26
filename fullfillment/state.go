package fullfillment

import (
	"fmt"
	log "log/slog"
)

// type DeviceConfig struct {
// 	Devices []Device
// }

// type Device struct {
// 	ID         string     `json:"id"`
// 	Type       string     `json:"type"`
// 	Name       Name       `json:"name"`
// 	DeviceInfo DeviceInfo `json:"deviceInfo,omitempty"`
// 	Topic      string     `json:"topic"`
// 	Traits     []string   `json:"traits"`
// }

// type Name struct {
// 	DefaultNames []any  `json:"defaultNames,omitempty"`
// 	Name         string `json:"name,omitempty"`
// 	Nicknames    []any  `json:"nicknames,omitempty"`
// }

// type DeviceInfo struct {
// 	Manufacturer string `json:"manufacturer,omitempty"`
// 	Model        string `json:"model,omitempty"`
// 	HwVersion    string `json:"hwVersion,omitempty"`
// 	SwVersion    string `json:"swVersion,omitempty"`
// }

func (f *Fullfillment) setState(deviceId string, payload map[string]interface{}) {
	if _, ok := payload["state"]; !ok {
		log.Info("failed to get state for device", "device", deviceId, "payload", payload)
		return
	}
	state := fmt.Sprintf("%v", payload["state"])
	if state != "OFF" && state != "ON" {
		log.Info("failed to get state for device", "device", deviceId, "payload", payload)
		return
	}

	device := f.devices[deviceId]
	device.State = LocalState{
		State: state,
		On:    state == "ON",
	}
	log.Info("change state", "device", device, "old", f.devices[deviceId].State, "new", device.State)
	f.devices[deviceId] = device
}

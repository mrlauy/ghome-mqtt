package fullfillment

import (
	log "log/slog"
)

/*
{
  "requestId": "ff36a3cc-ec34-11e6-b1a0-64510650abcf",
  "payload": {
    "agentUserId": "1836.15267389",
    "devices": [
      {
        "id": "123",
        "type": "action.devices.types.OUTLET",
        "traits": [
          "action.devices.traits.OnOff"
        ],
        "name": {
          "defaultNames": [
            "My Outlet 1234"
          ],
          "name": "Night light",
          "nicknames": [
            "wall plug"
          ]
        },
        "willReportState": false,
        "roomHint": "kitchen",
        "deviceInfo": {
          "manufacturer": "lights-out-inc",
          "model": "hs1234",
          "hwVersion": "3.2",
          "swVersion": "11.4"
        },
        "otherDeviceIds": [
          {
            "deviceId": "local-device-id"
          }
        ],
        "customData": {
          "fooValue": 74,
          "barValue": true,
          "bazValue": "foo"
        }
      },
      {
        "id": "456",
        "type": "action.devices.types.LIGHT",
        "traits": [
          "action.devices.traits.OnOff",
          "action.devices.traits.Brightness",
          "action.devices.traits.ColorSetting"
        ],
        "name": {
          "defaultNames": [
            "lights out inc. bulb A19 color hyperglow"
          ],
          "name": "lamp1",
          "nicknames": [
            "reading lamp"
          ]
        },
        "willReportState": false,
        "roomHint": "office",
        "attributes": {
          "colorModel": "rgb",
          "colorTemperatureRange": {
            "temperatureMinK": 2000,
            "temperatureMaxK": 9000
          },
          "commandOnlyColorSetting": false
        },
        "deviceInfo": {
          "manufacturer": "lights out inc.",
          "model": "hg11",
          "hwVersion": "1.2",
          "swVersion": "5.4"
        },
        "customData": {
          "fooValue": 12,
          "barValue": false,
          "bazValue": "bar"
        }
      }
    ]
  }
}
*/

type SyncResponse struct {
	RequestID string      `json:"requestId"` // Required. ID of the corresponding request.
	Payload   SyncPayload `json:"payload"`   // Required. Intent response payload.
}

type SyncPayload struct {
	AgentUserID string        `json:"agentUserId"`           //  Required. Reflects the unique (and immutable) user ID on the agent's platform. The string is opaque to Google, so if there's an immutable form vs a mutable form on the agent side, use the immutable form (e.g. an account number rather than email).
	Devices     []SyncDevices `json:"devices"`               // Required. List of devices owned by the user. Zero or more devices are returned (zero devices meaning the user has no devices, or has disconnected them all).
	ErrorCode   string        `json:"errorCode,omitempty"`   // For systematic errors on SYNC
	DebugString string        `json:"debugString,omitempty"` // Detailed error which will never be presented to users but may be logged or used during development.
}

type SyncDevices struct {
	ID                           string                `json:"id"`                                     // Required. The ID of the device in the developer's cloud. This must be unique for the user and for the developer, as in cases of sharing we may use this to dedupe multiple views of the same device. It should be immutable for the device; if it changes, the Assistant will treat it as a new device.
	Type                         string                `json:"type"`                                   // Required. The hardware type of device.
	Traits                       []string              `json:"traits"`                                 // Required. List of traits this device has. This defines the commands, attributes, and states that the device supports.
	Name                         SyncName              `json:"name"`                                   // Required. Names of this device.
	WillReportState              bool                  `json:"willReportState"`                        // Required.	Indicates whether this device will have its states updated by the Real Time Feed. (true to use the Real Time Feed for reporting state, and false to use the polling model.)
	RoomHint                     string                `json:"roomHint,omitempty"`                     // Provides the current room of the device in the user's home to simplify setup.
	NotificationSupportedByAgent bool                  `json:"notificationSupportedByAgent,omitempty"` // (Default: false) Indicates whether notifications are enabled for the device.
	DeviceInfo                   *SyncDeviceInfo       `json:"deviceInfo,omitempty"`                   // Contains fields describing the device for use in one-off logic if needed (e.g. 'broken firmware version X of light Y requires adjusting color', or 'security flaw requires notifying all users of firmware Z').
	OtherDeviceIds               []*SyncOtherDeviceIds `json:"otherDeviceIds,omitempty"`               // List of alternate IDs used to identify a cloud synced device for local execution.
	CustomData                   *SyncCustomData       `json:"customData,omitempty"`                   // Object defined by the developer which will be attached to future QUERY and EXECUTE requests, maximum of 512 bytes per device. Use this object to store additional information about the device your cloud service may need, such as the global region of the device. Data in this object has a few constraints: No sensitive information, including but not limited to Personally Identifiable Information.
	Attributes                   *SyncAttributes       `json:"attributes,omitempty"`                   // Aligned with per-trait attributes described in each trait schema reference.
}

type SyncName struct {
	DefaultNames []string `json:"defaultNames,omitempty"` // List of names provided by the developer rather than the user, often manufacturer names, SKUs, etc.
	Name         string   `json:"name"`                   // Required. Primary name of the device, generally provided by the user. This is also the name the Assistant will prefer to describe the device in responses.
	Nicknames    []string `json:"nicknames,omitempty"`    // Additional names provided by the user for the device.
}

type SyncDeviceInfo struct {
	Manufacturer string `json:"manufacturer,omitempty"` // Especially useful when the developer is a hub for other devices. Google may provide a standard list of manufacturers here so that e.g. TP-Link and Smartthings both describe 'osram' the same way.
	Model        string `json:"model,omitempty"`        // The model or SKU identifier of the particular device.
	HwVersion    string `json:"hwVersion,omitempty"`    // Specific version number attached to the hardware if available.
	SwVersion    string `json:"swVersion,omitempty"`    // Specific version number attached to the software/firmware, if available.
}

type SyncOtherDeviceIds struct {
	AgentId  string `json:"agentId,omitempty"` // The agent's ID. Generally, this is the project ID in the Actions console.
	DeviceID string `json:"deviceId"`          // Required. Device ID defined by the agent. The device ID must be unique.
}

type SyncCustomData struct {
	FooValue int    `json:"fooValue,omitempty"`
	BarValue bool   `json:"barValue,omitempty"`
	BazValue string `json:"bazValue,omitempty"`
}

type SyncAttributes struct {
	ColorModel              string                    `json:"colorModel,omitempty"`
	ColorTemperatureRange   SyncColorTemperatureRange `json:"colorTemperatureRange,omitempty"`
	CommandOnlyColorSetting bool                      `json:"commandOnlyColorSetting,omitempty"`
}

type SyncColorTemperatureRange struct {
	TemperatureMinK int `json:"temperatureMinK,omitempty"`
	TemperatureMaxK int `json:"temperatureMaxK,omitempty"`
}

func (f *Fullfillment) sync(request FullfillementRequest, userId string) SyncResponse {
	requestId := request.RequestID
	log.Info("handle sync", "request", requestId, "user", userId)
	return SyncResponse{
		RequestID: requestId,
		Payload: SyncPayload{
			AgentUserID: userId,
			Devices:     f.devices,
		},
	}
}

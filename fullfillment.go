package main

import "net/http"

type SyncRequest struct {
	RequestID string `json:"requestId"`
	Inputs    []struct {
		Intent string `json:"intent"`
	} `json:"inputs"`
}

type SyncResponse struct {
	RequestID string `json:"requestId"`
	Payload   struct {
		AgentUserID string `json:"agentUserId"`
		Devices     []struct {
			ID     string   `json:"id"`
			Type   string   `json:"type"`
			Traits []string `json:"traits"`
			Name   struct {
				DefaultNames []string `json:"defaultNames"`
				Name         string   `json:"name"`
				Nicknames    []string `json:"nicknames"`
			} `json:"name"`
			WillReportState bool   `json:"willReportState"`
			RoomHint        string `json:"roomHint"`
			DeviceInfo      struct {
				Manufacturer string `json:"manufacturer"`
				Model        string `json:"model"`
				HwVersion    string `json:"hwVersion"`
				SwVersion    string `json:"swVersion"`
			} `json:"deviceInfo"`
			OtherDeviceIds []struct {
				DeviceID string `json:"deviceId"`
			} `json:"otherDeviceIds,omitempty"`
			CustomData struct {
				FooValue int    `json:"fooValue"`
				BarValue bool   `json:"barValue"`
				BazValue string `json:"bazValue"`
			} `json:"customData"`
			Attributes struct {
				ColorModel            string `json:"colorModel"`
				ColorTemperatureRange struct {
					TemperatureMinK int `json:"temperatureMinK"`
					TemperatureMaxK int `json:"temperatureMaxK"`
				} `json:"colorTemperatureRange"`
				CommandOnlyColorSetting bool `json:"commandOnlyColorSetting"`
			} `json:"attributes,omitempty"`
		} `json:"devices"`
	} `json:"payload"`
}

type QueryRequest struct {
	RequestID string `json:"requestId"`
	Inputs    []struct {
		Intent  string `json:"intent"`
		Payload struct {
			Devices []struct {
				ID         string `json:"id"`
				CustomData struct {
					FooValue int    `json:"fooValue"`
					BarValue bool   `json:"barValue"`
					BazValue string `json:"bazValue"`
				} `json:"customData"`
			} `json:"devices"`
		} `json:"payload"`
	} `json:"inputs"`
}

type QueryResponse struct {
	RequestID string `json:"requestId"`
	Payload   struct {
		Devices struct {
			Num123 struct {
				On     bool `json:"on"`
				Online bool `json:"online"`
			} `json:"123"`
			Num456 struct {
				On         bool `json:"on"`
				Online     bool `json:"online"`
				Brightness int  `json:"brightness"`
				Color      struct {
					Name        string `json:"name"`
					SpectrumRGB int    `json:"spectrumRGB"`
				} `json:"color"`
			} `json:"456"`
		} `json:"devices"`
	} `json:"payload"`
}

type ExecuteRequest struct {
	RequestID string `json:"requestId"`
	Inputs    []struct {
		Intent  string `json:"intent"`
		Payload struct {
			Commands []struct {
				Devices []struct {
					ID         string `json:"id"`
					CustomData struct {
						FooValue int    `json:"fooValue"`
						BarValue bool   `json:"barValue"`
						BazValue string `json:"bazValue"`
					} `json:"customData"`
				} `json:"devices"`
				Execution []struct {
					Command string `json:"command"`
					Params  struct {
						On bool `json:"on"`
					} `json:"params"`
				} `json:"execution"`
			} `json:"commands"`
		} `json:"payload"`
	} `json:"inputs"`
}

type ExecuteResponse struct {
	RequestID string `json:"requestId"`
	Payload   struct {
		Commands []struct {
			Ids    []string `json:"ids"`
			Status string   `json:"status"`
			States struct {
				On     bool `json:"on"`
				Online bool `json:"online"`
			} `json:"states,omitempty"`
			ErrorCode string `json:"errorCode,omitempty"`
		} `json:"commands"`
	} `json:"payload"`
}

type DisconnectRequest struct {
	RequestID string `json:"requestId"`
	Inputs    []struct {
		Intent string `json:"intent"`
	}
}

type DisconnectResponse struct {
}

func FullfillmentHandler(w http.ResponseWriter, r *http.Request) {

}

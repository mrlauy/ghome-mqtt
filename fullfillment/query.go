package fullfillment

import log "log/slog"

type QueryResponse struct {
	RequestID string       `json:"requestId,omitempty"` // Required. ID of the corresponding request.
	Payload   QueryPayload `json:"payload,omitempty"`   // Required. Intent response payload.
}

type QueryPayload struct {
	Devices     map[string]QueryDevice `json:"devices,omitempty"`     // Required. Map of devices. Maps developer device ID to object of state properties.
	ErrorCode   string                 `json:"errorCode,omitempty"`   // An error code for the entire transaction for auth failures and developer system unavailability. For individual device errors use the errorCode within the device object.
	DebugString string                 `json:"debugString,omitempty"` // Detailed error which will never be presented to users but may be logged or used during development.
}

type QueryDevice struct {
	Online    bool   `json:"online,omitempty"`    // Required. Indicates if the device is online (that is, reachable) or not.
	Status    string `json:"status,omitempty"`    // Required. Result of the query operation. Supported values: SUCCESS Confirm that the query succeeded. OFFLINE Target device is in offline state or unreachable. EXCEPTIONS There is an issue or alert associated with a query. The query could succeed or fail. This status type is typically set when you want to send additional information about another connected device. ERROR Unable to query the target device.
	ErrorCode string `json:"errorCode,omitempty"` // Expanding ERROR state if needed from the preset error codes, which will map to the errors presented to users.

	On         bool   `json:"on,omitempty"`
	Brightness int    `json:"brightness,omitempty"`
	Color      *Color `json:"color,omitempty"`
}

type Color struct {
	SpectrumRgb int `json:"spectrumRgb,omitempty"`
}

func (f *Fullfillment) query(requestId string, payload PayloadRequest) QueryResponse {
	log.Info("handle sync request", "request", requestId, "payload", payload)
	devices := map[string]QueryDevice{}
	for _, device := range payload.Devices {
		devices[device.ID] = QueryDevice{
			Online: true,
			On:     f.state[device.ID].On,
		}
	}

	return QueryResponse{
		RequestID: requestId,
		Payload: QueryPayload{
			Devices: devices,
		},
	}
}

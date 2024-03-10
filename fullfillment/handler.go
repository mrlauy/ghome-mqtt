package fullfillment

import (
	"encoding/json"
	"fmt"
	"github.com/mrlauy/ghome-mqtt/config"
	log "log/slog"
	"net/http"
	"strings"
)

type FullfillementRequest struct {
	RequestID string         `json:"requestId,omitempty"`
	Inputs    []InputRequest `json:"inputs,omitempty"`
}

type InputRequest struct {
	Intent  string         `json:"intent,omitempty"`
	Payload PayloadRequest `json:"payload,omitempty"`
}

type PayloadRequest struct {
	AgentUserID string           `json:"agentUserId,omitempty"`
	Devices     []DeviceRequest  `json:"devices,omitempty"`
	Commands    []CommandRequest `json:"commands,omitempty"`
}

type DeviceRequest struct {
	ID         string             `json:"id,omitempty"`
	CustomData *CustomDataRequest `json:"customData,omitempty"`
}

type CustomDataRequest struct {
	FooValue int    `json:"fooValue,omitempty"`
	BarValue bool   `json:"barValue,omitempty"`
	BazValue string `json:"bazValue,omitempty"`
}

type CommandRequest struct {
	Devices   []DeviceRequest    `json:"devices,omitempty"`
	Execution []ExecutionRequest `json:"execution,omitempty"`
}

type ExecutionRequest struct {
	Command string        `json:"command,omitempty"`
	Params  ParamsRequest `json:"params,omitempty"`
}

type ParamsRequest struct {
	On bool `json:"on,omitempty"`
	// action.devices.traits.Volume
	Mute          bool `json:"mute,omitempty"`
	VolumeLevel   int  `json:"volumeLevel,omitempty"`
	RelativeSteps int  `json:"relativeSteps,omitempty"`
}

type EmptyResponse struct {
}

type Device struct {
	Topic string
	State LocalState
}
type LocalState struct {
	State        string
	On           bool
	DebugCommand []string
}

type Fullfillment struct {
	handler            MessageHandler
	devices            map[string]Device
	syncPayload        []SyncDevices
	executionTemplates map[string]string
}

type MessageHandler interface {
	SendMessage(topic string, message string)
}

func NewFullfillment(handler MessageHandler, deviceConfigs map[string]config.DeviceConfig, executionTemplates map[string]string) (*Fullfillment, error) {
	devices, err := initDevices(deviceConfigs)
	if err != nil {
		return nil, err
	}

	return &Fullfillment{
		handler:            handler,
		devices:            devices,
		syncPayload:        syncPayload(deviceConfigs),
		executionTemplates: executionTemplates,
	}, nil
}

func initDevices(deviceConfigs map[string]config.DeviceConfig) (map[string]Device, error) {
	devices := map[string]Device{}
	for id, config := range deviceConfigs {
		if !strings.Contains(config.Topic, "%s") {
			return nil, fmt.Errorf("no %%s in topic config for device '%s'", id)
		}
		devices[id] = Device{
			Topic: config.Topic,
			State: LocalState{
				State: "off",
				On:    true,
			},
		}
	}
	return devices, nil
}

func (f *Fullfillment) Handler(w http.ResponseWriter, r *http.Request) {
	var request FullfillementRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Error("fullfillment bad request", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := f.handle(request)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	logResponse(response)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error("failed to return response", err)
	}
}

func (f *Fullfillment) handle(request FullfillementRequest) interface{} {
	for _, input := range request.Inputs {
		switch input.Intent {
		case "action.devices.SYNC":
			return f.sync(request, "1234")
		case "action.devices.QUERY":
			return f.query(request.RequestID, input.Payload)
		case "action.devices.EXECUTE":
			return f.execute(request.RequestID, input.Payload)
		case "action.devices.DISCONNECT":
			return f.disconnect(request.RequestID, input.Payload)
		default:
			log.Error("handle intent failed", "input", input)
		}
	}

	log.Error("failed to handle unknown input", "inputs", request.Inputs)
	return EmptyResponse{}
}

func logResponse(response interface{}) {
	str, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Error("failed to return response", err)
	}

	log.Info("response:")
	fmt.Println(string(str))
}

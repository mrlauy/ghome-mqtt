package fullfillment

import (
	"fmt"
	log "log/slog"
	"regexp"
	"strconv"
	"strings"
)

/*
{
  "requestId": "ff36a3cc-ec34-11e6-b1a0-64510650abcf",
  "payload": {
    "commands": [
      {
        "ids": [
          "123"
        ],
        "status": "SUCCESS",
        "states": {
          "on": true,
          "online": true
        }
      },
      {
        "ids": [
          "456"
        ],
        "status": "ERROR",
        "errorCode": "deviceTurnedOff"
      }
    ]
  }
}

SUCCESS: The request succeeded.
OFFLINE: Target device is offline or otherwise unreachable.
EXCEPTIONS: There is an issue or alert associated with the request.
ERROR: The request failed with the corresponding errorCode.

*/

type ExecuteResponse struct {
	RequestID string         `json:"requestId"` // Required. ID of the corresponding request.
	Payload   ExecutePayload `json:"payload"`   // Required. Intent response payload.
}

type ExecutePayload struct {
	Commands    []ExecuteCommands `json:"commands,omitempty"`    // Each object contains one or more devices with response details. N.B. These may not be grouped the same way as in the request. For example, the request might turn 7 lights on, with 3 lights succeeding and 4 failing, thus with two groups in the response.
	ErrorCode   string            `json:"errorCode,omitempty"`   // An error code for the entire transaction for auth failures and developer system unavailability. For individual device errors, use the errorCode within the device object.
	DebugString string            `json:"debugString,omitempty"` // Detailed error which will never be presented to users but may be logged or used during development.
}

type ExecuteCommands struct {
	Ids    []string      `json:"ids,omitempty"`    // Required. List of device IDs corresponding to this status.
	Status ExecuteStatus `json:"status,omitempty"` // Required. Result of the execute operation.

	States    ExecuteStates `json:"states,omitempty"`    // Aligned with per-trait states described in each trait schema reference. These are the states after execution, if available.
	ErrorCode string        `json:"errorCode,omitempty"` // Expanding ERROR state if needed from the preset error codes, which will map to the errors presented to users.
}

type ExecuteStatus string

const (
	Success    ExecuteStatus = "SUCCESS"    // SUCCESS    Confirm that the command succeeded.
	Pending                  = "PENDING"    // PENDING    Command is enqueued but expected to succeed.
	Offline                  = "OFFLINE"    // OFFLINE    Target device is in offline state or unreachable.
	Exceptions               = "EXCEPTIONS" // EXCEPTIONS There is an issue or alert associated with a command. The command could succeed or fail. This status type is typically set when you want to send additional information about another connected device.
	Error                    = "ERROR"      // ERROR      Target device is unable to perform the command.
)

type ExecuteStates struct {
	Online bool `json:"online,omitempty"` // Indicates if the device is online (that is, reachable) or not.

	// action.devices.traits.OnOff
	On bool `json:"on,omitempty"`

	// action.devices.traits.Volume
	CurrentVolume int  `json:"currentVolume,omitempty"`
	IsMuted       bool `json:"isMuted,omitempty"`

	// action.devices.traits.TransportControl
	ActivityState string `json:"activityState,omitempty"` // Supported values: INACTIVE, STANDBY, ACTIVE
	PlaybackState string `json:"playbackState,omitempty"` // Supported values: PAUSED, PLAYING, FAST_FORWARDING, REWINDING, BUFFERING, STOPPED

}

func (f *Fullfillment) execute(requestId string, payload PayloadRequest) ExecuteResponse {
	log.Info("handle execute request", "request", requestId, "payload", payload)

	executeCommands := []ExecuteCommands{}
	for _, command := range payload.Commands {
		for _, device := range command.Devices {
			if deviceState, ok := f.devices[device.ID]; !ok {
				log.Error("failed to find local state", "device", device.ID, "state", deviceState)
				executeCommands = append(executeCommands, ExecuteCommands{
					Ids:    []string{device.ID},
					Status: Error,
					States: ExecuteStates{
						Online: false,
					},
				})
				break
			}

			for _, execution := range command.Execution {
				executeCommand := f.executeCommand(device.ID, execution)
				executeCommands = append(executeCommands, executeCommand)
			}
		}
	}

	log.Info("state after executing command", "devices", f.devices)

	return ExecuteResponse{
		RequestID: requestId,
		Payload: ExecutePayload{
			Commands: executeCommands,
		},
	}
}

func (f *Fullfillment) executeCommand(deviceId string, execution ExecutionRequest) ExecuteCommands {
	device := f.devices[deviceId]
	defer func() {
		f.devices[deviceId] = device
	}()

	switch execution.Command {
	case "action.devices.commands.OnOff":
		action := onOffValue(execution.Params.On)
		message, err := f.fillMessage(deviceId, execution.Command, action)
		if err != nil {
			log.Error("failed to execute command '%s'", execution.Command, err)
			return errorCommand(deviceId)
		}

		f.sentCommand(deviceId, message)
		device.State.On = execution.Params.On
		return ExecuteCommands{
			Ids:    []string{deviceId},
			Status: Success,
			States: ExecuteStates{
				On:     execution.Params.On,
				Online: true,
			},
		}
	case "action.devices.commands.mute":
		message, err := f.fillMessage(deviceId, execution.Command, strconv.FormatBool(execution.Params.Mute))
		if err != nil {
			log.Error("failed to execute command '%s'", execution.Command, err)
			return errorCommand(deviceId)
		}

		f.sentCommand(deviceId, message)
		return ExecuteCommands{
			Ids:    []string{deviceId},
			Status: Success,
			States: ExecuteStates{
				Online:        true,
				CurrentVolume: 10,
				IsMuted:       execution.Params.Mute,
			},
		}
	case "action.devices.commands.setVolume":
		volume := execution.Params.VolumeLevel
		message, err := f.fillMessage(deviceId, execution.Command, strconv.Itoa(volume))
		if err != nil {
			log.Error("failed to execute command '%s'", execution.Command, err)
			return errorCommand(deviceId)
		}

		f.sentCommand(deviceId, message)
		return ExecuteCommands{
			Ids:    []string{deviceId},
			Status: Success,
			States: ExecuteStates{
				Online:        true,
				CurrentVolume: 10,
				IsMuted:       false,
			},
		}
	case "action.devices.commands.volumeRelative":
		action := "decrease"
		if execution.Params.RelativeSteps > 0 {
			action = "increase"
		}
		message, err := f.fillMessage(deviceId, execution.Command, action)
		if err != nil {
			log.Error("failed to execute command '%s'", execution.Command, err)
			return errorCommand(deviceId)
		}

		f.sentCommand(deviceId, message)
		return ExecuteCommands{
			Ids:    []string{deviceId},
			Status: Success,
			States: ExecuteStates{
				Online:        true,
				CurrentVolume: 10 + execution.Params.RelativeSteps,
				IsMuted:       false,
			},
		}
	default:
		log.Info("execute command", "command", execution.Command, "device", deviceId)
		return ExecuteCommands{
			Ids:    []string{deviceId},
			Status: Error,
			States: ExecuteStates{
				Online: false,
			},
		}
	}
}

func (f *Fullfillment) fillMessage(deviceId string, command string, args ...any) (msg string, err error) {
	messageTemplate, commandFound := f.executionTemplates[command]
	if !commandFound {
		return "", fmt.Errorf("failed to find command `%s` for device `%s` in execution template", command, deviceId)

	}

	regex := regexp.MustCompile("%[d|s|v]")
	matches := regex.FindAllStringIndex(messageTemplate, -1)
	if len(matches) != len(args) {
		return "", fmt.Errorf("failed number of arguments %d doesn't, match arguments in template %d in: %s", len(args), strings.Count(messageTemplate, "%s"), messageTemplate)
	}

	message := fmt.Sprintf(messageTemplate, args...)
	return message, nil
}

func (f *Fullfillment) sentCommand(deviceId string, message string) {
	topic := f.devices[deviceId].Topic
	f.handler.SendMessage(fmt.Sprintf(topic, deviceId), message)
}

func errorCommand(deviceId string) ExecuteCommands {
	return ExecuteCommands{
		Ids:       []string{deviceId},
		Status:    Error,
		ErrorCode: "hardError",
	}
}

func onOffValue(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

package fullfillment

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillMessage(t *testing.T) {
	// messageHandlerMock := &MessageHandlerMock{map[string][]string{}}
	fullfillment := &Fullfillment{
		executionTemplates: ExecutionTemplates{
			"device": {
				"command.Test":    `{"argument":"%s"}`,
				"command.TestTwo": `{"first_argument":"%s", "second_argument":"%s"}`,
				"command.TestInt": `{"argument":"%d"}`,
				"command.TestVar": `{"argument":"%v"}`,
			},
			"no-template-device": {},
		},
	}

	tests := []struct {
		name           string
		device         string
		command        string
		args           []interface{}
		expectedResult string
		expectedError  error
	}{
		{
			name:           "Format a string argument",
			device:         "device",
			command:        "command.Test",
			args:           []interface{}{"this"},
			expectedResult: `{"argument":"this"}`,
			expectedError:  nil,
		},
		{
			name:           "Format two strings arguments",
			device:         "device",
			command:        "command.TestTwo",
			args:           []interface{}{"this", "that"},
			expectedResult: `{"first_argument":"this", "second_argument":"that"}`,
			expectedError:  nil,
		},
		{
			name:           "Incorrect number of arguments",
			device:         "device",
			command:        "command.Test",
			args:           []interface{}{"this", "that"},
			expectedResult: "",
			expectedError:  errors.New(`failed number of arguments 2 doesn't, match arguments in template 1 in: {"argument":"%s"}`),
		},
		{
			name:           "Format a Int argument / questionable typesafety",
			device:         "device",
			command:        "command.TestInt",
			args:           []interface{}{42},
			expectedResult: `{"argument":"42"}`,
			expectedError:  nil,
		},
		{
			name:           "Incorrect type of argument / questionable typesafety",
			device:         "device",
			command:        "command.Test",
			args:           []interface{}{42},
			expectedResult: `{"argument":"%!s(int=42)"}`,
			expectedError:  nil,
		},
		{
			name:           "Test var argument",
			device:         "device",
			command:        "command.TestVar",
			args:           []interface{}{42},
			expectedResult: `{"argument":"42"}`,
			expectedError:  nil,
		},
		{
			name:           "Unknown device error",
			device:         "unknown-device",
			command:        "trait",
			args:           []interface{}{},
			expectedResult: "",
			expectedError:  errors.New("failed to find device `unknown-device` in execution template"),
		},
		{
			name:           "Unknown template error",
			device:         "no-template-device",
			command:        "trait",
			args:           []interface{}{},
			expectedResult: "",
			expectedError:  errors.New("failed to find command `trait` for device `no-template-device` in execution template"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := fullfillment.fillMessage(test.device, test.command, test.args...)

			assert.Equal(t, test.expectedResult, result)
			if test.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestExecute(t *testing.T) {
	messageHandlerMock := &MessageHandlerMock{map[string]string{}}
	fullfillment := &Fullfillment{
		state: map[string]LocalState{
			"test-device": {},
		},
		handler: messageHandlerMock,
		executionTemplates: ExecutionTemplates{
			"test-device": {
				"action.devices.commands.volumeRelative": `{"volume":"%s"}`,
			},
		},
	}

	tests := []struct {
		name                string
		requestId           string
		payload             PayloadRequest
		expectedResult      ExecuteResponse
		expectedPublication bool
		expectedTopic       string
		expectedMessage     string
	}{
		// TODO test more cases
		{
			name:      "Test a request",
			requestId: "test-request",
			payload: PayloadRequest{
				AgentUserID: "",
				Commands: []CommandRequest{
					{
						Devices: []DeviceRequest{
							{
								ID: "test-device",
							},
						},
						Execution: []ExecutionRequest{
							{
								Command: "action.devices.commands.volumeRelative",
								Params: ParamsRequest{
									RelativeSteps: -1,
								},
							},
						},
					},
				},
			},
			expectedResult: ExecuteResponse{
				RequestID: "test-request",
				Payload: ExecutePayload{
					Commands: []ExecuteCommands{
						{
							Ids:    []string{"test-device"},
							Status: Success,
							States: ExecuteStates{
								Online:        true,
								CurrentVolume: 9,
							},
						},
					},
				},
			},
			expectedPublication: true,
			expectedTopic:       "device/test-device/set",
			expectedMessage:     `{"volume":"decrease"}`,
		},
		{
			name:      "Test an unknown template",
			requestId: "test-request",
			payload: PayloadRequest{
				AgentUserID: "",
				Commands: []CommandRequest{
					{
						Devices: []DeviceRequest{
							{
								ID: "test-device",
							},
						},
						Execution: []ExecutionRequest{
							{
								Command: "action.devices.commands.mute",
								Params: ParamsRequest{
									Mute: true,
								},
							},
						},
					},
				},
			},
			expectedResult: ExecuteResponse{
				RequestID: "test-request",
				Payload: ExecutePayload{
					Commands: []ExecuteCommands{
						{
							Ids:       []string{"test-device"},
							Status:    Error,
							ErrorCode: "hardError",
						},
					},
				},
			},
			expectedPublication: false,
		},
		{
			name:      "Test empty request",
			requestId: "test-empty",
			payload: PayloadRequest{
				AgentUserID: "",
				Commands:    []CommandRequest{},
			},
			expectedResult: ExecuteResponse{
				RequestID: "test-empty",
				Payload: ExecutePayload{
					Commands:  []ExecuteCommands{},
					ErrorCode: "",
				},
			},
			expectedPublication: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			messageHandlerMock.Reset()

			result := fullfillment.execute(test.requestId, test.payload)

			assert.Equal(t, test.expectedResult, result)
			if test.expectedPublication {
				assert.Contains(t, messageHandlerMock.messages, test.expectedTopic)
				assert.Equal(t, test.expectedMessage, messageHandlerMock.messages[test.expectedTopic])
			} else {
				assert.Empty(t, messageHandlerMock.messages)
			}
		})
	}
}

type MessageHandlerMock struct {
	messages map[string]string
}

func (m *MessageHandlerMock) Reset() {
	m.messages = map[string]string{}
}

func (m *MessageHandlerMock) SendMessage(topic string, message string) {
	m.messages[topic] = message
}

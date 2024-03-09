package auth

import (
	"fmt"
	"github.com/mrlauy/ghome-mqtt/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizeRequest(t *testing.T) {
	// GET https://myservice.example.com/auth?client_id=GOOGLE_CLIENT_ID&redirect_uri=REDIRECT_URI&state=STATE_STRING&scope=REQUESTED_SCOPES&response_type=code&user_locale=LOCALE
	expected := ``
	responseRecorder := httptest.NewRecorder()
	redirectLocation := "/login"

	url := fmt.Sprintf("https://myservice.example.com/oauth/authorize?client_id=CLIENT_ID&redirect_uri=REDIRECT_URI&state=STATE_STRING&scope=REQUESTED_SCOPES&response_type=code&user_locale=LOCALE")
	auth := NewAuth(authConfig())

	request, err := http.NewRequest(http.MethodGet, url, nil)
	assert.Nil(t, err, "failed to create request")

	handler := http.HandlerFunc(auth.Authorize)
	handler.ServeHTTP(responseRecorder, request)

	if status := responseRecorder.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if responseRecorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", responseRecorder.Body.String(), expected)
	}

	if responseRecorder.Header().Get("location") != redirectLocation {
		t.Errorf("redirect location unexpected: got %v want %v", responseRecorder.Header().Get("location"), redirectLocation)
	}
}

func authConfig() config.AuthConfig {
	return config.AuthConfig{
		Client: struct {
			Id     string `yaml:"id" env:"CLIENT_ID" env-default:"000000"`
			Secret string `yaml:"secret" env:"CLIENT_SECRET" env-default:"999999"`
			Domain string `yaml:"domain" env:"CLIENT_SECRET" env-default:"https://oauth-redirect.googleusercontent.com/r/project/project-id"`
		}(struct {
			Id     string
			Secret string
			Domain string
		}{
			Id: "CLIENT_ID", Secret: "client-secret", Domain: "http://localhost",
		}),
		Credientials: ".credentials",
		TokenStore:   ".tokenstore",
	}
}

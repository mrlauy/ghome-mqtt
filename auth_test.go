package main

import (
	"fmt"
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
	auth := NewAuth("CLIENT_ID", "client-secret", "http://localhost")

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

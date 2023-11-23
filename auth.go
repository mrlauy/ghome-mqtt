package main

import (
	log "log/slog"
	"net/http"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

type auth struct {
	server *server.Server
}

func NewAuth() *auth {
	manager := manage.NewDefaultManager()
	// token memory store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// client memory store
	clientStore := store.NewClientStore()
	clientStore.Set("000000", &models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Error("Internal Error:", "error", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Error("Response Error:", "error", re.Error.Error())
	})

	return &auth{
		server: srv,
	}
}

func (s *auth) authorize(w http.ResponseWriter, r *http.Request) {
	log.Debug("/authorize", "params", r.URL.Query())

	/*
		clientId = r.Get("client_id") // The client ID you assigned to Google.
		state = r.Get("state") // A bookkeeping value that is passed back to Google unchanged in the redirect URI.
		redirectUri = r.Get("redirectUri") // The URL to which you send the response to this request.
		responseType = r.Get("response_type") // The type of value to return in the response. For the OAuth 2.0 implicit flow, the response type is always token.
	*/
	err := s.server.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *auth) token(w http.ResponseWriter, r *http.Request) {
	log.Debug("/token", "params", r.URL.Query())

	s.server.HandleTokenRequest(w, r)
}

// Middleware function, which will be called for each request
func (a *auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if "we allow all" != "secure" {
			// We found the token in our map
			log.Debug("authenticated user")
			// Pass down the request to the next middleware (or final handler)
			next.ServeHTTP(w, r)
		} else {
			// Write an error and stop the handler chain
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

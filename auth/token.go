package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-session/session"
	log "log/slog"
	"net/http"
	"time"
)

func (a *Auth) Authorize(w http.ResponseWriter, r *http.Request) {
	log.Debug("handle oauth/authorize")

	if r.Method == http.MethodPost && r.Form == nil {
		if err := r.ParseForm(); err != nil {
			log.Error("error parsing authorize form", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	sessionStore, err := session.Start(r.Context(), w, r)
	if err != nil {
		log.Error("failed to get session", err)
		http.Error(w, "server side error: #3100", http.StatusInternalServerError)
		return
	}

	if uid, ok := sessionStore.Get("LoggedInUserID"); ok {
		userId := uid.(string)

		state, stateFound := sessionStore.Get("state")
		if !stateFound {
			log.Error("failed to find state in session", sessionStore)
			http.Error(w, "server side error: #3201", http.StatusInternalServerError)
			return
		}
		redirectUri, redirectUriFound := sessionStore.Get("redirectUri")
		if !redirectUriFound {
			log.Error("failed to find redirectUri in session", sessionStore)
			http.Error(w, "server side error: #3202", http.StatusInternalServerError)
			return
		}

		authorizationCode, err := a.generateCode(r.Context(), "code", userId, redirectUri.(string), "", r)
		if err != nil {
			log.Error("failed to create code", err)
			http.Error(w, "server side error: #3203", http.StatusInternalServerError)
			return
		}

		log.Info("/authorize", "user", userId)
		log.Debug("/authorize generate code", "code", authorizationCode)

		client, clientFound := sessionStore.Get("client")
		if !clientFound {
			log.Error("failed to find client in session", "session", sessionStore)
			http.Error(w, "server side error: #3204", http.StatusInternalServerError)
			return
		}

		a.codes[client.(string)] = authorizationCode
		sessionStore.Set("authorizationCode", authorizationCode)
		sessionStore.Delete("client")
		sessionStore.Delete("scope")
		sessionStore.Delete("state")
		sessionStore.Delete("redirectUri")
		err = sessionStore.Save()
		if err != nil {
			log.Error("failed to store session", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		responseUrl := fmt.Sprintf("%s?code=%s&state=%s", redirectUri, authorizationCode, state)
		log.Info("/authorize redirect", "location", responseUrl)

		w.Header().Set("Location", responseUrl)
		w.WriteHeader(http.StatusFound)
		return
	}

	responseType := getInput(r, "response_type")
	client := getInput(r, "client_id")
	redirectUri := getInput(r, "redirect_uri")
	scope := getInput(r, "scope")
	state := getInput(r, "state")

	if responseType != "code" {
		log.Error("response type is not supported", "responseType", responseType)
		http.Error(w, "response type is not supported", http.StatusBadRequest)
		return
	}

	if client != a.clientId {
		log.Error("invalid client", "client", client)
		http.Error(w, "invalid_client", http.StatusInternalServerError)
		return
	}

	// TODO deal with scope?

	sessionStore.Set("client", client)
	sessionStore.Set("state", state)
	sessionStore.Set("redirectUri", redirectUri)
	sessionStore.Set("scope", scope)
	err = sessionStore.Save()
	if err != nil {
		log.Error("failed to store session", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	log.Info("/authorize", "client", client, "responseType", responseType, "scope", scope, "redirectUri", redirectUri)

	w.Header().Set("Location", "/login")
	w.WriteHeader(http.StatusFound)
}

func (a *Auth) Token(w http.ResponseWriter, r *http.Request) {
	log.Debug("handle oauth/token")

	if r.Method == http.MethodPost && r.Form == nil {
		if err := r.ParseForm(); err != nil {
			log.Error("error parsing token form:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	grantType := getInput(r, "grant_type")
	if grantType != "authorization_code" && grantType != "refresh_token" {
		log.Error("unsupported grant type", "grant type", grantType)
		http.Error(w, "unsupported grant type", http.StatusInternalServerError)
		return
	}

	client := getInput(r, "client_id")
	secret := getInput(r, "client_secret")
	if client != a.clientId || secret != a.clientSecret {
		log.Error("invalid client", "client", client, "secret", secret)
		http.Error(w, "invalid_client", http.StatusInternalServerError)
		return
	}

	var response map[string]interface{}
	var err error
	if grantType == "authorization_code" {
		code := getInput(r, "code")

		authorizationCode, ok := a.codes[client]
		if !ok {
			log.Error("failed to find authorizationCode in session", "client", client, "codes", a.codes)
			http.Error(w, "server side error: #3301", http.StatusInternalServerError)
			return
		}

		if code != authorizationCode {
			log.Error("invalid code", "client", client, "code", code)
			http.Error(w, "invalid code", http.StatusInternalServerError)
			return
		}

		// todo check it again
		redirectUri := getInput(r, "redirect_uri")
		scope := getInput(r, "scope")

		response, err = a.generateAuthorizationToken(client, secret, redirectUri, scope, authorizationCode, grantType, r)
		if err != nil {
			log.Error("failed to generate code", "client", client, err)
			http.Error(w, "failed to generate code", http.StatusInternalServerError)
		}
	} else if grantType == "refresh_token" {
		refreshToken := getInput(r, "refresh_token")
		scope := getInput(r, "scope")

		response, err = a.generateRefreshToken(client, secret, refreshToken, scope, grantType, r)
		if err != nil {
			log.Error("failed to generate code", "client", client, err)
			http.Error(w, "failed to generate code", http.StatusInternalServerError)
			return
		}
	}

	log.Debug("/token grant token", "client", client)
	log.Info("/token grant", "grantType", grantType, "response", response)

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error("failed to generate response", "client", client, err)
		http.Error(w, "failed to generate response", http.StatusInternalServerError)
		return
	}
}

func (a *Auth) generateAuthorizationToken(client, secret string, redirectUri string, scope string, authorizationCode string, grantType string, r *http.Request) (map[string]interface{}, error) {
	tokenGenerateRequest := &oauth2.TokenGenerateRequest{
		ClientID:     client,
		ClientSecret: secret,
		RedirectURI:  redirectUri,
		Scope:        scope,
		Code:         authorizationCode,
		Request:      r,
	}
	tokenInfo, err := a.manager.GenerateAccessToken(r.Context(), oauth2.GrantType(grantType), tokenGenerateRequest)
	if err != nil {
		return nil, err
	}

	log.Info("token info", "info", tokenInfo)

	response := map[string]interface{}{
		"token_type":   "bearer",
		"access_token": tokenInfo.GetAccess(),
		"expires_in":   int64(tokenInfo.GetAccessExpiresIn() / time.Second),
	}

	if scope := tokenInfo.GetScope(); scope != "" {
		response["scope"] = scope
	}

	if refresh := tokenInfo.GetRefresh(); refresh != "" {
		response["refresh_token"] = refresh
	}
	return response, nil
}

func (a *Auth) generateRefreshToken(client, secret string, refreshToken string, scope string, grantType string, r *http.Request) (map[string]interface{}, error) {
	tokenGenerateRequest := &oauth2.TokenGenerateRequest{
		ClientID:       client,
		ClientSecret:   secret,
		Refresh:        refreshToken,
		AccessTokenExp: 24 * time.Hour,
		Scope:          scope,
		Request:        r,
	}

	tokenInfo, err := a.manager.GenerateAccessToken(r.Context(), oauth2.GrantType(grantType), tokenGenerateRequest)
	if err != nil {
		return nil, err
	}

	log.Info("token info", "info", tokenInfo)

	response := map[string]interface{}{
		"token_type":   "bearer",
		"access_token": tokenInfo.GetAccess(),
		"expires_in":   int64(tokenInfo.GetAccessExpiresIn() / time.Second),
	}

	if scope := tokenInfo.GetScope(); scope != "" {
		response["scope"] = scope
	}

	if refresh := tokenInfo.GetRefresh(); refresh != "" {
		response["refresh_token"] = refresh
	}
	return response, nil
}

func (a *Auth) generateCode(ctx context.Context, responseType string, userId string, redirectUri string, scope string, r *http.Request) (string, error) {
	tokenRequest := &oauth2.TokenGenerateRequest{
		ClientID:       a.clientId,
		UserID:         userId,
		RedirectURI:    redirectUri,
		Scope:          scope,
		AccessTokenExp: 2 * time.Minute,
		Request:        r,
	}

	tokenInfo, err := a.manager.GenerateAuthToken(ctx, oauth2.ResponseType(responseType), tokenRequest)
	if err != nil {
		return "", err
	}
	return tokenInfo.GetCode(), nil
}

func getInput(r *http.Request, key string) string {
	if r.URL.Query().Has(key) {
		return r.URL.Query().Get(key)
	}
	return r.FormValue(key)
}

package main

import (
	"bufio"
	"context"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-session/session"

	"encoding/json"
	"fmt"
	"html/template"
	log "log/slog"
)

type auth struct {
	clientId     string
	clientSecret string
	credentials  map[string]string
	manager      *manage.Manager
	codes        map[string]string
}

func NewAuth(cfg AuthConfig) *auth {
	clientStore := store.NewClientStore()
	clientStore.Set(cfg.Client.Id, &models.Client{
		ID:     cfg.Client.Id,
		Secret: cfg.Client.Secret,
		Domain: cfg.Client.Domain,
	})

	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.MapClientStorage(clientStore)
	manager.MustTokenStorage(store.NewFileTokenStore(cfg.TokenStore))
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte("00000000"), jwt.SigningMethodHS512))

	credentials, err := loadCredentials(cfg.Credientials)
	if err != nil || len(credentials) < 1 {
		log.Warn("no user credentials, add users to .credentials", err)
	}

	return &auth{
		clientId:     cfg.Client.Id,
		clientSecret: cfg.Client.Secret,
		credentials:  credentials,
		manager:      manager,
		codes:        map[string]string{},
	}
}

func (a *auth) Login(page *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionStore, err := session.Start(r.Context(), w, r)
		if err != nil {
			log.Error("failed to get login session", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Method == http.MethodPost {
			username := strings.Trim(strings.ToLower(r.FormValue("username")), " ")
			password := strings.Trim(r.FormValue("password"), " ")

			storedPassword, ok := a.credentials[username]
			if !ok {
				log.Warn("user unknown", "user", username)
				http.Error(w, "wrong credentials", http.StatusUnauthorized)
				return
			}

			if !checkPasswordHash(password, storedPassword) {
				// TODO return error in page
				log.Warn("wrong credentials", "user", username)
				http.Error(w, "wrong credentials", http.StatusUnauthorized)
				return
			}

			log.Info("/login", "username", username)

			sessionStore.Set("LoggedInUserID", username)
			err = sessionStore.Save()
			if err != nil {
				log.Error("failed to store session", err)
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Location", "/confirm")
			w.WriteHeader(http.StatusFound)
			return
		}

		log.Info("/login")
		err = page.Execute(w, "data")
		if err != nil {
			log.Error("failed to render login page", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
	}
}

func (a *auth) Confirm(page *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionStore, err := session.Start(r.Context(), w, r)
		if err != nil {
			log.Error("error starting session in confirm handler", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userId, ok := sessionStore.Get("LoggedInUserID")
		if !ok {
			log.Warn("/confirm failed to get user in session")
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusFound)
			return
		}

		log.Info("/confirm", "user", userId)
		err = page.Execute(w, "data")
		if err != nil {
			log.Error("failed to render auth page", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
	}
}

func (a *auth) Authorize(w http.ResponseWriter, r *http.Request) {
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

func (a *auth) Token(w http.ResponseWriter, r *http.Request) {
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
			log.Error("failed to generate code", "client", client, "tokenGenerateRequest", tokenGenerateRequest, err)
			http.Error(w, "failed to generate code", http.StatusInternalServerError)
			return
		}

		log.Info("token info", "info", tokenInfo)

		response = map[string]interface{}{
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
	} else if grantType == "refresh_token" {
		refreshToken := getInput(r, "refresh_token")

		scope := getInput(r, "scope")
		tokenGenerateRequest := &oauth2.TokenGenerateRequest{
			ClientID:       client,
			ClientSecret:   secret,
			Refresh:        refreshToken,
			AccessTokenExp: 2 * time.Minute,
			Scope:          scope,
			Request:        r,
		}

		tokenInfo, err := a.manager.GenerateAccessToken(r.Context(), oauth2.GrantType(grantType), tokenGenerateRequest)
		if err != nil {
			log.Error("failed to generate code", "client", client, "tokenGenerateRequest", tokenGenerateRequest, err)
			http.Error(w, "failed to generate code", http.StatusInternalServerError)
			return
		}

		log.Info("token info", "info", tokenInfo)

		response = map[string]interface{}{
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
	}

	log.Debug("/token grant token", "client", client)
	log.Info("/token grant", "grantType", grantType, "response", response)

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error("failed to generate response", "client", client, err)
		http.Error(w, "failed to generate response", http.StatusInternalServerError)
		return
	}
}

func (a *auth) ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := getBearerToken(r)
		if !ok {
			log.Warn("unauthorized: no token")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenInfo, err := a.manager.LoadAccessToken(r.Context(), token)
		if err != nil {
			log.Warn("unauthorized", "token", token, err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		log.Info("grant access", "URL", r.URL, "token", tokenInfo)

		next.ServeHTTP(w, r)
	})
}

func (a *auth) generateCode(ctx context.Context, responseType string, userId string, redirectUri string, scope string, r *http.Request) (string, error) {
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

func loadCredentials(filename string) (map[string]string, error) {
	credentials := map[string]string{}
	file, err := os.Open(filename)
	if err != nil {
		return credentials, fmt.Errorf("failed to load .credentials")
	}

	fscanner := bufio.NewScanner(file)
	for fscanner.Scan() {
		line := fscanner.Text()
		lastInd := strings.LastIndex(line, ":")
		username := line[:lastInd]
		password := line[lastInd+1:]

		credentials[username] = password
	}
	return credentials, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// BearerAuth parse bearer token
func getBearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	prefix := "Bearer "
	token := ""

	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	} else {
		token = r.FormValue("access_token")
	}

	return token, token != ""
}

func getInput(r *http.Request, key string) string {
	if r.URL.Query().Has(key) {
		return r.URL.Query().Get(key)
	}
	return r.FormValue(key)
}

func dump(vars map[string]string) interface{} {
	bytes, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		log.Error("failed to marchal dump", "vars", vars, err)
	}
	return string(bytes)
}

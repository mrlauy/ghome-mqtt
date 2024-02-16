package auth

import (
	"bufio"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/golang-jwt/jwt"
	"github.com/mrlauy/ghome-mqtt/config"
	"os"
	"strings"

	"fmt"
	log "log/slog"
)

type Auth struct {
	clientId     string
	clientSecret string
	credentials  map[string]string
	manager      *manage.Manager
	codes        map[string]string
}

func NewAuth(cfg config.AuthConfig) *Auth {
	clientStore := store.NewClientStore()
	err := clientStore.Set(cfg.Client.Id, &models.Client{
		ID:     cfg.Client.Id,
		Secret: cfg.Client.Secret,
		Domain: cfg.Client.Domain,
	})
	if err != nil {
		log.Error("failed to load client config in client store", err)
		return nil
	}

	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.MapClientStorage(clientStore)
	manager.MustTokenStorage(store.NewFileTokenStore(cfg.TokenStore))
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte("00000000"), jwt.SigningMethodHS512))

	credentials, err := loadCredentials(cfg.Credientials)
	if err != nil || len(credentials) < 1 {
		log.Warn("no user credentials, add users to .credentials", err)
	}

	return &Auth{
		clientId:     cfg.Client.Id,
		clientSecret: cfg.Client.Secret,
		credentials:  credentials,
		manager:      manager,
		codes:        map[string]string{},
	}
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

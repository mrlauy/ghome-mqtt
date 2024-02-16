package auth

import (
	log "log/slog"
	"net/http"
	"strings"
)

func (a *Auth) ValidateToken(next http.Handler) http.Handler {
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

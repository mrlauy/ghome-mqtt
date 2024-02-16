package auth

import (
	"github.com/go-session/session"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	log "log/slog"
	"net/http"
	"strings"
)

func (a *Auth) Login(page *template.Template) http.HandlerFunc {
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

func (a *Auth) Confirm(page *template.Template) http.HandlerFunc {
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

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

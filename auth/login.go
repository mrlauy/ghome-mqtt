package auth

import (
	"github.com/go-session/session"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	log "log/slog"
	"net/http"
	"strings"
)

type PageData struct {
	Error string
}

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
				responseError(w, page, "wrong credentials", http.StatusUnauthorized)
				return
			}

			if !checkPasswordHash(password, storedPassword) {
				log.Warn("wrong credentials", "user", username)
				responseError(w, page, "wrong credentials", http.StatusUnauthorized)
				return
			}

			log.Info("/login", "username", username)

			sessionStore.Set("LoggedInUserID", username)
			err = sessionStore.Save()
			if err != nil {
				log.Error("failed to store session", err)
				responseError(w, page, "server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Location", "/confirm")
			w.WriteHeader(http.StatusFound)
			return
		}

		log.Info("/login")
		err = page.Execute(w, nil)
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
			responseError(w, page, "server error", http.StatusInternalServerError)
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
		err = page.Execute(w, nil)
		if err != nil {
			log.Error("failed to render login page", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
	}
}

func responseError(w http.ResponseWriter, page *template.Template, message string, code int) {
	err := page.Execute(w, PageData{Error: message})
	if err != nil {
		log.Error("failed to render login page", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

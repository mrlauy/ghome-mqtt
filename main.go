package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"

	"github.com/gorilla/mux"
	"github.com/mrlauy/ghome-mqtt/fullfillment"
)

const requestFullDump = false

var loginPage *template.Template
var authPage *template.Template

func main() {
	cfg, err := ReadConfig()
	if err != nil {
		log.Fatal("failed to read config: ", err)
	}

	auth := NewAuth("sampleClientId", "sampleClientSecret", "http://localhost")
	mqtt, err := NewMqtt(cfg.Mqtt)
	if err != nil {
		log.Fatal("failed to start mqtt: ", err)
	}

	fullfillmentManager, err := fullfillment.NewFullfillment(cfg.Devices.File, mqtt, cfg.ExecutionTemplates)
	if err != nil {
		log.Fatal("failed to start fullfillment handler: ", err)
	}

	loginPage = template.Must(template.ParseFiles("templates/login.html"))
	authPage = template.Must(template.ParseFiles("templates/auth.html"))

	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/login", auth.Login(loginPage))
	router.HandleFunc("/confirm", auth.Confirm(authPage))
	router.HandleFunc("/oauth/authorize", auth.Authorize)
	router.HandleFunc("/oauth/token", auth.Token)

	smarthomeRouter := router.PathPrefix("/smarthome").Subrouter()
	smarthomeRouter.Use(auth.ValidateToken)
	smarthomeRouter.HandleFunc("/fulfillment", fullfillmentManager.Handler).Methods("POST")

	http.Handle("/", router)

	port := cfg.Server.Port
	log.Printf("start server on port: %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestFullDump {
			headers := r.Header

			// only print the following headers
			r.Header = map[string][]string{}
			r.Header.Add("Cookie", headers.Get("Cookie"))
			r.Header.Add("Referer", headers.Get("Referer"))
			r.Header.Add("Authorization", headers.Get("Authorization"))

			data, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Println("error dumping request:", err)
				return
			}
			r.Header = headers

			log.Printf("\n> %s \n%v", r.URL, string(data))

			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)

			dump, err := httputil.DumpResponse(recorder.Result(), true)
			if err != nil {
				log.Println("error dumping response:", err)
				return
			}
			log.Printf("\n< %s \n %v\n", r.URL, string(dump))

			// we copy the captured response headers to our new response
			for k, v := range recorder.Header() {
				w.Header()[k] = v
			}

			// grab the captured response body
			response := recorder.Body.Bytes()

			w.WriteHeader(recorder.Code)
			w.Write(response)
		} else {
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		}
	})
}

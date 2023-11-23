package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	auth := NewAuth()

	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	router.HandleFunc("/", HomeHandler)
	router.HandleFunc("/authorize", auth.authorize)
	router.HandleFunc("/token", auth.token)

	router.HandleFunc("/fulfillment", FullfillmentHandler).Methods("POST")
	router.HandleFunc("/smarthome/update", UpdateHandler).Methods("POST")
	router.HandleFunc("/smarthome/create", UpdateHandler).Methods("POST")
	router.HandleFunc("/smarthome/delete", UpdateHandler).Methods("POST")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":9096", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>hello<h1>")
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>hello<h1>")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

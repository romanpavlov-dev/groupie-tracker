package main

import (
	"groupie-tracker/backend/api"
	"groupie-tracker/backend/logics"
	"log"
	"net/http"
)

func main() {
	store, err := api.LoadStore()

	if err != nil {
		log.Fatal(err)
	}

	h := &logics.HandlerStore{Store: store}

	http.HandleFunc("/", h.MainHandler)
	http.HandleFunc("/artist", h.ArtistHandler)

	log.Fatal(http.ListenAndServe(":1010", nil))

}

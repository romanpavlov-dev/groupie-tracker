package logics

import (
	"groupie-tracker/backend/api"
	"groupie-tracker/backend/models"
	"log"
	"net/http"
	"text/template"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	var index models.Index

	if err := api.FetchJSON(baseURL, &index); err != nil {
		log.Println(err)
		http.Error(w, "Ошибка получения данных API", http.StatusBadRequest)
		// return fmt.Errorf("error occured while fetching json from index: %e", err)
		return

	}

	var artists []models.Artist

	if err := api.FetchJSON(index.Artists, &artists); err != nil {
		log.Println(err)
		http.Error(w, "Ошибка получения данных API", http.StatusInternalServerError)
		// return fmt.Errorf("error occured while fetching json from artists: %e", err)
		return
	}

	tmp, err := template.ParseFiles("frontend/templates/artists.html")

	if err != nil {
		http.Error(w, "Ошибка парсинга", http.StatusInternalServerError)
		// return fmt.Errorf("error occured while parsing template: %e", err)
		return
	}

	if err := tmp.Execute(w, artists); err != nil {
		log.Println(err)
		// return fmt.Errorf("error occured while executing information: %e", err)
		return
	}

}

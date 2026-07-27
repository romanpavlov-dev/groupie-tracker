package logics

import (
	"groupie-tracker/backend/api"
	"log"
	"net/http"
	"text/template"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

func MainHandler(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseFiles("frontend/templates/artists.html")
	if err != nil {
		http.Error(w, "Ошибка парсинга", http.StatusInternalServerError)
		return
	}

	if err := tmp.Execute(w, api.Views); err != nil {
		log.Println(err)

		return
	}

}

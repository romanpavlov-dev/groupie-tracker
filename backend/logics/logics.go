package logics

import (
	"groupie-tracker/backend/api"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

type HandlerStore struct {
	Store *api.Store
}

func (h *HandlerStore) MainHandler(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseFiles("frontend/templates/mainpage.html")
	if err != nil {
		http.Error(w, "Ошибка парсинга", http.StatusInternalServerError)
		return
	}

	if err := tmp.Execute(w, h.Store.All()); err != nil {
		log.Println(err)

		return
	}

}

func (h *HandlerStore) ArtistHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		return
	}

	view, ok := h.Store.ByID(id)
	if !ok {
		http.Error(w, "no mathces", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	tmp, err := template.ParseFiles("frontend/templates/artist.html")

	if err != nil {
		http.Error(w, "xz", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if err := tmp.Execute(w, view); err != nil {
		log.Println(err)

		return
	}

}

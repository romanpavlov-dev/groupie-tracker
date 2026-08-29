package handlers

import (
	"encoding/json"
	"groupie-tracker/backend/api"
	"groupie-tracker/backend/filters"
	"groupie-tracker/backend/search"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

type HandlerStore struct {
	Store      *api.Store
	StartupErr error
}

func (h *HandlerStore) requireGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return false
	}

	return true
}

func (h *HandlerStore) requireStore(w http.ResponseWriter) bool {
	if h.StartupErr != nil || h.Store == nil {
		http.Error(w, "Failed to load application data", http.StatusInternalServerError)
		return false
	}

	return true
}

func (h *HandlerStore) MainHandler(w http.ResponseWriter, r *http.Request) {

	if !h.requireGET(w, r) {
		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if !h.requireStore(w) {
		return
	}

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

	if !h.requireGET(w, r) {
		return
	}

	if !h.requireStore(w) {
		return
	}

	idStr := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 || idStr == "" {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	view, ok := h.Store.ByID(id)
	if !ok {
		http.Error(w, "Artist not found", http.StatusNotFound)
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

func (h *HandlerStore) HandleFilter(w http.ResponseWriter, r *http.Request) {

	if !h.requireGET(w, r) {
		return
	}

	if !h.requireStore(w) {
		return
	}

	criteria, errs := filters.ParseFilterCriteria(r)

	if len(errs) > 0 {
		w.Header().Set("Content-type", "application/json") //выставляем заголовки
		w.WriteHeader(http.StatusBadRequest)               //выставляем статус код ответа
		json.NewEncoder(w).Encode(map[string]interface{}{  //пишем тело json c ошибками если существуют
			"errors": errs,
		})
		return
	}

	q := r.URL.Query()
	result := filters.ApplyFilters(h.Store.All(), criteria)
	result = search.Apply(result, q.Get("q"), q.Get("match_type"), q.Get("match_value"))

	w.Header().Set("Content-type", "application/json")        //выставляем заголовки
	if err := json.NewEncoder(w).Encode(result); err != nil { //отправляем тело ответа
		log.Println(err)
		http.Error(w, "failed to encode result", http.StatusInternalServerError)
		return
	}

}

func (h *HandlerStore) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if !h.requireGET(w, r) {
		return
	}

	if !h.requireStore(w) {
		return
	}

	suggestions := search.Suggestions(h.Store.All(), r.URL.Query().Get("q"))

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(suggestions); err != nil {
		log.Println(err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *HandlerStore) HandleFilterMeta(w http.ResponseWriter, r *http.Request) {

	if !h.requireGET(w, r) {
		return
	}

	if !h.requireStore(w) {
		return
	}

	meta := h.Store.Meta()

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		log.Println(err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (h *HandlerStore) ArtistLocationsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.requireGET(w, r) {
		return
	}

	if !h.requireStore(w) {
		return
	}

	idStr := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 || idStr == "" {
		http.Error(w, "invalid id", http.StatusBadRequest)
		log.Println(err)
		return
	}

	view, ok := h.Store.ByID(id)
	if !ok {
		http.Error(w, "no matches", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(view.Markers); err != nil {
		log.Println(err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

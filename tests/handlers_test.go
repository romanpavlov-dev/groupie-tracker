package tests

import (
	"errors"
	"groupie-tracker/backend/api"
	"groupie-tracker/backend/geo"
	"groupie-tracker/backend/handlers"
	"groupie-tracker/backend/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testHandler() *handlers.HandlerStore {
	return &handlers.HandlerStore{Store: api.NewStore([]models.ArtistView{
		{Artist: models.Artist{ID: 1, Name: "Test Artist"}},
	})}
}

func TestMainHandlerRejectsUnknownPathAndMethods(t *testing.T) {
	h := testHandler()

	tests := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{name: "unknown path", method: http.MethodGet, path: "/anything", status: http.StatusNotFound},
		{name: "wrong method", method: http.MethodPost, path: "/", status: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()

			h.MainHandler(recorder, req)

			if recorder.Code != tt.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.status)
			}
		})
	}
}

func TestArtistHandlerValidatesIDAndNotFound(t *testing.T) {
	h := testHandler()

	tests := []struct {
		path   string
		status int
	}{
		{path: "/artist", status: http.StatusBadRequest},
		{path: "/artist?id=abc", status: http.StatusBadRequest},
		{path: "/artist?id=999", status: http.StatusNotFound},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		recorder := httptest.NewRecorder()

		h.ArtistHandler(recorder, req)

		if recorder.Code != tt.status {
			t.Errorf("%s: status = %d, want %d", tt.path, recorder.Code, tt.status)
		}
	}
}

func TestHandlersReturnInternalServerErrorWhenStoreLoadFails(t *testing.T) {
	h := &handlers.HandlerStore{StartupErr: errors.New("API unavailable")}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	h.MainHandler(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestCleanLocationUsesCityCountryOrder(t *testing.T) {
	if got := geo.CleanLocation("new_york-usa"); got != "New York, Usa" {
		t.Fatalf("CleanLocation() = %q, want %q", got, "New York, Usa")
	}
}

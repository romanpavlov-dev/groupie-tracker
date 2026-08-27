package tests

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"groupie-tracker/backend/filters"
	"groupie-tracker/backend/models"
)

func TestParseFilterCriteriaParsesQueryValues(t *testing.T) {
	request, err := http.NewRequest("GET", "/?creation_min=2000&creation_max=2020&album_min=01-01-2005&album_max=31-12-2015&members_min=2&members_max=5&locations=Paris,London", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	criteria, errs := filters.ParseFilterCriteria(request)
	if len(errs) != 0 {
		t.Fatalf("ParseFilterCriteria() returned unexpected validation errors: %#v", errs)
	}

	if !criteria.HasCreationMin || criteria.CreationMin != 2000 {
		t.Fatalf("CreationMin parse failed: %+v", criteria)
	}
	if !criteria.HasCreationMax || criteria.CreationMax != 2020 {
		t.Fatalf("CreationMax parse failed: %+v", criteria)
	}
	if !criteria.HasAlbumMin || !criteria.AlbumMin.Equal(time.Date(2005, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("AlbumMin parse failed: %+v", criteria)
	}
	if !criteria.HasAlbumMax || !criteria.AlbumMax.Equal(time.Date(2015, time.December, 31, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("AlbumMax parse failed: %+v", criteria)
	}
	if !criteria.HasMembersMin || criteria.MembersMin != 2 {
		t.Fatalf("MembersMin parse failed: %+v", criteria)
	}
	if !criteria.HasMembersMax || criteria.MembersMax != 5 {
		t.Fatalf("MembersMax parse failed: %+v", criteria)
	}
	if !reflect.DeepEqual(criteria.Locations, []string{"Paris", "London"}) {
		t.Fatalf("Locations parse failed: got %#v", criteria.Locations)
	}
}

func TestApplyFiltersFiltersArtistViews(t *testing.T) {
	views := []models.ArtistView{
		{
			Artist: models.Artist{
				ID:           1,
				Name:         "Alpha",
				CreationDate: 2000,
				FirstAlbum:   "15-03-2005",
				Members:      []string{"A", "B", "C"},
			},
			Locations: []string{"Paris", "Berlin"},
		},
		{
			Artist: models.Artist{
				ID:           2,
				Name:         "Beta",
				CreationDate: 1998,
				FirstAlbum:   "02-06-2001",
				Members:      []string{"A", "B"},
			},
			Locations: []string{"Rome"},
		},
		{
			Artist: models.Artist{
				ID:           3,
				Name:         "Gamma",
				CreationDate: 2010,
				FirstAlbum:   "20-11-2012",
				Members:      []string{"A", "B", "C", "D"},
			},
			Locations: []string{"London"},
		},
	}

	criteria := models.FilterCriteria{
		CreationMin:    2000,
		HasCreationMin: true,
		CreationMax:    2010,
		HasCreationMax: true,
		AlbumMin:       time.Date(2004, time.January, 1, 0, 0, 0, 0, time.UTC),
		HasAlbumMin:    true,
		Locations:      []string{"Paris"},
	}

	filtered := filters.ApplyFilters(views, criteria)
	if len(filtered) != 1 {
		t.Fatalf("ApplyFilters() returned %d artists, want 1", len(filtered))
	}
	if filtered[0].Name != "Alpha" {
		t.Fatalf("ApplyFilters() returned %q, want %q", filtered[0].Name, "Alpha")
	}
}

func TestParseAlbumDateParsesCustomFormat(t *testing.T) {
	parsed, err := filters.ParseAlbumDate("12-05-2020")
	if err != nil {
		t.Fatalf("ParseAlbumDate() returned unexpected error: %v", err)
	}

	want := time.Date(2020, time.May, 12, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(want) {
		t.Fatalf("ParseAlbumDate() = %v, want %v", parsed, want)
	}
}

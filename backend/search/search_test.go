package search

import (
	"groupie-tracker/backend/models"
	"reflect"
	"testing"
)

func sampleArtists() []models.ArtistView {
	return []models.ArtistView{
		{
			Artist: models.Artist{
				ID:           1,
				Name:         "Queen",
				Members:      []string{"Freddie Mercury", "Brian May", "John Deacon", "Roger Taylor"},
				CreationDate: 1970,
				FirstAlbum:   "13-07-1973",
			},
			Locations: []string{"london-uk", "paris-france"},
		},
		{
			Artist: models.Artist{
				ID:           2,
				Name:         "Phil Collins",
				Members:      []string{"Phil Collins"},
				CreationDate: 1981,
				FirstAlbum:   "09-01-1981",
			},
			Locations: []string{"london-uk"},
		},
		{
			Artist: models.Artist{
				ID:           3,
				Name:         "The Beatles",
				Members:      []string{"John Lennon", "Paul McCartney", "George Harrison", "Ringo Starr"},
				CreationDate: 1960,
				FirstAlbum:   "22-03-1963",
			},
			Locations: []string{"liverpool-uk", "new_york-usa"},
		},
	}
}

func suggestionTypes(suggestions []Suggestion, text string) []string {
	var types []string
	for _, s := range suggestions {
		if s.Text == text {
			types = append(types, s.Type)
		}
	}
	return types
}

func findSuggestion(suggestions []Suggestion, text, typ string) (Suggestion, bool) {
	for _, s := range suggestions {
		if s.Text == text && s.Type == typ {
			return s, true
		}
	}
	return Suggestion{}, false
}

func TestSuggestionsEmptyQueryReturnsNone(t *testing.T) {
	got := Suggestions(sampleArtists(), "   ")
	if len(got) != 0 {
		t.Fatalf("empty query should return no suggestions, got %d", len(got))
	}
}

func TestSuggestionsAreCaseInsensitive(t *testing.T) {
	got := Suggestions(sampleArtists(), "FREDDIE")
	if _, ok := findSuggestion(got, "Freddie Mercury", TypeMember); !ok {
		t.Fatalf("expected Freddie Mercury as member, got %#v", got)
	}
}

func TestSuggestionsIdentifyArtistAndMemberSeparately(t *testing.T) {
	got := Suggestions(sampleArtists(), "phil")
	types := suggestionTypes(got, "Phil Collins")

	want := []string{TypeArtistBand, TypeMember}
	if !reflect.DeepEqual(types, want) {
		t.Fatalf("Phil Collins types = %v, want %v", types, want)
	}
}

func TestSuggestionsCoverAllSearchCases(t *testing.T) {
	views := sampleArtists()

	tests := []struct {
		query string
		text  string
		typ   string
	}{
		{"queen", "Queen", TypeArtistBand},
		{"brian", "Brian May", TypeMember},
		{"paris", "paris-france", TypeLocation},
		{"1973", "13-07-1973", TypeFirstAlbumDate},
		{"1960", "1960", TypeCreationDate},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			got := Suggestions(views, tt.query)
			if _, ok := findSuggestion(got, tt.text, tt.typ); !ok {
				t.Fatalf("query %q: missing %#v / %#v in %#v", tt.query, tt.text, tt.typ, got)
			}
		})
	}
}

func TestSuggestionsGroupDuplicateLocations(t *testing.T) {
	got := Suggestions(sampleArtists(), "london")
	s, ok := findSuggestion(got, "london-uk", TypeLocation)
	if !ok {
		t.Fatal("expected london-uk location suggestion")
	}
	if !reflect.DeepEqual(s.ArtistIDs, []int{1, 2}) {
		t.Fatalf("artist ids = %v, want [1 2]", s.ArtistIDs)
	}
}

func TestSuggestionsMatchNormalizedLocation(t *testing.T) {
	got := Suggestions(sampleArtists(), "new york")
	if _, ok := findSuggestion(got, "new_york-usa", TypeLocation); !ok {
		t.Fatalf("expected new_york-usa for query 'new york', got %#v", got)
	}
}

func TestApplyFiltersArtistsByQuery(t *testing.T) {
	got := Apply(sampleArtists(), "beatles", "", "")
	if len(got) != 1 || got[0].Name != "The Beatles" {
		t.Fatalf("query beatles: got %#v", names(got))
	}
}

func TestApplyExactMemberKeepsOnlyMatchingArtists(t *testing.T) {
	got := Apply(sampleArtists(), "phil", TypeMember, "Phil Collins")
	if len(got) != 1 || got[0].ID != 2 {
		t.Fatalf("exact member match: got %#v", names(got))
	}
}

func TestApplyEmptyQueryKeepsAll(t *testing.T) {
	views := sampleArtists()
	got := Apply(views, "  ", "", "")
	if len(got) != len(views) {
		t.Fatalf("empty query should keep all artists, got %d", len(got))
	}
}

func names(views []models.ArtistView) []string {
	out := make([]string, 0, len(views))
	for _, v := range views {
		out = append(out, v.Name)
	}
	return out
}

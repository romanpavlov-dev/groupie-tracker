// Package geo turns raw Groupie Tracker location keys (e.g. "germany-mainz")
// into {lat, lon} coordinates, geocoding each one exactly once via
// Nominatim and caching the result to disk.
//
// Nothing here runs on a live request path. Call EnsureCache() once,
// at startup, right after LoadStore() builds your views — then every
// handler afterwards just reads from the in-memory cache.
package geo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Coord is a plain lat/lon pair.
type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Marker is what gets attached to an ArtistView for the frontend map —
// one per concert location, with the dates that belong to it.
type Marker struct {
	Raw   string   `json:"raw"`  // "germany-mainz" — the original key
	Name  string   `json:"name"` // "Mainz, Germany" — for the popup label
	Lat   float64  `json:"lat"`
	Lon   float64  `json:"lon"`
	Dates []string `json:"dates"`
}

type nominatimResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

const cacheFilePath = "location_cache.json"

var (
	cache      = map[string]Coord{}
	cacheMutex sync.RWMutex
	loaded     bool
	loadOnce   sync.Once
)

// loadCache reads location_cache.json into memory. Safe to call
// multiple times — only actually reads the file once.
func loadCache() {
	loadOnce.Do(func() {
		data, err := os.ReadFile(cacheFilePath)
		if os.IsNotExist(err) {
			log.Println("geo: no existing location cache, starting fresh")
			loaded = true
			return
		}
		if err != nil {
			log.Printf("geo: failed to read cache file: %v\n", err)
			loaded = true
			return
		}

		cacheMutex.Lock()
		defer cacheMutex.Unlock()
		if err := json.Unmarshal(data, &cache); err != nil {
			log.Printf("geo: failed to parse cache file: %v\n", err)
		} else {
			log.Printf("geo: loaded %d cached locations\n", len(cache))
		}
		loaded = true
	})
}

func saveCache() error {
	cacheMutex.RLock()
	data, err := json.MarshalIndent(cache, "", "  ")
	cacheMutex.RUnlock()
	if err != nil {
		return fmt.Errorf("marshalling cache: %w", err)
	}
	return os.WriteFile(cacheFilePath, data, 0644)
}

// knownAliases maps raw location keys with outdated/unrecognizable
// country names (the Groupie Tracker API has some stale data) to a
// corrected query string that will actually resolve.
var knownAliases = map[string]string{
	// "Netherlands Antilles" was dissolved in 2010; Willemstad is now
	// the capital of Curaçao specifically.
	"willemstad-netherlands_antilles": "Willemstad, Curacao",
}

// CleanLocation turns "germany-mainz" into "Mainz, Germany" — the
// query string sent to the geocoder. Groupie Tracker's raw format is
// "country-city", dash-separated, with underscores for spaces.
func CleanLocation(raw string) string {
	if alias, ok := knownAliases[raw]; ok {
		return alias
	}

	parts := strings.SplitN(raw, "-", 2)
	for i := range parts {
		parts[i] = strings.ReplaceAll(parts[i], "_", " ")
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) == 2 {
		return fmt.Sprintf("%s, %s", strings.Title(parts[1]), strings.Title(parts[0]))
	}
	return strings.Title(parts[0])
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// geocodeQuery calls Nominatim for one cleaned query string.
func geocodeQuery(query string) (Coord, error) {
	endpoint := "https://nominatim.openstreetmap.org/search?format=json&limit=1&q=" +
		url.QueryEscape(query)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return Coord{}, err
	}
	// Required by Nominatim's usage policy.
	req.Header.Set("User-Agent", "groupie-tracker-geolocalization/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return Coord{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Coord{}, fmt.Errorf("nominatim returned status %d", resp.StatusCode)
	}

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return Coord{}, fmt.Errorf("decoding response: %w", err)
	}
	if len(results) == 0 {
		return Coord{}, fmt.Errorf("no results for %q", query)
	}

	var coord Coord
	if _, err := fmt.Sscanf(results[0].Lat, "%f", &coord.Lat); err != nil {
		return Coord{}, fmt.Errorf("parsing lat: %w", err)
	}
	if _, err := fmt.Sscanf(results[0].Lon, "%f", &coord.Lon); err != nil {
		return Coord{}, fmt.Errorf("parsing lon: %w", err)
	}
	return coord, nil
}

// GetCoords looks up a raw location key in the cache only — no
// network call. Safe to call from request handlers.
func GetCoords(raw string) (Coord, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	c, ok := cache[raw]
	return c, ok
}

// EnsureCache loads the on-disk cache (once) and geocodes any of the
// given raw locations that aren't cached yet, saving after each new
// batch. Call this once at startup, right after LoadStore() has
// built your artist views — pass every unique raw location string
// across all artists.
//
// Rate-limited to ~1 request/second to respect Nominatim's usage
// policy. Safe to call on every server start: already-cached
// locations are skipped instantly, so a warm cache costs nothing.
func EnsureCache(rawLocations []string) {
	loadCache()

	toFetch := make([]string, 0)
	cacheMutex.RLock()
	for _, raw := range rawLocations {
		if _, ok := cache[raw]; !ok {
			toFetch = append(toFetch, raw)
		}
	}
	cacheMutex.RUnlock()

	if len(toFetch) == 0 {
		log.Println("geo: cache already warm, nothing to geocode")
		return
	}

	log.Printf("geo: geocoding %d new locations...\n", len(toFetch))
	for _, raw := range toFetch {
		query := CleanLocation(raw)
		coord, err := geocodeQuery(query)
		if err != nil {
			log.Printf("geo:   failed to geocode %q (%q): %v\n", raw, query, err)
			continue
		}

		cacheMutex.Lock()
		cache[raw] = coord
		cacheMutex.Unlock()

		log.Printf("geo:   %-30s -> %.4f, %.4f\n", raw, coord.Lat, coord.Lon)
		time.Sleep(1100 * time.Millisecond)
	}

	if err := saveCache(); err != nil {
		log.Printf("geo: warning: failed to persist cache: %v\n", err)
	}
}

// BuildMarkers turns one artist's DateLocation map into markers,
// using only the cache. Locations with no cached coordinate (e.g.
// the geocoder had no match) are skipped with a log line rather than
// producing a marker at (0,0).
func BuildMarkers(dateLocation map[string][]string) []Marker {
	markers := make([]Marker, 0, len(dateLocation))

	for raw, dates := range dateLocation {
		coord, ok := GetCoords(raw)
		if !ok {
			log.Printf("geo: no cached coords for %q, skipping marker\n", raw)
			continue
		}
		markers = append(markers, Marker{
			Raw:   raw,
			Name:  CleanLocation(raw),
			Lat:   coord.Lat,
			Lon:   coord.Lon,
			Dates: dates,
		})
	}
	return markers
}

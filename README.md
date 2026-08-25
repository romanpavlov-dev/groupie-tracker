# Groupie Tracker — Geolocalization

A web app that displays artists/bands from the [Groupie Tracker API](https://groupietrackers.herokuapp.com/api), with each artist's concert locations plotted on an interactive map.

## Features

- Browse all artists with filtering (creation date, first album date, member count, location) and search
- Per-artist page showing members, concert dates, and locations
- **Interactive map** on each artist page with a marker for every concert location, using [Leaflet](https://leafletjs.com/) + [OpenStreetMap](https://www.openstreetmap.org/)

## Getting Started

```bash
go run main.go
```

The server listens on `http://localhost:1010`.

### ⏳ First run: wait 4–5 minutes before the maps are fully populated

On the **very first run**, the app needs to convert every concert location (e.g. `"germany-mainz"`) into geographic coordinates before it can place map markers. This is done via the [Nominatim](https://nominatim.openstreetmap.org/) geocoding API, which enforces a rate limit of **~1 request per second** to stay within its free usage policy.

With roughly 200–300 unique locations across all artists, this background process takes **about 4–5 minutes** to complete after the server starts.

**Important:** the server itself starts immediately — you don't have to wait to browse artists, view their pages, or use search/filters. Only the **map markers** are affected: while geocoding is still in progress, an artist's map may show **fewer pins than expected, or none at all**. Simply wait a few minutes and refresh the page — markers fill in progressively as their locations finish geocoding in the background.

Once this first run completes, all coordinates are saved to `location_cache.json` in the project root. **Every subsequent server restart is instant** — the app reads from this cache file instead of calling the geocoding API again.

> If you delete `location_cache.json`, the next startup will need to re-geocode everything from scratch, so expect another 4–5 minute wait.

## Project Structure

```
groupie-tracker/
├── main.go                        # entrypoint, route registration
├── location_cache.json            # auto-generated geocoding cache (git-ignored)
├── backend/
│   ├── api/
│   │   └── fetch.go               # fetches & assembles data from the Groupie Tracker API
│   ├── models/
│   │   └── models.go               # shared data structures
│   ├── filters/
│   │   └── filters.go              # filter parsing & application logic
│   ├── search/
│   │   └── search.go               # search & autocomplete logic
│   ├── geo/
│   │   └── geocode.go              # location string cleaning, geocoding, caching
│   └── handlers/
│       └── handlers.go             # HTTP handlers
└── frontend/
    └── templates/
        ├── mainpage.html            # artist list + filters
        └── artist.html              # single artist page + map
```

## How the Geolocation Feature Works

1. **`LoadStore()`** (in `fetch.go`) fetches artists, locations, dates, and relations from the Groupie Tracker API and builds an in-memory `Store`.
2. **`geo.WarmCacheAsync()`** is kicked off from `main.go` right after the store loads. It runs in a background goroutine — the HTTP server starts listening immediately, without waiting for it.
3. For each unique raw location string (e.g. `"usa-new_york"`), the background process:
   - Cleans it into a geocoder-friendly query (`"New York, Usa"`)
   - Sends it to Nominatim, respecting the 1 request/second rate limit
   - Stores the returned `{lat, lon}` in an in-memory cache, saved to `location_cache.json`
4. When a user visits an artist page, the frontend calls **`GET /api/artist/locations?id=<id>`**, which builds markers **live** from whatever is currently in the cache (a fast in-memory lookup — no network calls happen on this request path). Leaflet then renders a pin per location with a popup showing the formatted concert dates.

This design means:
- The server never blocks startup on external API calls
- Live requests never hit the geocoding API directly (no rate-limit risk, no slow page loads)
- Locations that fail to geocode (e.g. outdated or unrecognized place names) are logged and skipped rather than crashing anything

## API Endpoints

| Endpoint | Description |
|---|---|
| `GET /` | Artist list page |
| `GET /artist?id=<id>` | Single artist page with map |
| `GET /api/filter` | Filtered/searched artist list (JSON) |
| `GET /api/filters/meta` | Available filter ranges/options (JSON) |
| `GET /api/search` | Search suggestions (JSON) |
| `GET /api/artist/locations?id=<id>` | Concert location markers for one artist (JSON) |

## Notes

- Geocoding uses [Nominatim](https://nominatim.openstreetmap.org/), a free service — no API key required, but usage is rate-limited per their [usage policy](https://operations.osmfoundation.org/policies/nominatim/).
- Coordinates are cached indefinitely in `location_cache.json`, since concert city locations don't change.
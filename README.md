# Groupie Tracker

A small Go web application that pulls artist, concert location, concert date,
and tour-relation data from the public [Groupie Trackers API](https://groupietrackers.herokuapp.com/api)
and renders it as a browsable website: a main page listing all artists, and
a detail page per artist with their info, lineup, and concert history.

## Features

- Fetches artists, locations, dates, and relations from the upstream API
  concurrently at startup.
- In-memory store built once at server start — no database required.
- Main page: grid/list of all artists.
- Artist detail page: bio info, band members, and a concert list grouped
  by location with the dates played there.
- Raw location/date strings (e.g. `san_diego-usa`, `16-08-2019`) are
  formatted into readable form (`San Diego, Usa`, `16 Aug 2019`) on the
  client side with JavaScript.

## Project structure

```
groupie-tracker/
├── backend/
│   ├── models/    # Data shapes (Artist, ArtistView, API response types)
│   ├── api/       # Fetching from the upstream API + in-memory data store
│   └── logics/    # HTTP handlers (main page, artist detail page)
├── frontend/
│   └── templates/ # HTML templates (mainpage.html, artist.html)
└── main.go        # Entry point: loads data, registers routes, starts server
```

### Package responsibilities

- **`models`** — plain data structs only (`Artist`, `ArtistView`, `Index`,
  `LocationsResponse`, `DatesResponse`, `RelationsResponse`). No logic,
  no state.
- **`api`** — `FetchJSON` (generic HTTP GET + JSON decode helper) and the
  logic that loads all four upstream endpoints, joins them by artist ID,
  and builds the in-memory `Views` / `ViewsByID` collections used to serve
  every request.
- **`logics`** — HTTP handlers that read from the `api` package's store
  and render the HTML templates.

## Data flow

1. On startup, `main.go` calls `api.JsonToMap()`.
2. That function fetches the API index (`/api`), then fetches artists,
   locations, dates, and relations concurrently.
3. Location, date, and relation data is joined to each artist by ID into
   a single `ArtistView` per artist, and stored in two package-level
   lookups: `Views` (slice, for listing) and `ViewsByID` (map, for direct
   lookup by ID on the detail page).
4. The HTTP server starts only after this data is loaded, so all handlers
   read from data that's already fully populated.

## Routes

| Route      | Handler        | Description                                  |
|------------|----------------|-----------------------------------------------|
| `/`        | `MainHandler`  | Lists all artists                              |
| `/artist?id={id}` | `ArtistHandler` | Shows detail page for one artist (by numeric ID) |

## Running locally

Requires Go installed (1.18+ recommended).

```bash
git clone <repo-url>
cd groupie-tracker
go run main.go
```

The server starts on **port 1010**:

```
http://localhost:1010/
```

## Data source

All artist, location, date, and relation data comes from the public
Groupie Trackers API:

```
https://groupietrackers.herokuapp.com/api
```

No API key or authentication is required.

## Notes / known limitations

- Data is loaded once at startup and cached in memory for the lifetime of
  the server process — it does not auto-refresh. Restart the server to
  pick up upstream changes.
- Location/date formatting happens client-side in JavaScript rather than
  on the server, so raw values are briefly visible before formatting runs.
- Search and filtering (by artist name, location, or date/year) are
  planned but not yet implemented.

## Possible next steps

- Add search bar (by artist name) and filters (by location, by
  creation/first-album year, by concert date range).
- Add a `/reload` endpoint or periodic refresh if upstream data needs to
  stay current without a server restart.
- Add basic tests for the data-loading and formatting logic.

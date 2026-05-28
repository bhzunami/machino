# Machino Todos

Eine schlanke Client-Server-ToDo-App mit Go, gorilla/mux, gorilla/websocket und Svelte.

## Funktionen

- Registrierung, Login, Logout, Profilbearbeitung und Passwort-Reset-Flow
- Projekte mit Titel, Beschreibung, Farbe und benutzerspezifischem Favorit
- Todos mit Titel, Beschreibung, Faelligkeitsdatum, Prioritaet, Status und manueller Sortierung
- WebSocket-Updates pro Projekt, damit parallele Aenderungen sofort erscheinen
- PWA-Frontend mit IndexedDB-Cache und Offline-Warteschlange fuer lokale Arbeit

## Entwicklung starten

```bash
make install
go run ./cmd/api
```

In einem zweiten Terminal:

```bash
cd web
npm run dev
```

Das Backend laeuft standardmaessig auf `:8080`, das Frontend auf Vite. Die API wird im
Dev-Server ueber `/api` proxied.

## Produktion

```bash
cd web && npm run build
cd ..
go build -o bin/machino-api ./cmd/api
./bin/machino-api
```

Das Backend serviert `web/dist` automatisch. Konfiguration erfolgt ueber Umgebungsvariablen:
`HTTP_ADDR`, `DATABASE_PATH` und `STATIC_DIR`.

## Android und iOS

Die App ist bereits als PWA installierbar. Fuer App-Store-Pakete ist der naheliegende Weg
Capacitor: Svelte mit `npm run build` bauen, `web/dist` als Web-Asset einbinden und native
Android-/iOS-Projekte generieren. Das Go-Backend bleibt dabei unveraendert als API bestehen.

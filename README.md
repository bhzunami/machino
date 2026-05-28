# Machino — Mach I No

> *Schweizerdeutsch für «mach ich noch»*

Eine schlanke, kollaborative ToDo-App mit Echtzeit-Sync, Offline-Support und modernem Aurora-Dark-Design.

**Stack:** Go · gorilla/mux · gorilla/websocket · SQLite · Svelte 4 · Vite · PWA

---

## Funktionen

- **Projekte** mit Titel, Beschreibung und Farbe — als Favorit markierbar
- **Todos** mit Titel, Beschreibung, Fälligkeitsdatum, Priorität und manueller Sortierung per Drag & Drop
- **Echtzeit-Kollaboration** — mehrere Benutzer können gleichzeitig im selben Projekt arbeiten (WebSocket)
- **Offline-Support** — Änderungen werden lokal gecacht (IndexedDB) und beim Reconnect synchronisiert
- **Authentifizierung** — Registrierung, Login, Logout, Profil, Passwort-Reset per E-Mail
- **Sicherheit** — Rate-Limiting auf Auth-Endpunkten, Security-Header (CSP, HSTS), CSRF-Schutz
- **Mobile-freundlich** — responsives Design mit Slide-in-Sidebar

---

## Lokale Entwicklung

### Voraussetzungen

- Go 1.25+
- Node.js 20+

### Setup

```bash
# Abhängigkeiten installieren
make install

# Backend starten (Port 8080)
make run

# Frontend starten (eigenes Terminal, Port 5173)
cd web && npm run dev
```

Der Vite-Dev-Server proxied `/api` automatisch auf das Backend.
Öffne [http://localhost:5173](http://localhost:5173).

### Nützliche Befehle

```bash
make build        # Go-Binary bauen → bin/machino-api
make web-build    # Svelte bauen → web/dist
make test         # Tests ausführen
make fmt          # Go-Code formatieren
```

---

## Konfiguration (Umgebungsvariablen)

Kopiere `.env.example` nach `.env` und passe die Werte an:

```bash
cp .env.example .env
```

| Variable | Standard | Beschreibung |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Bind-Adresse des Servers |
| `DATABASE_PATH` | `machino.db` | Pfad zur SQLite-Datenbank |
| `STATIC_DIR` | `web/dist` | Pfad zum kompilierten Frontend |
| `REGISTRATION_ENABLED` | `true` | Registrierung erlauben (`false` nach Setup) |
| `COOKIE_SECURE` | `false` | `true` wenn hinter HTTPS-Proxy (Traefik) |
| `APP_DOMAIN` | `machino.localhost` | Domain für Traefik-Labels |
| `LOG_LEVEL` | `info` | `debug` · `info` · `warn` · `error` |
| `LOG_FORMAT` | `text` | `text` (Entwicklung) · `json` (Produktion) |
| `SMTP_HOST` | — | SMTP-Server (optional) |
| `SMTP_PORT` | `587` | SMTP-Port |
| `SMTP_USERNAME` | — | SMTP-Benutzername |
| `SMTP_PASSWORD` | — | SMTP-Passwort |
| `SMTP_FROM` | — | Absender-Adresse |

> Ohne SMTP-Konfiguration läuft der Passwort-Reset im **Demo-Modus**: der Reset-Token wird direkt in der API-Antwort zurückgegeben.

---

## Deployment mit Docker & Traefik

### Voraussetzungen auf dem Server

```bash
# Traefik-Netzwerk erstellen (einmalig)
docker network create traefik-public
```

Traefik muss mit `--certificatesresolvers.letsencrypt.acme.email=...` und den Entrypoints `web` (80) und `websecure` (443) konfiguriert sein.

### Deployment

```bash
# .env anlegen
cp .env.example .env
# APP_DOMAIN, SMTP_* etc. anpassen

# Image bauen und starten
docker compose up -d --build

# Logs beobachten
docker compose logs -f
```

Die App ist dann unter `https://$APP_DOMAIN` erreichbar. Let's-Encrypt-Zertifikat wird automatisch ausgestellt.

### Nach dem ersten Start

```bash
# Registrierung deaktivieren sobald alle Accounts angelegt sind
# In .env setzen:
REGISTRATION_ENABLED=false

docker compose up -d
```

### Datenbank-Backup

```bash
# Volume-Pfad ermitteln
docker volume inspect machino_machino_data

# SQLite-Datei kopieren (laufender Betrieb möglich — WAL-Modus)
cp /var/lib/docker/volumes/machino_machino_data/_data/machino.db ./backup-$(date +%Y%m%d).db
```

---

## Logging

Im **JSON-Format** (Produktion) lassen sich Logs einfach in Aggregatoren wie Loki, ELK oder CloudWatch weiterleiten:

```json
{"time":"2026-01-15T10:23:46Z","level":"INFO","msg":"http request","method":"POST","path":"/api/auth/login","status":200,"duration_ms":42,"ip":"192.168.1.1"}
{"time":"2026-01-15T10:23:46Z","level":"INFO","msg":"login success","email":"user@example.com","user_id":"abc123","ip":"192.168.1.1"}
{"time":"2026-01-15T10:23:50Z","level":"WARN","msg":"login failed","email":"unknown@x.com","ip":"10.0.0.5"}
```

Gelogte Auth-Events: `login success`, `login failed`, `user registered`, `logout`, `password reset requested`, `websocket connected/disconnected`.

---

## Mobile (Android & iOS)

Die App ist bereits als **PWA installierbar** — im Browser auf «Zum Homescreen hinzufügen».

Für native App-Store-Pakete:

1. Frontend bauen: `cd web && npm run build`
2. [Capacitor](https://capacitorjs.com/) einrichten: `npx cap init` → `web/dist` als Web-Asset
3. Native Projekte generieren: `npx cap add android` / `npx cap add ios`
4. Das Go-Backend bleibt als API unverändert bestehen

---

## Projektstruktur

```
machino/
├── cmd/api/          # Einstiegspunkt (main.go)
├── internal/
│   ├── handler/      # HTTP-Handler, Middleware (Auth, Rate-Limit, Logging)
│   ├── model/        # Datenmodelle
│   ├── realtime/     # WebSocket-Hub
│   ├── store/        # SQLite-Datenbankschicht
│   └── mailer/       # SMTP-Mailer
├── web/              # Svelte 4 Frontend
│   └── src/
│       └── components/
├── Dockerfile        # Multi-Stage Build
├── docker-compose.yml
└── .env.example
```

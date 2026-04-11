# Chirpy

A REST API built with Go and PostgreSQL.

## Stack

- Go (stdlib `net/http`)
- PostgreSQL
- [sqlc](https://sqlc.dev) — type-safe SQL
- [goose](https://github.com/pressly/goose) — migrations
- JWT authentication

## Setup

```bash
cp .env.example .env  # set DB_URL, JWT_SECRET, POLKA_KEY
goose -dir sql/schema postgres $DB_URL up
go run .
```

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/users` | — | Create user |
| PUT | `/api/users` | JWT | Update email/password |
| POST | `/api/login` | — | Login |
| POST | `/api/refresh` | — | Refresh access token |
| POST | `/api/revoke` | — | Revoke refresh token |
| POST | `/api/chirps` | JWT | Create chirp |
| GET | `/api/chirps` | — | List chirps (`?author_id=`, `?sort=asc\|desc`) |
| GET | `/api/chirps/{id}` | — | Get chirp |
| DELETE | `/api/chirps/{id}` | JWT | Delete chirp (author only) |
| POST | `/api/polka/webhooks` | API key | Upgrade user to Chirpy Red |

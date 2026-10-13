# Website Module

## Database

- `websites` table with soft delete support (`gorm.Model`)
- Unique index on `domain`
- Unique index on `website_key`
- Auto Migration and backfill migration scripts

## Model

- `Website` model (`internal/websites/model.go`)
- Platform constants (`WordPress`, `React`, `Python`, `Laravel`, `NodeJS`, `HTML`, `Java`, `VueJS`, `Go`, `Other`)
- Cryptographically secure random key generator (`GenerateWebsiteKey`) with collision retry logic

## DTO

- `CreateWebsiteRequest`
- `UpdateWebsiteRequest`
- `WebsiteResponse` (omits sensitive `website_key`)
- `WebsiteIntegrationResponse` (exposes `id`, `domain`, `website_key` exclusively on the protected integration endpoint)

## Repository

- `Create`
- `GetAll` (with search, sort, pagination)
- `GetByID`
- `GetByDomain`
- `GetByKey`
- `Update`
- `Delete`

## Routes & Endpoints

- `POST /api/v1/websites` — Create website (authenticated)
- `GET /api/v1/websites` — List websites (authenticated, paginated)
- `GET /api/v1/websites/:id` — Get website details (authenticated)
- `GET /api/v1/websites/:id/integration` — Get website integration details (authenticated)
- `PUT /api/v1/websites/:id` — Update website (authenticated)
- `DELETE /api/v1/websites/:id` — Soft-delete website (authenticated)

## Status

Completed ✅ (Pre-Phase 4 Hardened)
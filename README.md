# Pet project xkcd search

Simple pet project made for studying hexagonal architecture and microservices.

## Life cycle

1. **Update service** periodically fetches missing comics from `xkcd.com`, normalizes words via the `words` service, stores them in postgres.
2. **Search service** listens for nats messages and rebuilds its in memory index. Have 2 gRPC endpoints:
   - `Search` - via postgres (trigram + custom ranking)
   - `ISearch` - uses the local index for fast keyword matching
3. **API service** REST, concurrency limits, rate limits, and JWT authentication. Metrics via VictoriaMetrics and Prometheus

## Containers

- **api** - REST gateway (JWT auth, rate limiting, concurrency control)
- **words** - Text normalization (stemming, stop words, deduplication)
- **search** - Full-text search over comics (via PostgreSQL trigrams) + in memory index search
- **update** - Fetches new comics from XKCD, stores them, and triggers index rebuild
- **postgres** - Main database (comics, words, precomputed search index)
- **nats** - Message broker for notifying `search` when the database changes

## Technologies

- **Go** - All services written in Go
- **gRPC + Protocol Buffers** - Inter‑service communication
- **PostgreSQL** - Full‑text search and denormalized index table
- **NATS** - Publish‑subscribe between `update` and `search`
- **JWT** - API authentication (admin only)
- **snowball** - Stemming and stop‑word removal used inside the `words` service
- **golang‑migrate** - Schema migrations (run by `update` on startup)
- **VictoriaMetrics** - Metrics collection

## API Endpoints (REST, port 28080)

| Method | Path             | Description                                                  |
| ------ | ---------------- | ------------------------------------------------------------ |
| POST   | `/api/login`     | Authenticate admin, receive JWT token                        |
| GET    | `/api/ping`      | Health check (returns status of words/update/search)         |
| GET    | `/api/search`    | Full‑text search - query params: `?phrase=...&limit=...`     |
| GET    | `/api/isearch`   | Index search (in memory index) - same params                 |
| POST   | `/api/db/update` | Trigger background update of all new comics (requires token) |
| DELETE | `/api/db`        | Drop the whole comics table (requires token)                 |
| GET    | `/api/db/stats`  | Show word/comic statistics                                   |
| GET    | `/api/db/status` | Check if an update is currently running (`idle`/`running`)   |

All protected endpoints (except `/login` and `/ping`) require the `Authorization: Token <jwt>` header.

Commits from 8 repositories squished and merged into one.
Made as part of a course by yadro.

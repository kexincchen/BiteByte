## 1. Start PostgreSQL via Docker Compose

```bash
docker compose --env-file .env up -d
```
## 2. Test database connection in Go
```bash
source .env
env | grep -E 'POSTGRES_(USER|PASSWORD|DB|HOST|PORT)'
go run ./cmd/pingdb # this should show 'Postgres connection OK'
```


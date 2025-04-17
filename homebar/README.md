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

## 3. Create table in the PostgreSQL on macOS
```bash
brew update
brew install postgresql@16
echo 'export PATH="/opt/homebrew/opt/postgresql@16/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## 4. Connect to the database and list out tables
```bash
psql -h localhost -U admin -d bitebyte
\dt
```

## 5. Create table if it does not exist
```bash
CREATE TABLE IF NOT EXISTS products (
  id SERIAL PRIMARY KEY,
  merchant_id INT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  price NUMERIC(10,2) NOT NULL,
  category TEXT,
  image_url TEXT,
  is_available BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id         SERIAL PRIMARY KEY,
  username   TEXT      NOT NULL UNIQUE,
  email      TEXT      NOT NULL UNIQUE,
  password   TEXT      NOT NULL,
  role       TEXT      NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);  

```
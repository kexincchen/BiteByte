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
  mime_type TEXT,
  image_data BYTEA,
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

CREATE TABLE IF NOT EXISTS orders (
  id           SERIAL PRIMARY KEY,
  customer_id  INT      NOT NULL,
  merchant_id  INT      NOT NULL,
  total_amount NUMERIC(10,2) NOT NULL,
  status       TEXT     NOT NULL,
  notes        TEXT,
  created_at   TIMESTAMPTZ NOT NULL,
  updated_at   TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS order_items (
  id         SERIAL PRIMARY KEY,
  order_id   INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id INT NOT NULL,
  quantity   INT NOT NULL,
  price      NUMERIC(10,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS customers (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    address TEXT,
    phone VARCHAR(50)
);
  
CREATE TABLE IF NOT EXISTS merchants (
  id            SERIAL PRIMARY KEY,
  user_id       INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  business_name TEXT NOT NULL,
  description   TEXT,
  address       TEXT,
  phone         TEXT,
  username      TEXT UNIQUE NOT NULL,
  is_verified   BOOLEAN DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL,
  updated_at    TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS ingredients (
    id SERIAL PRIMARY KEY,
    merchant_id INTEGER NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    quantity NUMERIC(10, 2) NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL,
    low_stock_threshold NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create product_ingredients table
CREATE TABLE IF NOT EXISTS product_ingredients (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    ingredient_id INTEGER NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    quantity NUMERIC(10, 2) NOT NULL,
);

ALTER TABLE product_ingredients 
ADD CONSTRAINT unique_product_ingredient 
UNIQUE (product_id, ingredient_id);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_ingredients_merchant_id ON ingredients(merchant_id);
CREATE INDEX IF NOT EXISTS idx_product_ingredients_product_id ON product_ingredients(product_id);
CREATE INDEX IF NOT EXISTS idx_product_ingredients_ingredient_id ON product_ingredients(ingredient_id); 
  
```
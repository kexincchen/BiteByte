# Internal

This directory contains the core logic of the application.

## API

The `api` package contains the API handlers for the application.

Here are the handlers for the following endpoints:

- `/api/merchants`
- `/api/merchants/:id`
- `/api/merchants/:id/products`
- `/api/merchants/:id/products/:id`
- `/api/merchants/:id/products/:id/orders`

## DB

The `db` package contains the database connection and the repository implementations.

Here are the repositories for the following entities:

- `Merchant`
- `Product`
- `Order`
- `OrderItem`

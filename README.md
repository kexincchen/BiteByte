
# Home Bar Ordering Distributed Project - BiteByte

## Project Overview

This is a Home Bar ordering distributed system using Raft algorithm to achieve consensus that supports merchant menu management, customer ordering, inventory management, and other essential features. It also supports user authentication, order status tracking, and other common e-commerce functionalities. The system adopts a modern architecture design with high concurrency processing capabilities and good scalability.

## Core Features

### User Management
- User registration and authentication
- Multi-role support (customer, merchant)
- User profile management

### Product Management
- Merchant menu upload and management
- Product categorization and search
- Product details display

### Order System
- Customer ordering process
- Order status tracking
- Order history queries

### Inventory Management
- Real-time ingredient inventory tracking
- Automatic ingredient stock verification during ordering
- Low inventory alert notifications
- Inventory transaction history

### Notification System
- Order status change notifications
- Inventory alert notifications
- Merchant fulfillment notifications

### Review System
- Customer reviews and ratings
- Review management

## Technical Architecture

### Architecture Diagram

The project adopts a layered architecture, primarily divided into: Client Layer, API Layer, Service Layer, Core Layer, and Infrastructure Layer.

- **Client Layer**: Web interface, mobile applications
- **API Layer**: API gateway, authentication middleware
- **Service Layer**: User service, merchant service, product service, order service, payment service, notification service, review service, inventory management service
- **Core Layer**: Domain models, repository interfaces
- **Infrastructure Layer**: Database, cache, message queue, object storage, payment gateway integration

### Order Inventory Verification Process

When a customer places an order, the system will:
1. Retrieve information on all ingredients required for the products
2. Verify if the inventory for each ingredient is sufficient
3. If all ingredient inventory is sufficient, reserve the inventory and notify the merchant for fulfillment
4. If some ingredients are low in stock, send a low inventory warning to the merchant and notify the customer of possible delivery delays

### Technology Stack

- **Language**: Go
- **Web Framework**: Gin
- **Frontend**: React
- **Service Communication**: gRPC
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Queue**: Kafka
- **Authentication**: JWT
- **Consensus Algorithm**: Raft

## Project Structure

```
homebar/
├── cmd/                # Application entry points
│   └── server/         # API server
├── internal/           # Internal packages
│   ├── domain/         # Domain models/entities
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic layer
│   ├── api/            # API handlers
│   ├── middleware/     # HTTP middleware
│   └── config/         # Configuration
├── pkg/                # Exportable public packages
├── scripts/            # Build and deployment scripts
├── configs/            # Configuration files
├── migrations/         # Database migrations
└── web/                # Frontend resources
```

## Domain Models

The system includes the following main domain models:

- **User**: Basic user information
- **Customer**: Customer-specific information
- **Merchant**: Merchant-specific information
- **Product**: Product information
- **Ingredient**: Ingredient information
- **Inventory**: Inventory information
- **ProductIngredient**: Product and ingredient association
- **InventoryTransaction**: Inventory transaction records
- **Order**: Order information
- **OrderItem**: Order items
- **Notification**: Notification information

## Development Notes

### Development Recommendations

- Follow Go's "composition over inheritance" principle when designing domain models
- Fully utilize Go's concurrency features for high-concurrency scenarios (like order peaks)
- Use interfaces to define interaction contracts between layers
- Follow the single responsibility principle when designing services

## Future Plans

- Add big data analytics functionality
- Integrate recommendation systems
- Develop mobile management applications for merchants
- Add multi-language support


## Project Setup

### Prerequisites

- Go 1.20+
- Node.js 18+

### Backend Setup

1. Clone the repository

```bash
git clone https://github.com/yourusername/homebar.git
```

2. Install dependencies

```bash
cd homebar
go mod download
go mod tidy
```

3. Run the server

```bash
go run cmd/server/main.go
```

### Frontend Setup

1. Install dependencies

```bash
cd homebar/web
npm install
```

2. Start the development server

```bash
npm start
```

## Start the Raft cluster

```bash
bash homebar/scripts/monitor_raft.sh
```

## Run the server with multiple nodes

```
NODE_ID=1 PORT=8080 go run cmd/server/main.go
NODE_ID=2 PORT=8081 go run cmd/server/main.go
NODE_ID=3 PORT=8082 go run cmd/server/main.go
```



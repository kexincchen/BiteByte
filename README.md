
# Home Bar Ordering Distributed Project - BiteByte

## Project Overview

The BiteByte project implements a distributed home bar ordering system using the **Raft consensus algorithm** to ensure data consistency across multiple nodes. The system follows a clean, layered architecture that separates concerns and promotes maintainability.

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

## Technical Architecture

### Structure 

The project adopts a layered architecture, primarily divided into: Client Layer, API Layer, Service Layer, and Core Layer.

- **Client Layer**: Web interface, mobile applications
- **API Layer**: API gateway, authentication middleware
- **Service Layer**: User service, merchant service, product service, order service, payment service, notification service, review service, inventory management service
- **Core Layer**: Domain models, repository interfaces

### Order Inventory Verification Process

When a customer places an order, the system will:
1. Retrieve information on all ingredients required for the products
2. Verify if the inventory for each ingredient is sufficient
3. If all ingredient inventory is sufficient, reserve the inventory and notify the merchant for fulfillment
4. If some ingredients are low in stock, send a low inventory warning to the merchant 

### Technology Stack

- **Language**: Go
- **Web Framework**: Gin
- **Frontend**: React
- **Database**: PostgreSQL
- **Consensus Algorithm**: Raft

## Project Structure

```
homebar/
├── backend/                # Backend implementation
│   ├── cmd/                # Application entry points
│   │   ├── server/         # API server
│   │   └── pingdb/         # PingDB implementation
│   ├── internal/           # Internal packages
│   │   ├── domain/         # Domain models/entities
│   │   ├── db/             # Database implementation
│   │   ├── raft/           # Raft consensus implementation
│   │   ├── repository/     # Data access layer
│   │   ├── service/        # Business logic layer
│   │   ├── api/            # API handlers
│   │   └── config/         # Configuration
│   └── scripts/            # Scripts for monitoring and testing
├── web/                    # Frontend resources
│   ├── public/             # Static files
│   └── src/                # Source code
├── tests/                  # Test files
├── docker-compose.yml      # Docker Compose file
├── README.md               # Project overview
└── .gitignore              # Git ignore rules
```

## Domain Models

The system includes the following main domain models:

- **User**: Basic user information
- **Customer**: Customer-specific information
- **Merchant**: Merchant-specific information
- **Product**: Product information
- **Ingredient**: Ingredient information
- **ProductIngredient**: Product and ingredient association
- **Order**: Order information

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

```bash
cd homebar/backend
source .env
NODE_ID=1 PORT=8080 go run cmd/server/main.go
NODE_ID=2 PORT=8081 go run cmd/server/main.go
NODE_ID=3 PORT=8082 go run cmd/server/main.go
```



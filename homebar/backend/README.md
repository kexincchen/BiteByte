# Homebar Backend

## Overview

The Homebar backend implements a distributed home bar ordering system using the Raft consensus algorithm to ensure data consistency across multiple nodes. 


## Core Components

### Domain Models
- `User`: Basic user information with authentication details
- `Customer`: Customer-specific information linked to a user
- `Merchant`: Merchant-specific information linked to a user
- `Product`: Product information with details like price, description, etc.
- `Ingredient`: Ingredient information for inventory tracking
- `Order`: Order information with details about products ordered

### API Handlers

The API handlers are defined in the `internal/api` package:

- `UserHandler`: Handles user registration, login
- `MerchantHandler`: Manages merchant profiles
- `ProductHandler`: Manages product creation, updates, and queries
- `OrderHandler`: Processes order creation and management
- `IngredientHandler`: Manages inventory ingredients
- `ProductIngredientHandler`: Associates products with ingredients

### Services

Services in the `internal/service` package implement business logic:

- `UserService`: User authentication and management
- `ProductService`: Product management
- `OrderService`: Order processing and status management
- `MerchantService`: Merchant profile management
- `IngredientService`: Inventory management
- `RaftService`: Distributed consensus service

### Raft Consensus

The application implements a distributed consensus algorithm (Raft) to ensure consistency across multiple nodes. Key components include:

- `RaftNode`: Core implementation of the Raft protocol
- `ClusterCoordinator`: Manages communication between nodes
- `Leader Election`: Ensures a single leader for write operations
- `Request Forwarding`: Redirects requests to the leader node
- `RaftService`: Manages Raft group membership and consensus
- `RaftStorage`: Persists Raft state and log


## API Endpoints

### Authentication

```
POST /api/auth/register - Register a new user (customer or merchant)
POST /api/auth/login - Authenticate a user
```

### Products

```
GET /api/products - List all products
GET /api/products/:id - Get product details
POST /api/products - Create a new product
PUT /api/products/:id - Update product details
DELETE /api/products/:id - Delete a product
GET /api/products/merchant/:id - Get products by merchant
GET /api/products/:id/image - Get product image
POST /api/products/availability - Check product availability
```

### Orders

```
GET /api/orders - List orders
GET /api/orders/:id - Get order details
POST /api/orders - Create a new order
PUT /api/orders/:id - Update order details
PUT /api/orders/:id/status - Update order status
DELETE /api/orders/:id - Delete an order
```

### Merchants

```
GET /api/merchants - List all merchants
GET /api/merchants/:id - Get merchant details
POST /api/merchants - Create a merchant profile
GET /api/merchants/username/:username - Get merchant by username
GET /api/merchants/user/:userID - Get merchant by user ID
```

### Inventory

```
GET /api/merchants/:id/inventory - List ingredients for a merchant
POST /api/merchants/:id/inventory - Add an ingredient
GET /api/merchants/:id/inventory/summary - Get inventory summary
GET /api/merchants/:id/inventory/:ingredientId - Get ingredient details
PUT /api/merchants/:id/inventory/:ingredientId - Update ingredient
DELETE /api/merchants/:id/inventory/:ingredientId - Delete ingredient
```


### Product Ingredients

```
GET /api/products/:id/ingredients - Get ingredients for a product
POST /api/products/:id/ingredients - Add ingredient to product
PUT /api/products/:id/ingredients/:ingredientId - Update product ingredient
DELETE /api/products/:id/ingredients/:ingredientId - Remove ingredient from product
```

### Raft Consensus Implementation

The backend implements the Raft consensus algorithm to ensure consistency across distributed nodes:

- `Leader Election`: The system elects a leader responsible for processing write operations
- `Request Forwarding`: Non-leader nodes forward write requests to the current leader
- `Consistency`: The leader ensures data changes are replicated to followers
- `Fault Tolerance`: The system continues to operate if nodes fail (as long as majority remains)

During order processing, the system:
- Validates the order details
- Checks inventory availability
- Reserves ingredients if available
- Processes the order through the Raft consensus protocol
- Updates inventory accordingly

### Health Check

The system provides a health check endpoint at `/health` that returns a 200 OK status when the service is operating correctly.

### Raft Cluster Monitoring

The `scripts/monitor_raft.sh` script provides a real-time view of the Raft cluster status:

- Displays active nodes and their states
- Shows current leader information
- Provides recent log entries

To run the monitoring script:

```
./scripts/monitor_raft.sh
```

```bash
cd Desktop/demo/homebar/backend
source .env
```

```bash
NODE_ID=1 PORT=9001 go run cmd/server/main.go
```

```bash
NODE_ID=2 PORT=9002 go run cmd/server/main.go
```

```bash
NODE_ID=3 PORT=9003 go run cmd/server/main.go
```

```bash
curl http://127.0.0.1:8091/cluster/status
curl http://127.0.0.1:8092/cluster/status
curl http://127.0.0.1:8093/cluster/status
```
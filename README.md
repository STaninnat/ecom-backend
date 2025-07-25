# ecom-backend

[![Build Status](https://github.com/STaninnat/ecom-backend/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/STaninnat/ecom-backend/actions)
[![Coverage Status](https://img.shields.io/codecov/c/github/STaninnat/ecom-backend/master.svg)](https://codecov.io/gh/STaninnat/ecom-backend)

Welcome to my personal e-commerce backend project. This is more than just a codebase—it's my playground for learning, experimenting, and building something real from scratch. I wanted to see how far I could take a Go backend, and this is the result!

## 🚀 Features (with Details)

- **User Authentication**: JWT-based auth, refresh tokens, and Google OAuth. Secure, stateless, and supports role-based access (admin/user).
- **Product & Category Management**: CRUD for products and categories, with admin-only endpoints for creation and updates. Public endpoints are cached for performance.
- **Cart System**: Supports both authenticated user carts (MongoDB) and guest carts (session-based). Handles merging carts on login.
- **Order Management**: Users can place orders, view their order history, and admins can manage all orders.
- **Payment Integration**: Stripe for payment intents, confirmations, refunds, and webhook handling.
- **File Uploads**: Product images can be uploaded to local storage or AWS S3, with the backend auto-detecting which to use.
- **Reviews**: Users can leave reviews (with ratings and media) on products. Supports filtering, pagination, and moderation.
- **Robust Middleware**: Logging, security headers, rate limiting (Redis), CORS, request IDs, error handling, and more.
- **API Documentation**: Swagger/OpenAPI docs auto-generated and browsable at `/v1/swagger/index.html`.
- **Testing & Quality**: Extensive unit and integration tests, code coverage, and CI with GitHub Actions.

## 🛠️ Tech Stack (and Why)

- **Go**: Fast, simple, and great for backend services. I wanted to master Go’s concurrency and type system.
- **PostgreSQL**: Reliable, powerful relational DB. Used for core business data (users, products, orders).
- **MongoDB**: Flexible NoSQL for features like carts and reviews that benefit from document storage.
- **Redis**: Lightning-fast cache and rate limiter. Keeps things snappy and secure.
- **AWS S3**: Industry-standard for file storage. Optional, but great for production.
- **Stripe**: Real-world payment processing, with webhooks and refunds.
- **chi**: Lightweight, idiomatic Go router.
- **logrus**: Structured logging for observability.
- **sqlc**: Type-safe SQL for Go—no more hand-written queries.
- **swaggo**: Auto-generates OpenAPI docs from code comments.

## 🏗️ Architecture Overview

The project is organized for clarity, modularity, and testability. Here’s a high-level view:

```txt
main.go
  ├── auth/                 # Auth logic, JWT, OAuth, password, cookies
  ├── handlers/
        ├── auth/           # Auth endpoints
        ├── cart/           # Cart endpoints
        ├── category/       # Category endpoints
        ├── order/          # Order endpoints
        ├── payment/        # Payment endpoints
        ├── product/        # Product endpoints
        ├── review/         # Review endpoints
        ├── upload/         # File upload endpoints
        └── user/           # User endpoints
  ├── internal/
        ├── config/         # Loads env vars, sets up DB/Redis/Mongo/S3/Stripe, DI
        ├── database/       # SQLC-generated DB access, models
        ├── mongo/          # MongoDB integration
        ├── router/         # Route registration, adapters, middleware wiring
        └── testutil/       # Test helpers
  ├── logs/                 # Log output (rotated)
  ├── middlewares/          # HTTP middleware (auth, logging, security, etc)
  ├── models/               # Data models for API and DB
  ├── uploads/, images/     # File storage (local)
  ├── utils/                # Helpers: cache, logger, shutdown, null types, etc
```

- **main.go**: Entry point, loads config, sets up logger, starts server.
- **internal/config**: Loads environment variables, sets up all core services, dependency injection.
- **handlers/**: All business logic for API endpoints, organized by resource.
- **middlewares/**: Cross-cutting concerns (auth, logging, rate limiting, etc).
- **internal/router**: Route registration, adapters, and middleware wiring.
- **models/**: Data models for API and DB.
- **utils/**: Helpers for logging, caching, UUIDs, etc.
- **auth/**: Auth logic, JWT, OAuth, password, cookies.

---

## 📁 Directory Structure

```code
handlers/         # All API endpoint logic, grouped by resource
internal/         # Core infrastructure: config, router, database, mongo, testutils
models/           # Data models for API and DB
middlewares/      # HTTP middleware (auth, logging, security, etc)
auth/             # Authentication, JWT, OAuth, password, cookies
utils/            # Helpers: cache, logger, shutdown, null types, etc
uploads/, images/ # File storage (local)
logs/             # Log output (rotated)
.sqlc.yaml        # SQLC config for DB codegen
.golangci.yml     # Linter config
.github/          # CI/CD workflows
```

## 🏃 Getting Started (Step-by-Step)

1. **Clone the repo**
2. **Set up environment variables**
   - Copy `.env.example` (if present) or check `internal/config` for required keys (DB URLs, JWT secrets, Stripe keys, etc).
3. **Run dependencies**
   - Easiest: use Docker Compose for PostgreSQL, MongoDB, and Redis.
   - Or install them locally and set the connection URLs.
4. **Run database migrations**
   - (If using SQLC, run `sqlc generate` to update Go code from SQL.)
5. **Install Go dependencies**
   - `go mod download`
6. **Run the server**
   - `go run main.go`
7. **Access API docs**
   - [http://localhost:8080/v1/swagger/index.html](http://localhost:8080/v1/swagger/index.html)

## 📚 API Usage (Examples)

Here are a few example requests to get you started:

- **Signup**

  ```http
  POST /v1/auth/signup
  Content-Type: application/json
  {
    "name": "Alice",
    "email": "alice@example.com",
    "password": "supersecret"
  }
  // Response: 201 Created
  {
    "message": "Signup successful"
  }
  ```

- **Signin**

  ```http
  POST /v1/auth/signin
  Content-Type: application/json
  {
    "email": "alice@example.com",
    "password": "supersecret"
  }
  // Response: 200 OK
  // (Sets JWT and refresh token cookies)
  {
    "message": "Signin successful"
  }
  ```

- **Get Products**

  ```http
  GET /v1/products/
  Authorization: Bearer <JWT>
  // Response: 200 OK
  [
    {
      "id": "prod_123",
      "name": "Cool T-shirt",
      "price": 19.99,
      ...
    },
    ...
  ]
  ```

- **Create Order**

  ```http
  POST /v1/orders/
  Authorization: Bearer <JWT>
  Content-Type: application/json
  {
    "items": [
      { "product_id": "prod_123", "quantity": 2 }
    ],
    "shipping_address": "123 Main St, City, Country"
  }
  // Response: 201 Created
  {
    "order_id": "order_456",
    "message": "Order created successfully"
  }
  ```

- **Swagger UI**
  - Browse and try all endpoints at: [http://localhost:8080/v1/swagger/index.html](http://localhost:8080/v1/swagger/index.html)

## 🧪 Testing & Quality

- **Run all tests**: `go test ./...`
- **Lint**: `golangci-lint run` (see `.golangci.yml` for rules)
- **Code coverage**: `go test -coverprofile=coverage.out ./...` (view with `go tool cover -html=coverage.out`)
- **CI/CD**: Automated with GitHub Actions (`.github/workflows/ci.yml`)
- **Codecov**: Coverage reporting integrated

## 🧩 Extending the Project

Want to add a new resource (e.g., wishlist)?

1. Add a new handler in `handlers/wishlist/`
2. Define models in `models/`
3. Add DB queries (SQLC or MongoDB)
4. Register routes in `internal/router/router.go`
5. Add tests in `handlers/wishlist/`
6. Update Swagger docs/comments

## 💡 Personal Reflections

This project taught me:

- How to design a real-world Go backend from scratch
- The value of modularity, dependency injection, and clean code
- How to integrate with real services (Stripe, S3, Redis, etc)
- The importance of good logging, error handling, and testing

**What was hard?**

- Balancing flexibility with simplicity
- Managing config and secrets for local vs. cloud
- Writing good tests for async and external-service code

**What’s next?**

- Add more e2e tests
- Explore GraphQL or gRPC APIs
- Deploy to cloud (Kubernetes, AWS, etc)
- Open source the frontend!

---

If you made it this far, thanks for reading! If you want to chat, collaborate, or just geek out about Go/backend stuff, reach out any time. PRs and feedback are always welcome!

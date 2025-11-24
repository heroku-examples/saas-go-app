# SaaS Go App

A SaaS web application backend built with Go and Gin, featuring Vue.js frontend, Heroku Postgres Advanced with follower pools, Redis for background jobs, and JWT authentication.

## Features

- **RESTful API** with CRUD operations for customers and accounts
- **JWT Authentication** for secure API access
- **Background Jobs** using Asynq for data aggregation
- **Analytics Endpoints** that read from follower pools
- **Health Checks** and Prometheus metrics
- **Vue.js Frontend** with Bootstrap styling

## Tech Stack

- **Backend**: Go + Gin
- **Frontend**: Vue.js 3 + Bootstrap 5
- **Database**: PostgreSQL (Heroku Postgres Advanced)
- **Cache/Jobs**: Redis + Asynq
- **Monitoring**: Prometheus

## Project Structure

```
saas-go-app/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/                 # API handlers
│   ├── auth/                # JWT authentication
│   ├── db/                  # Database connection and migrations
│   ├── jobs/                # Background job handlers
│   └── models/              # Data models
├── web/
│   └── frontend/            # Vue.js frontend application
├── Makefile                 # Common tasks
├── go.mod                   # Go dependencies
└── README.md
```

## Setup

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL (or Heroku Postgres)
- Redis (optional, for background jobs)

### Backend Setup

1. Install dependencies:
```bash
make deps
```

2. Create a `.env` file (see `env.example`):
```bash
# For local development, set up PostgreSQL and Redis locally
# Or use Docker:
#   docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres
#   docker run -d -p 6379:6379 redis

DATABASE_URL=postgres://user:password@localhost:5432/saas_go_app?sslmode=disable
ANALYTICS_DB_URL=postgres://user:password@localhost:5432/saas_go_app?sslmode=disable  # Can use same DB for local dev
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=your-secret-key-change-in-production
PORT=8080
```

**Note**: On Heroku, `DATABASE_URL`, `ANALYTICS_DB_URL`, `REDIS_URL`, and `PORT` are automatically set by Heroku addons. You only need to set `JWT_SECRET` manually.

3. Run the server:
```bash
make run
# or
go run ./cmd/server
```

The server will automatically create database tables on startup.

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd web/frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

The frontend will be available at `http://localhost:3000` and will proxy API requests to `http://localhost:8080`.

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login and get JWT token
- `POST /api/auth/register` - Register a new user

### Customers (Protected)
- `GET /api/customers` - Get all customers
- `GET /api/customers/:id` - Get customer by ID
- `POST /api/customers` - Create a new customer
- `PUT /api/customers/:id` - Update customer
- `DELETE /api/customers/:id` - Delete customer

### Accounts (Protected)
- `GET /api/accounts` - Get all accounts
- `GET /api/accounts/:id` - Get account by ID
- `POST /api/accounts` - Create a new account
- `PUT /api/accounts/:id` - Update account
- `DELETE /api/accounts/:id` - Delete account

### Analytics (Protected)
- `GET /api/analytics` - Get overall analytics
- `GET /api/analytics/customers/:customer_id` - Get customer-specific analytics

### Health & Metrics
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics

## Heroku Deployment

> **📘 For detailed Heroku deployment instructions, see [HEROKU_SETUP.md](HEROKU_SETUP.md)**

### Quick Setup

```bash
# Login to Heroku
heroku login

# Create Heroku app
heroku create saas-go-app

# Provision Postgres Advanced
heroku addons:create heroku-postgresql:standard-0 --name saas-go-db

# (Optional) Provision Redis
heroku addons:create heroku-redis:hobby-dev

# Get DATABASE_URL
heroku config:get DATABASE_URL

# Set up follower pool for analytics (via Heroku Dashboard)
# Go to: https://dashboard.heroku.com/apps/saas-go-app
# Click on the database addon → Create Follower/Follower Pool
# Or check HEROKU_SETUP.md for detailed instructions

# Get ANALYTICS_DB_URL from follower
heroku config:get ANALYTICS_DB_URL

# Set environment variables
heroku config:set JWT_SECRET=your-production-secret-key

# Deploy
git push heroku main

# Open app
heroku open
```

**Note**: The `Procfile` tells Heroku how to run your app. Heroku's Go buildpack will automatically detect `go.mod` and build your application. The binary name matches your module name (`saas-go-app`).

### Environment Variables on Heroku

**Automatically Set by Heroku:**
- `DATABASE_URL` - Automatically set by `heroku-postgresql` addon (primary database)
- `ANALYTICS_DB_URL` - Automatically set when you create a follower pool (via Heroku Dashboard)
- `REDIS_URL` - Automatically set by `heroku-redis` addon (if provisioned)
- `PORT` - Automatically set by Heroku platform

**Manually Configured:**
- `JWT_SECRET` - **You must set this manually**: `heroku config:set JWT_SECRET=your-secret-key`

**Important**: When you provision Heroku Postgres Advanced and create a follower pool, Heroku automatically provides the connection URLs. You don't need to manually configure `DATABASE_URL` or `ANALYTICS_DB_URL` - they're set automatically by the addons.

## Development

### Running Tests

```bash
make test
# or with coverage
make test-coverage
```

### Building

```bash
make build
```

### Code Formatting

```bash
make fmt
```

## Background Jobs

Background jobs are processed using Asynq. Jobs are enqueued for data aggregation tasks. The job processor runs automatically when `REDIS_URL` is configured.

Example: Enqueue an aggregation task (can be added to API handlers):

```go
import "saas-go-app/internal/jobs"

client, _ := jobs.NewClient(os.Getenv("REDIS_URL"))
jobs.EnqueueAggregationTask(client, time.Now())
```

## License

MIT


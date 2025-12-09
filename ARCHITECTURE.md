# Architecture Documentation

This document provides a comprehensive overview of the SaaS Go App architecture, with a focus on the leader/follower database pattern implementation.

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Components](#architecture-components)
3. [Leader/Follower Database Pattern](#leaderfollower-database-pattern)
4. [Data Flow](#data-flow)
5. [Request Routing](#request-routing)
6. [Database Connection Management](#database-connection-management)
7. [Background Jobs](#background-jobs)
8. [Authentication & Authorization](#authentication--authorization)
9. [Deployment Architecture](#deployment-architecture)

---

## System Overview

The SaaS Go App is a full-stack web application built with:
- **Backend**: Go (Golang) with Gin web framework
- **Frontend**: Vue.js 3 with Bootstrap 5
- **Primary Database**: PostgreSQL (Heroku Postgres)
- **Analytics Database**: PostgreSQL Follower Pool (optional)
- **Job Queue**: Redis with Asynq (optional)
- **Monitoring**: Prometheus metrics
- **Documentation**: Swagger/OpenAPI

```mermaid
graph TB
    subgraph "Client Layer"
        Browser[Web Browser]
        API_Client[API Client]
    end
    
    subgraph "Application Layer"
        Frontend[Vue.js Frontend<br/>Static Files]
        Backend[Go Backend<br/>Gin Server]
    end
    
    subgraph "Data Layer"
        PrimaryDB[(Primary PostgreSQL<br/>Leader)]
        FollowerDB[(Follower Pool<br/>Read Replica)]
        Redis[(Redis<br/>Job Queue)]
    end
    
    Browser --> Frontend
    API_Client --> Backend
    Frontend --> Backend
    Backend --> PrimaryDB
    Backend --> FollowerDB
    Backend --> Redis
    
    PrimaryDB -.->|Replication| FollowerDB
```

---

## Architecture Components

### Component Diagram

```mermaid
graph LR
    subgraph "Frontend (Vue.js)"
        UI[User Interface]
        Router[Vue Router]
        API_Client[Axios Client]
    end
    
    subgraph "Backend (Go/Gin)"
        Router_Gin[Gin Router]
        Auth[Auth Middleware]
        Handlers[API Handlers]
        DB_Layer[Database Layer]
        Job_Processor[Job Processor]
    end
    
    subgraph "Storage"
        Primary[(Primary DB)]
        Follower[(Follower Pool)]
        Redis_Store[(Redis)]
    end
    
    UI --> Router
    Router --> API_Client
    API_Client --> Router_Gin
    Router_Gin --> Auth
    Auth --> Handlers
    Handlers --> DB_Layer
    DB_Layer --> Primary
    DB_Layer --> Follower
    Handlers --> Job_Processor
    Job_Processor --> Redis_Store
```

### Key Components

#### 1. **Frontend (Vue.js)**
- **Location**: `web/frontend/`
- **Purpose**: Single Page Application (SPA) for user interaction
- **Features**:
  - Dashboard with analytics visualization
  - Customer and Account management UI
  - JWT token-based authentication
  - Responsive Bootstrap styling

#### 2. **Backend API (Go/Gin)**
- **Location**: `main.go`, `internal/api/`
- **Purpose**: RESTful API server
- **Features**:
  - JWT authentication
  - CRUD operations for customers and accounts
  - Analytics endpoints
  - Health checks and metrics
  - Swagger documentation

#### 3. **Database Layer**
- **Location**: `internal/db/`
- **Purpose**: Database connection and query management
- **Components**:
  - `PrimaryDB`: Connection to primary PostgreSQL (leader)
  - `AnalyticsDB`: Connection to follower pool (optional)

#### 4. **Background Jobs**
- **Location**: `internal/jobs/`
- **Purpose**: Asynchronous task processing
- **Queue System**: Redis with Asynq

---

## Leader/Follower Database Pattern

### Overview

The application implements a **leader/follower (primary/replica)** database pattern to optimize read-heavy analytics queries while maintaining write performance on the primary database.

### Pattern Benefits

1. **Performance**: Offloads read-only analytics queries to follower pool
2. **Scalability**: Reduces load on primary database
3. **Availability**: Follower can serve reads if primary is temporarily unavailable
4. **Graceful Degradation**: Falls back to primary if follower is not configured

### Architecture Diagram

```mermaid
graph TB
    subgraph "Application Layer"
        App[Go Application]
        WriteHandler[Write Handlers<br/>Customers, Accounts]
        ReadHandler[Read Handlers<br/>Customers, Accounts]
        AnalyticsHandler[Analytics Handler<br/>Aggregations, Stats]
    end
    
    subgraph "Database Layer"
        Primary[(Primary Database<br/>Leader<br/>Read/Write)]
        Follower[(Follower Pool<br/>Read-Only Replica)]
    end
    
    subgraph "Replication Process"
        WALStream[PostgreSQL<br/>Streaming Replication]
    end
    
    WriteHandler -->|All Writes| Primary
    ReadHandler -->|Reads| Primary
    AnalyticsHandler -->|Read-Only Queries| Follower
    
    Primary -.->|Streams WAL| WALStream
    WALStream -.->|Applies Changes| Follower
    
    style Primary fill:#4a90e2
    style Follower fill:#50c878
    style AnalyticsHandler fill:#ffa500
```

### Key Implementation Points

#### 1. **Database Initialization**

The application initializes two separate database connections:

```go
// Primary database (required)
PrimaryDB *sql.DB  // Connected via DATABASE_URL

// Analytics database (optional)
AnalyticsDB *sql.DB  // Connected via ANALYTICS_DB_URL
```

**Initialization Flow**:
1. `InitPrimaryDB()` - Always connects to primary database
2. `InitAnalyticsDB()` - Attempts to connect to follower pool
   - If `ANALYTICS_DB_URL` is set → Connects to follower pool
   - If not set → Falls back to `PrimaryDB` reference

#### 2. **Query Routing Logic**

The application routes queries based on operation type:

```mermaid
flowchart TD
    Start[API Request] --> CheckType{Request Type?}
    
    CheckType -->|Write Operation<br/>POST, PUT, DELETE| Primary[Route to PrimaryDB]
    CheckType -->|Read Operation<br/>GET /customers, /accounts| Primary
    CheckType -->|Analytics Operation<br/>GET /analytics| CheckFollower{AnalyticsDB<br/>Configured?}
    
    CheckFollower -->|Yes| Follower[Route to AnalyticsDB<br/>Follower Pool]
    CheckFollower -->|No| Primary
    
    Primary --> Execute[Execute Query]
    Follower --> Execute
    Execute --> Response[Return Response]
    
    style Follower fill:#50c878
    style Primary fill:#4a90e2
```

#### 3. **Code Implementation**

**Analytics Handler** (`internal/api/analytics_handler.go`):
```go
func GetAnalytics(c *gin.Context) {
    // Use analytics DB (follower pool) for read-only analytics queries
    analyticsDB := db.AnalyticsDB
    if analyticsDB == nil {
        analyticsDB = db.PrimaryDB  // Fallback to primary
    }
    
    // Execute read-only queries on follower pool
    analyticsDB.QueryRow("SELECT COUNT(*) FROM customers")
    // ...
}
```

**CRUD Handlers** (`internal/api/customer_handler.go`, `account_handler.go`):
```go
func CreateCustomer(c *gin.Context) {
    // All writes go to primary database
    db.PrimaryDB.Exec("INSERT INTO customers ...")
    // ...
}
```

### Replication Details

#### PostgreSQL Streaming Replication

Heroku Postgres Advanced uses PostgreSQL's native streaming replication:

```mermaid
sequenceDiagram
    participant App as Application
    participant Primary as Primary DB (Leader)
    participant WAL as Write-Ahead Log
    participant Follower as Follower Pool
    
    App->>Primary: INSERT/UPDATE/DELETE
    Primary->>Primary: Commit Transaction
    Primary->>WAL: Write to WAL
    WAL->>Follower: Stream WAL Records
    Follower->>Follower: Apply Changes
    Note over Follower: Data is eventually consistent<br/>(typically < 1 second lag)
    
    App->>Follower: SELECT (Read-Only)
    Follower->>App: Return Results
```

**Replication Characteristics**:
- **Type**: Asynchronous streaming replication
- **Lag**: Typically < 1 second (depends on network and load)
- **Consistency**: Eventually consistent (reads may see slightly stale data)
- **Failover**: Automatic promotion to leader if primary fails (Heroku managed)

---

## Data Flow

### Write Operations Flow

```mermaid
sequenceDiagram
    participant User as User/Browser
    participant Frontend as Vue.js Frontend
    participant API as Go API Server
    participant Auth as Auth Middleware
    participant Handler as API Handler
    participant PrimaryDB as Primary Database
    
    User->>Frontend: Create Customer
    Frontend->>API: POST /api/customers<br/>{name, email}
    API->>Auth: Validate JWT Token
    Auth->>API: Token Valid
    API->>Handler: CreateCustomer()
    Handler->>PrimaryDB: INSERT INTO customers
    PrimaryDB->>Handler: Return ID
    Handler->>API: JSON Response
    API->>Frontend: 201 Created
    Frontend->>User: Show Success
```

### Read Operations Flow (CRUD)

```mermaid
sequenceDiagram
    participant User as User/Browser
    participant Frontend as Vue.js Frontend
    participant API as Go API Server
    participant Handler as API Handler
    participant PrimaryDB as Primary Database
    
    User->>Frontend: View Customers
    Frontend->>API: GET /api/customers
    API->>Handler: GetCustomers()
    Handler->>PrimaryDB: SELECT * FROM customers
    PrimaryDB->>Handler: Return Results
    Handler->>API: JSON Response
    API->>Frontend: 200 OK
    Frontend->>User: Display Customers
```

### Analytics Operations Flow (Follower Pool)

```mermaid
sequenceDiagram
    participant User as User/Browser
    participant Frontend as Vue.js Frontend
    participant API as Go API Server
    participant Handler as Analytics Handler
    participant FollowerDB as Follower Pool
    participant PrimaryDB as Primary Database
    
    User->>Frontend: View Dashboard
    Frontend->>API: GET /api/analytics
    API->>Handler: GetAnalytics()
    
    alt Follower Pool Configured
        Handler->>FollowerDB: SELECT COUNT(*) FROM customers
        Handler->>FollowerDB: SELECT COUNT(*) FROM accounts
        FollowerDB->>Handler: Return Aggregated Data
    else Follower Pool Not Configured
        Handler->>PrimaryDB: SELECT COUNT(*) FROM customers
        Handler->>PrimaryDB: SELECT COUNT(*) FROM accounts
        PrimaryDB->>Handler: Return Aggregated Data
    end
    
    Handler->>API: JSON Response
    API->>Frontend: 200 OK
    Frontend->>User: Display Analytics
```

---

## Request Routing

### API Endpoint Classification

```mermaid
graph TD
    Start[Incoming Request] --> Auth{Authentication<br/>Required?}
    
    Auth -->|No| Public[Public Endpoints]
    Auth -->|Yes| Protected[Protected Endpoints]
    
    Public --> Login[POST /api/auth/login]
    Public --> Register[POST /api/auth/register]
    Public --> Health[GET /health]
    Public --> Metrics[GET /metrics]
    Public --> Swagger[GET /swagger/*]
    
    Protected --> CRUD[CRUD Operations]
    Protected --> Analytics[Analytics Operations]
    
    CRUD --> Customers[GET/POST/PUT/DELETE<br/>/api/customers]
    CRUD --> Accounts[GET/POST/PUT/DELETE<br/>/api/accounts]
    
    Analytics --> AnalyticsOverview[GET /api/analytics]
    Analytics --> CustomerAnalytics[GET /api/analytics/customers/:id]
    
    Customers --> PrimaryDB[(Primary DB)]
    Accounts --> PrimaryDB
    AnalyticsOverview --> FollowerDB[(Follower Pool)]
    CustomerAnalytics --> FollowerDB
    
    style FollowerDB fill:#50c878
    style PrimaryDB fill:#4a90e2
```

### Database Selection Matrix

| Endpoint | Method | Database | Reason |
|----------|--------|----------|--------|
| `/api/auth/login` | POST | Primary | Write operation (session tracking) |
| `/api/auth/register` | POST | Primary | Write operation (user creation) |
| `/api/customers` | GET | Primary | Read operation (transactional data) |
| `/api/customers` | POST | Primary | Write operation |
| `/api/customers/:id` | GET | Primary | Read operation (transactional data) |
| `/api/customers/:id` | PUT | Primary | Write operation |
| `/api/customers/:id` | DELETE | Primary | Write operation |
| `/api/accounts` | GET | Primary | Read operation (transactional data) |
| `/api/accounts` | POST | Primary | Write operation |
| `/api/accounts/:id` | GET | Primary | Read operation (transactional data) |
| `/api/accounts/:id` | PUT | Primary | Write operation |
| `/api/accounts/:id` | DELETE | Primary | Write operation |
| `/api/analytics` | GET | **Follower** | Read-only aggregation (can tolerate slight lag) |
| `/api/analytics/customers/:id` | GET | **Follower** | Read-only aggregation (can tolerate slight lag) |

**Key Decision Points**:
- **Transactional Reads** (customers, accounts) → Primary DB (need latest data)
- **Analytics Reads** (aggregations, counts) → Follower Pool (can tolerate replication lag)
- **All Writes** → Primary DB (only leader accepts writes)

---

## Database Connection Management

### Connection Initialization

```mermaid
flowchart TD
    Start[Application Startup] --> LoadEnv[Load Environment Variables]
    LoadEnv --> InitJWT[Initialize JWT]
    InitJWT --> InitPrimary[InitPrimaryDB]
    
    InitPrimary --> CheckDATABASE{DATABASE_URL<br/>Set?}
    CheckDATABASE -->|No| Error1[Fatal Error:<br/>DATABASE_URL required]
    CheckDATABASE -->|Yes| ConnectPrimary[Connect to Primary DB]
    ConnectPrimary --> PingPrimary[Ping Primary DB]
    PingPrimary --> Success1[Primary DB Connected]
    
    Success1 --> InitAnalytics[InitAnalyticsDB]
    InitAnalytics --> CheckANALYTICS{ANALYTICS_DB_URL<br/>Set?}
    
    CheckANALYTICS -->|Yes| ConnectFollower[Connect to Follower Pool]
    ConnectFollower --> PingFollower[Ping Follower DB]
    PingFollower --> Success2[Follower Pool Connected]
    PingFollower -->|Error| Warn1[Warning: Follower connection failed<br/>Fallback to Primary]
    
    CheckANALYTICS -->|No| Fallback[Set AnalyticsDB = PrimaryDB]
    Warn1 --> Fallback
    Fallback --> Log[Log: Using Primary for Analytics]
    
    Success2 --> CreateTables[Create Database Tables]
    Log --> CreateTables
    CreateTables --> SeedData{SEED_DATA<br/>= true?}
    SeedData -->|Yes| Seed[Seed Sample Data]
    SeedData -->|No| StartServer
    Seed --> StartServer[Start HTTP Server]
    
    style Success2 fill:#50c878
    style Success1 fill:#4a90e2
    style Error1 fill:#ff6b6b
    style Warn1 fill:#ffa500
```

### Connection Pool Management

The application uses Go's `database/sql` package which provides built-in connection pooling:

- **Connection Pool**: Managed automatically by `sql.DB`
- **Max Open Connections**: Default (unlimited, but PostgreSQL has limits)
- **Max Idle Connections**: Default (2)
- **Connection Lifetime**: Managed by PostgreSQL server settings

**Best Practices**:
- Connections are reused across requests
- Idle connections are kept alive for quick reuse
- Connections are automatically closed when the application shuts down

---

## Background Jobs

### Architecture

```mermaid
graph TB
    subgraph "Application"
        API[API Handlers]
        JobClient[Asynq Client]
        JobServer[Asynq Server]
        JobHandler[Job Handlers]
    end
    
    subgraph "Redis"
        Queue[Job Queue]
        CriticalQ[Critical Queue]
        DefaultQ[Default Queue]
        LowQ[Low Priority Queue]
    end
    
    API -->|Enqueue Job| JobClient
    JobClient --> Queue
    Queue --> CriticalQ
    Queue --> DefaultQ
    Queue --> LowQ
    
    JobServer -->|Poll Queue| Queue
    JobServer --> JobHandler
    JobHandler --> PrimaryDB[(Primary DB)]
    
    style Queue fill:#dc3545
    style PrimaryDB fill:#4a90e2
```

### Job Processing Flow

```mermaid
sequenceDiagram
    participant API as API Handler
    participant Client as Asynq Client
    participant Redis as Redis Queue
    participant Server as Asynq Server
    participant Handler as Job Handler
    participant DB as Primary Database
    
    API->>Client: EnqueueAggregationTask()
    Client->>Redis: Push to Queue
    Redis->>Server: Job Available
    Server->>Handler: HandleAggregationTask()
    Handler->>DB: Aggregate Data
    DB->>Handler: Return Results
    Handler->>DB: Store Aggregated Results
    Handler->>Server: Job Complete
    Server->>Redis: Acknowledge Completion
```

### Queue Priorities

The application uses three priority queues:

1. **Critical Queue** (Priority 6): High-priority tasks
2. **Default Queue** (Priority 3): Standard tasks
3. **Low Queue** (Priority 1): Background processing

**Configuration** (`main.go`):
```go
asynq.Config{
    Concurrency: 10,
    Queues: map[string]int{
        "critical": 6,
        "default":  3,
        "low":      1,
    },
}
```

---

## Authentication & Authorization

### JWT Authentication Flow

```mermaid
sequenceDiagram
    participant User as User
    participant Frontend as Frontend
    participant API as API Server
    participant Auth as Auth Module
    participant DB as Primary Database
    
    User->>Frontend: Enter Credentials
    Frontend->>API: POST /api/auth/login<br/>{username, password}
    API->>DB: SELECT password_hash FROM users
    DB->>API: Return Hash
    API->>Auth: CheckPasswordHash()
    Auth->>API: Password Valid
    API->>Auth: GenerateToken()
    Auth->>API: JWT Token
    API->>Frontend: {token: "..."}
    Frontend->>Frontend: Store Token (localStorage)
    
    Note over Frontend,API: Subsequent Requests
    
    Frontend->>API: GET /api/customers<br/>Authorization: Bearer {token}
    API->>Auth: ValidateToken()
    Auth->>API: Token Valid
    API->>DB: SELECT * FROM customers
    DB->>API: Return Data
    API->>Frontend: 200 OK
```

### Middleware Protection

```mermaid
flowchart TD
    Request[Incoming Request] --> CheckPath{Path?}
    
    CheckPath -->|/api/auth/*| Allow[Allow Request]
    CheckPath -->|/health, /metrics| Allow
    CheckPath -->|/swagger/*| Allow
    CheckPath -->|/api/*| CheckAuth{Authorization<br/>Header?}
    
    CheckAuth -->|No| Reject1[401 Unauthorized]
    CheckAuth -->|Yes| ExtractToken[Extract Bearer Token]
    ExtractToken --> ValidateToken{Token<br/>Valid?}
    
    ValidateToken -->|No| Reject2[401 Unauthorized]
    ValidateToken -->|Yes| ExtractClaims[Extract Username]
    ExtractClaims --> Allow2[Allow Request<br/>Continue to Handler]
    
    style Reject1 fill:#ff6b6b
    style Reject2 fill:#ff6b6b
    style Allow fill:#50c878
    style Allow2 fill:#50c878
```

---

## Deployment Architecture

### Heroku Deployment

```mermaid
graph TB
    subgraph "Heroku Platform"
        subgraph "Dyno"
            App[Go Application]
        end
        
        subgraph "Addons"
            PrimaryAddon[Heroku Postgres<br/>Primary Database]
            FollowerAddon[Heroku Postgres<br/>Follower Pool]
            RedisAddon[Heroku Redis<br/>Optional]
        end
    end
    
    subgraph "Build Process"
        Buildpack1[Node.js Buildpack<br/>Build Frontend]
        Buildpack2[Go Buildpack<br/>Build Backend]
    end
    
    subgraph "External"
        Users[Users/Browsers]
        CDN[Static Assets CDN]
    end
    
    Users --> App
    App --> PrimaryAddon
    App --> FollowerAddon
    App --> RedisAddon
    App --> CDN
    
    Buildpack1 --> App
    Buildpack2 --> App
    
    PrimaryAddon -.->|Replication| FollowerAddon
    
    style FollowerAddon fill:#50c878
    style PrimaryAddon fill:#4a90e2
    style RedisAddon fill:#dc3545
```

### Environment Configuration

```mermaid
graph LR
    subgraph "Environment Variables"
        DATABASE_URL[DATABASE_URL<br/>Auto-set by Heroku]
        ANALYTICS_DB_URL[ANALYTICS_DB_URL<br/>Manually Set]
        REDIS_URL[REDIS_URL<br/>Auto-set if provisioned]
        JWT_SECRET[JWT_SECRET<br/>Manually Set]
        PORT[PORT<br/>Auto-set by Heroku]
        SEED_DATA[SEED_DATA<br/>Optional]
    end
    
    subgraph "Application"
        Config[Configuration]
    end
    
    DATABASE_URL --> Config
    ANALYTICS_DB_URL --> Config
    REDIS_URL --> Config
    JWT_SECRET --> Config
    PORT --> Config
    SEED_DATA --> Config
    
    style ANALYTICS_DB_URL fill:#ffa500
    style JWT_SECRET fill:#ffa500
```

---

## Key Architectural Decisions

### 1. **Why Leader/Follower for Analytics?**

**Problem**: Analytics queries (COUNT, AVG, aggregations) can be resource-intensive and slow down transactional operations.

**Solution**: Route analytics queries to a read-only follower pool.

**Benefits**:
- Primary database remains responsive for transactional operations
- Analytics queries don't block writes
- Can scale analytics independently
- Follower pool can be optimized for read-heavy workloads

**Trade-offs**:
- Slight replication lag (typically < 1 second)
- Additional infrastructure cost
- More complex connection management

### 2. **Why Not Use Follower for All Reads?**

**Decision**: Only use follower pool for analytics, not for transactional reads.

**Reasoning**:
- Transactional reads (GET /api/customers/:id) need the latest data
- Users expect to see their changes immediately
- Analytics can tolerate slight lag (counts, averages)
- Simpler mental model: "writes and transactional reads → primary, analytics → follower"

### 3. **Graceful Degradation**

**Implementation**: If `ANALYTICS_DB_URL` is not set, analytics endpoints use the primary database.

**Benefits**:
- Application works without follower pool (development, small deployments)
- No code changes needed to enable/disable follower pool
- Easy to test locally with single database

### 4. **Connection Management**

**Decision**: Use separate connection pools for primary and follower.

**Benefits**:
- Independent connection limits
- Can tune each pool separately
- Clear separation of concerns
- Easy to monitor connection usage per database

---

## Monitoring & Observability

### Health Check Endpoint

The `/health` endpoint reports the status of both database connections:

```json
{
  "status": "healthy",
  "database": "connected",
  "analytics_db": "connected"  // or "using primary" if not configured
}
```

### Metrics Endpoint

The `/metrics` endpoint provides Prometheus-compatible metrics for:
- HTTP request counts
- Response times
- Error rates
- Database connection pool stats (if instrumented)

---

## Future Enhancements

### Potential Improvements

1. **Connection Pooling Configuration**
   - Make connection pool sizes configurable
   - Add metrics for connection pool usage

2. **Read Replicas for Transactional Reads**
   - Add option to route some transactional reads to follower
   - Implement read preference based on consistency requirements

3. **Caching Layer**
   - Add Redis caching for frequently accessed data
   - Cache analytics results with TTL

4. **Database Sharding**
   - Partition customers/accounts by region or ID range
   - Route queries to appropriate shard

5. **Event Sourcing**
   - Store events instead of current state
   - Rebuild analytics from event stream

---

## Conclusion

The SaaS Go App implements a clean separation between transactional operations and analytics queries through the leader/follower database pattern. This architecture provides:

- **Performance**: Optimized read and write operations
- **Scalability**: Can scale reads independently from writes
- **Reliability**: Graceful degradation if follower pool is unavailable
- **Simplicity**: Clear routing logic and easy to understand

The application is designed to work with or without a follower pool, making it suitable for both development and production environments.


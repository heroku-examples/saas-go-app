# Next Generation Postgres (NGPG) Setup Guide

This guide explains how to use Heroku Postgres Advanced (Next Generation) with automatic read/write routing.

## Quick Answer

**No code changes needed!** Just configure your environment variables:

1. Set `DATABASE_URL` to your NGPG connection string (with automatic routing)
2. Don't set `ANALYTICS_DB_URL` (or unset it if already set)

The database will automatically route writes to the leader and reads to the follower pool.

## Configuration Options

### Option 1: Automatic Routing (Recommended for NGPG)

**Configuration:**
```bash
# Set DATABASE_URL to your NGPG connection string
heroku config:set DATABASE_URL="postgres://..." --app your-app

# Don't set ANALYTICS_DB_URL (or unset it)
heroku config:unset ANALYTICS_DB_URL --app your-app
```

**How it works:**
- All queries (writes and reads) use `DATABASE_URL`
- The database automatically routes:
  - Write operations (INSERT, UPDATE, DELETE) → Leader
  - Read operations (SELECT) → Follower pool (if configured)
- No application code changes needed
- Simplest configuration

**When to use:**
- Using Heroku Postgres Advanced (Next Generation)
- Follower pool is configured in Heroku Dashboard
- You want the database to handle routing automatically

### Option 2: Explicit Routing (Current Implementation)

**Configuration:**
```bash
# Set DATABASE_URL to leader connection
heroku config:set DATABASE_URL="postgres://leader..." --app your-app

# Set ANALYTICS_DB_URL to follower pool connection
heroku config:set ANALYTICS_DB_URL="postgres://follower..." --app your-app
```

**How it works:**
- Writes and transactional reads use `DATABASE_URL` (leader)
- Analytics reads use `ANALYTICS_DB_URL` (follower pool)
- Application code explicitly chooses which database to query
- More control over query routing

**When to use:**
- You want explicit control over which queries go where
- Using legacy Postgres tiers (Standard, Premium)
- You need fine-grained routing logic

## Current Application Behavior

The application is designed to work with both approaches:

1. **If `ANALYTICS_DB_URL` is set:**
   - Analytics queries use the explicit follower pool connection
   - Writes and transactional reads use `DATABASE_URL`

2. **If `ANALYTICS_DB_URL` is NOT set:**
   - Analytics queries use `DATABASE_URL` (same as primary)
   - With NGPG automatic routing, the database routes reads to followers automatically
   - No code changes needed!

## Verification

After setting up, check your configuration:

```bash
# Check environment variables
heroku config --app your-app

# Verify database connections
heroku run "go run main.go" --app your-app
# Look for log messages:
# - "Primary database connection established"
# - "ANALYTICS_DB_URL not set, analytics endpoints will use primary DB connection"
# - "With Heroku Postgres Advanced, if DATABASE_URL has automatic routing configured, read queries will be automatically routed to the follower pool."
```

## References

- [Heroku Postgres Advanced Announcement](https://www.heroku.com/blog/introducing-the-next-generation-of-heroku-postgres/)
- [Postgres Performance Guide](https://devcenter.heroku.com/articles/getting-started-postgres-performance)
- [Architecture Documentation](ARCHITECTURE.md)


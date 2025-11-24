# Heroku Deployment Fix

## Issue
The app was crashing with `command not found` because Heroku's Go buildpack couldn't find the binary.

## Solution
Heroku's Go buildpack automatically detects Go apps, but when `main.go` is in a subdirectory (`cmd/server/`), we need to help it find the main package.

## Fixed Files

1. **Procfile** - Updated to reference the correct binary name
2. **compile** - Added compile script to tell Heroku where the main package is
3. **Database name** - Changed from `SAAS_GO_DB` to `saas-go-db` (no underscores allowed)

## Deployment Steps

1. **Fix the database name** (if you already created it with underscores):
```bash
# If you need to recreate with correct name
heroku addons:destroy heroku-postgresql --app saas-go-app --confirm saas-go-app
heroku addons:create heroku-postgresql:standard-0 --name saas-go-db --app saas-go-app
```

2. **Commit and push the fixes**:
```bash
git add .
git commit -m "Fix Heroku deployment: Procfile and database name"
git push heroku main
```

3. **Verify the deployment**:
```bash
heroku logs --tail --app saas-go-app
heroku open --app saas-go-app
```

## How It Works

The `compile` script sets environment variables that Heroku's Go buildpack uses to locate the main package. The buildpack will:
1. Detect `go.mod` in the root
2. Read `GO_INSTALL_PACKAGE_PATH` from the compile script
3. Build the binary from `./cmd/server`
4. Name it `saas-go-app` (based on module name) and place it in `bin/`
5. The Procfile runs `bin/saas-go-app`

**Note**: If this doesn't work, you may need to move `main.go` to the root directory temporarily, or use a different deployment method.

## Alternative Solution

If the compile script doesn't work, you can also:
1. Move `main.go` to the root (not recommended, breaks project structure)
2. Use a custom buildpack
3. Use Docker deployment instead


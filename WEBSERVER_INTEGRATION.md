# Integrating Your App into Rick's Tools Webserver

This document explains how to add your Go web application to the parent webserver at:

```
/home/zditech/CharmToolWeb/cmd/webserver/main.go
```

The webserver uses Go's standard `http.ServeMux` to mount independent apps under their own URL prefix (e.g., `/charmtool/`, `/yourapp/`). Each app is a self-contained `http.Handler`. The root landing page (`/`) automatically lists all registered apps with links and descriptions.

## Architecture

- The webserver handles TLS directly using Let's Encrypt certificates (no nginx)
- When `TLS_CERT` and `TLS_KEY` env vars are set, it serves HTTPS on port 443 and redirects HTTP port 80 to HTTPS
- Without those vars, it serves plain HTTP on `PORT` (default 8080) for local development
- The landing page at `/` is auto-generated from the app registry
- Deployment is handled by `deploy.sh` which builds, copies, and installs a systemd service

## What You Need to Provide

Your app must expose a function that returns an `http.Handler`. This handler registers all of its own routes (API endpoints, static files, etc.) internally. Routes must be root-relative (e.g., `/api/...`, `/`) because the parent server uses `http.StripPrefix` to remove the mount prefix before passing requests to your handler.

### Example: NewApp Function

```go
package yourapp

import "net/http"

// NewApp returns an http.Handler with all routes for this app.
// cookiePath should be set to the mount prefix (e.g., "/yourapp/") so
// session cookies are scoped correctly.
func NewApp(cookiePath string) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("/api/data", handleData)
    mux.Handle("/", http.FileServer(http.Dir("./web/yourapp")))

    return mux
}
```

## How to Register Your App

Edit `/home/zditech/CharmToolWeb/cmd/webserver/main.go`. There are two steps:

### Step 1: Mount Your Handler

Import your package, create the handler, and mount it with `http.StripPrefix`:

```go
import "yourmodule/internal/yourapp"

// Mount your app at /yourapp/
yourAppHandler := yourapp.NewApp("/yourapp/")
mux.Handle("/yourapp/", http.StripPrefix("/yourapp", yourAppHandler))
```

### Step 2: Register in the App List

Add your app to the `apps` slice so it appears on the landing page with a link and description:

```go
apps = append(apps, App{
    Name:        "Your App",
    Path:        "/yourapp/",
    Description: "One line description of what your app does",
})
```

### Full Context

Here is the relevant section of `cmd/webserver/main.go` showing where to add your code:

```go
mux := http.NewServeMux()
apps := []App{}

// --- CharmToolWeb mounted at /charmtool/ ---
charmtoolApp := handlers.NewApp(store, staticDir, "/charmtool/")
mux.Handle("/charmtool/", http.StripPrefix("/charmtool", charmtoolApp))
apps = append(apps, App{
    Name:        "CharmTool",
    Path:        "/charmtool/",
    Description: "Convert KiCad POS files to Charmhigh pick-and-place DPV format",
})

// --- Add your app here ---
// yourHandler := yourapp.NewApp("/yourapp/")
// mux.Handle("/yourapp/", http.StripPrefix("/yourapp", yourHandler))
// apps = append(apps, App{
//     Name:        "Your App",
//     Path:        "/yourapp/",
//     Description: "One line description of your app",
// })
```

## Requirements

### Frontend: Use Relative URLs

All fetch/API calls in your HTML/JS must use **relative** URLs (no leading `/`):

```javascript
// CORRECT - works under any mount prefix
fetch('api/data')
fetch(`api/item?id=${id}`)

// WRONG - breaks when mounted under a prefix
fetch('/api/data')
```

When the browser is on `https://host/yourapp/` and you call `fetch('api/data')`, it resolves to `https://host/yourapp/api/data`. The server strips `/yourapp` and your handler sees `/api/data`.

### Session Cookies: Scope to Your Prefix

If your app uses cookies, set the cookie `Path` to your mount prefix so cookies don't leak between apps:

```go
http.SetCookie(w, &http.Cookie{
    Name:   "yourapp_session",
    Value:  sessionID,
    Path:   cookiePath,  // "/yourapp/" - not "/"
})
```

### Go Module

Your app must be importable by the webserver. Options:

- **Same module**: Put your code under a new directory in the CharmToolWeb repo (e.g., `internal/yourapp/`). It's already importable as `charmtool/internal/yourapp`.
- **Separate module**: Add it as a dependency in `/home/zditech/CharmToolWeb/go.mod`:
  ```
  go get yourmodule@latest
  ```
  Or for a local module during development, add a replace directive:
  ```
  // In go.mod
  replace yourmodule => ../YourAppDir
  ```

### Static Files

Put your static files in a directory your handler can reference. To keep apps separate, use a dedicated subdirectory like `web/yourapp/`. Make sure `deploy.sh` copies your static files to the server (see Deployment below).

## Deployment

The project includes `deploy.sh` which builds a Linux binary, copies it and static files to the remote server, and manages a systemd service.

### Updating deploy.sh for Your Static Files

If your app has its own static files, add copy lines to `deploy.sh`:

```bash
# In the "Creating deployment bundle" section, add:
mkdir -p "${STAGING}/web/yourapp"
cp -r web/yourapp/*  "${STAGING}/web/yourapp/"

# In the "Copying files" section, add:
scp -r "${STAGING}/web/yourapp/"  "${REMOTE}:${REMOTE_DIR}/web/"
```

### Deploy Commands

```bash
cd /home/zditech/CharmToolWeb

# Deploy with TLS (production)
./deploy.sh root@srv629042.hstgr.cloud \
  /etc/letsencrypt/live/richymac.com/fullchain.pem \
  /etc/letsencrypt/live/richymac.com/privkey.pem

# Deploy without TLS (plain HTTP on port 8080)
./deploy.sh root@srv629042.hstgr.cloud
```

### Server Management

```bash
ssh root@srv629042.hstgr.cloud journalctl -u charmtoolweb -f    # view logs
ssh root@srv629042.hstgr.cloud systemctl restart charmtoolweb    # restart
ssh root@srv629042.hstgr.cloud systemctl stop charmtoolweb       # stop
```

## Checklist

- [ ] Your app exposes a function returning `http.Handler`
- [ ] All internal routes use root-relative paths (`/api/...`, `/`)
- [ ] All frontend fetch calls use relative URLs (`api/...`, not `/api/...`)
- [ ] Session cookies use the mount prefix as their `Path`
- [ ] Your package is importable by `cmd/webserver/main.go`
- [ ] You added the `mux.Handle` + `http.StripPrefix` lines to `cmd/webserver/main.go`
- [ ] You added an `apps = append(apps, App{...})` entry for the landing page
- [ ] You updated `deploy.sh` to copy your static files if needed
- [ ] You redeployed with `./deploy.sh`

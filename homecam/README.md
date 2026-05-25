# Sentinel NOC - Security Camera Network Operations Center

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-61DAFB?style=for-the-badge&logo=react&logoColor=black" alt="React">
  <img src="https://img.shields.io/badge/MongoDB-47A248?style=for-the-badge&logo=mongodb&logoColor=white" alt="MongoDB">
  <img src="https://img.shields.io/badge/Kubernetes-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white" alt="Kubernetes">
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
</p>

A comprehensive Security Camera Network Operations Center (NOC) / Video Management System built with **Go (Gin)**, **React 19**, and **MongoDB**. Features a **zero-trust security model** with encrypted credentials, JWT + TOTP 2FA authentication, full RBAC, ONVIF camera integration, and real RTSP→HLS live streaming via MediaMTX.

---

## Architecture

```
                        ┌──────────────────────────┐
                        │      Client Browser       │
                        └────────────┬─────────────┘
                                     │ HTTP / HLS / WebSocket
                        ┌────────────▼─────────────┐
                        │  Nginx Ingress Controller  │
                        │  (port 80 / 443)           │
                        └──┬──────────┬──────────┬──┘
                           │ /        │ /api     │ /hls
              ┌────────────▼──┐  ┌────▼──────┐  ┌▼──────────────┐
              │   Frontend    │  │  Backend  │  │   MediaMTX    │
              │  React 19     │  │  Go/Gin   │  │  RTSP→HLS     │
              │  Nginx :80    │  │  :8001    │  │  :8888 / :8554│
              └───────────────┘  └────┬──────┘  └───────────────┘
                                      │ motor driver
                              ┌───────▼───────┐
                              │   MongoDB 6   │
                              │   :27017      │
                              └───────────────┘
```

**Kubernetes namespace**: `sentinel-noc`  
**Cluster tooling**: k3d (k3s-in-Docker)  
**Ingress**: ingress-nginx

---

## Features

### Security & Authentication
- **JWT Tokens** — 30-min access tokens + 7-day refresh tokens
- **TOTP 2FA** — Google Authenticator / Authy support
- **bcrypt** password hashing with strength validation
- **Rate limiting** — 100 req/min per IP (in-memory)
- **Account lockout** — 15-min lockout after 5 failed login attempts
- **AES-256-GCM encryption** for all camera credentials at rest
- **Comprehensive audit logging** — every security action logged with IP/user

### RBAC (Role-Based Access Control)
| Feature | Admin | Security Operator | Viewer |
|---------|-------|-------------------|--------|
| Live viewing | ✅ | ✅ | ✅ |
| PTZ control | ✅ | ✅ | ❌ |
| Add/edit cameras | ✅ | ✅ | ❌ |
| Delete cameras | ✅ | ❌ | ❌ |
| Alarm management | ✅ | ✅ | ❌ |
| User management | ✅ | ❌ | ❌ |
| Audit logs | ✅ | ❌ | ❌ |
| System settings | ✅ | ❌ | ❌ |

### Camera Management
- CRUD with encrypted credential storage
- ONVIF device capability detection (PTZ, relay outputs, audio)
- RTSP → HLS live streaming via MediaMTX
- Motion detection flag, recording enable/disable
- Online/offline status tracking
- System mode overrides (home/away/none)

### Streaming (RTSP → HLS)
- **MediaMTX** media server handles RTSP pull and HLS packaging
- Low-latency HLS: 1-second segments, 200ms parts
- Backend manages stream lifecycle via MediaMTX REST API
- Streams accessible at `http://<host>/hls/cam-<id>/index.m3u8`

### Event System
- Motion, alarm, and connection event types
- Severity levels: info, warning, critical
- Acknowledgment workflow
- Filtering by camera, type, severity, date

---

## Quick Start

### Option A — Docker Compose (Development)

```bash
# Clone the repository
git clone <repository-url>
cd sentinel-noc

# Start all services
make up

# Access
# Frontend:  http://localhost:3000
# Backend:   http://localhost:8001/api
# HLS:       http://localhost:8888
# Login:     admin / P@ssw0rd!
```

### Option B — Kubernetes via k3d (Recommended for Production-like)

```bash
# Prerequisites: Docker, k3d, kubectl
# k3d install: https://k3d.io/#installation

# One-command full setup
make k8s-up

# Access
# Frontend: http://localhost   (via ingress on port 80)
# Backend:  http://localhost/api
# HLS:      http://localhost:8888
# RTSP:     rtsp://localhost:8554
```

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Docker | 24+ | Container runtime |
| docker-compose | 2+ | Local development |
| k3d | 5+ | Local Kubernetes cluster |
| kubectl | 1.28+ | Kubernetes CLI |
| make | any | Task runner |

Install k3d:
```bash
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

---

## Kubernetes Deployment

### Cluster Management

```bash
make k3d-create      # Create k3d cluster + install ingress-nginx
make k3d-delete      # Delete cluster (destroys all data)
make k3d-start       # Start a stopped cluster
make k3d-stop        # Suspend cluster without deleting it
make k3d-list        # List all k3d clusters
```

### Full Lifecycle

```bash
# Stand up everything (cluster + images + deploy)
make k8s-up

# Tear down everything (undeploy + delete cluster)
make k8s-down
```

### Image Management

```bash
make k8s-build   # Build backend + frontend Docker images
make k8s-push    # Import local images into the k3d cluster registry
```

### Deployment

```bash
make k8s-deploy       # Apply all k8s manifests
make k8s-undeploy     # Remove all resources (keeps namespace + secrets)
make k8s-redeploy     # Rebuild images, import, and rolling-restart
make k8s-rollout      # Rolling restart of all deployments (no rebuild)
make k8s-status       # Show pod, service, and ingress status
make k8s-health       # Run an in-cluster health check
```

### Secrets (Production)

The `k8s/secrets.yaml` file contains **template placeholder values**. Before deploying to a real environment, generate secure secrets:

```bash
make k8s-secrets      # Generate + apply cryptographically secure secrets
```

Or apply manually:
```bash
kubectl create secret generic sentinel-secrets \
  --namespace sentinel-noc \
  --from-literal=jwt-secret="$(openssl rand -base64 48)" \
  --from-literal=encryption-key="$(openssl rand -base64 32 | head -c 44)=" \
  --from-literal=mongo-username="admin" \
  --from-literal=mongo-password="$(openssl rand -base64 24)"
```

### Accessing the Application in Kubernetes

**Via Ingress (default, port 80):**
```
http://localhost        → Frontend
http://localhost/api    → Backend API
http://localhost/hls    → HLS stream segments
```

**Via Port-Forward (direct pod access):**
```bash
make k8s-port-forward
# Frontend → http://localhost:3000
# Backend  → http://localhost:8001
# HLS      → http://localhost:8888
```

### Scaling

```bash
make k8s-scale-backend REPLICAS=3
make k8s-scale-frontend REPLICAS=3
```

The backend also has a HorizontalPodAutoscaler configured (2–5 replicas, 70% CPU target).

### Logs & Debugging

```bash
make k8s-logs               # All pods (streamed)
make k8s-logs-backend       # Backend pods only
make k8s-logs-frontend      # Frontend pods only
make k8s-logs-db            # MongoDB pod
make k8s-logs-mediamtx      # MediaMTX streaming server

make k8s-shell-backend      # Shell into a backend pod
make k8s-shell-frontend     # Shell into a frontend pod
make k8s-shell-db           # mongosh into the MongoDB pod
```

---

## Kubernetes Resource Overview

```
k8s/
├── k3d-config.yaml              # k3d cluster definition
├── namespace.yaml               # sentinel-noc namespace
├── configmap.yaml               # Non-sensitive configuration
├── secrets.yaml                 # Secret template (replace before prod)
├── ingress.yaml                 # Nginx ingress routes
├── network-policy.yaml          # Network isolation + camera egress
├── mongodb/
│   ├── statefulset.yaml         # MongoDB StatefulSet (10Gi PVC)
│   └── service.yaml             # mongodb-service ClusterIP :27017
├── backend/
│   ├── deployment.yaml          # Backend Deployment (2 replicas)
│   ├── service.yaml             # backend-service ClusterIP :8001
│   └── hpa.yaml                 # HPA: 2–5 replicas, 70% CPU
├── frontend/
│   ├── deployment.yaml          # Frontend Deployment (2 replicas)
│   └── service.yaml             # frontend-service ClusterIP :80
├── mediamtx/
│   ├── configmap.yaml           # mediamtx.yml configuration
│   ├── deployment.yaml          # MediaMTX Deployment
│   └── service.yaml             # mediamtx-service ClusterIP :8554/:8888/:9997
└── ingress-nginx/
    └── install.sh               # ingress-nginx installation script
```

### Network Policy

The network policies enforce:
- **Intra-namespace**: all pods can talk to each other freely
- **Ingress-nginx**: ingress controller can reach all pods
- **Backend egress**: unrestricted (required for ONVIF/RTSP camera access on the LAN)
- **MediaMTX egress**: unrestricted (required to pull RTSP streams from cameras)
- **DNS**: all pods can resolve DNS

### Host Network Access for Cameras

IP cameras typically reside on the same LAN as the host machine. k3d nodes run as Docker containers on the host, so they have access to the host network via Docker bridge routing. No additional configuration is needed for most setups.

If cameras are only reachable from the host (not from Docker bridge), add this to the backend Deployment:
```yaml
spec:
  template:
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
```

---

## Docker Compose (Development)

```bash
make up               # Start all services (detached)
make dev              # Start with hot-reload (air for Go, CRA for React)
make down             # Stop all services
make restart          # Restart all services
make status           # Show container status
make health           # Check service health endpoints

make logs             # Follow all logs
make logs-backend     # Backend logs only
make logs-frontend    # Frontend logs only
make logs-db          # MongoDB logs only
make logs-mediamtx    # MediaMTX logs only

make shell-backend    # sh into backend container
make shell-frontend   # sh into frontend container
make shell-db         # mongosh into MongoDB container

make test             # Run all tests
make test-backend     # Go tests (go test ./...)
make test-frontend    # React tests (yarn test)

make lint             # Lint all code
make lint-backend     # go vet ./...
make lint-frontend    # yarn lint

make backup-db        # Dump MongoDB to ./backups/
make restore-db BACKUP=./backups/backup-xxx  # Restore from dump
make reset-db         # Drop and re-seed database (DESTRUCTIVE)

make clean            # Remove containers + volumes
make clean-images     # Remove Docker images
make prune            # docker system prune (all unused resources)
```

---

## Configuration

### Environment Variables (Backend)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8001` | API server port |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `MONGO_URL` | `mongodb://localhost:27017` | MongoDB connection string |
| `DB_NAME` | `sentinel_noc` | MongoDB database name |
| `JWT_SECRET` | *(required)* | JWT signing key (min 32 chars) |
| `ENCRYPTION_KEY` | *(required)* | Fernet key for credential encryption |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | `30` | JWT access token lifetime |
| `REFRESH_TOKEN_EXPIRE_DAYS` | `7` | JWT refresh token lifetime |
| `RATE_LIMIT_MAX` | `100` | Max requests per IP per window |
| `RATE_LIMIT_WINDOW` | `60` | Rate limit window in seconds |
| `CORS_ORIGINS` | `http://localhost:3000` | Comma-separated allowed origins |
| `MEDIAMTX_URL` | `http://localhost:9997` | MediaMTX REST API URL |

### Environment Variables (Frontend)

| Variable | Default | Description |
|----------|---------|-------------|
| `REACT_APP_BACKEND_URL` | `http://localhost:8001` | Backend API base URL |

---

## API Endpoints

All protected endpoints require `Authorization: Bearer <access_token>`.

### Authentication
```
POST /api/auth/register          Register new user
POST /api/auth/login             Login (returns access + refresh tokens)
POST /api/auth/refresh           Refresh access token
GET  /api/auth/me                Get current user profile
POST /api/auth/2fa/setup         Initialize TOTP 2FA
POST /api/auth/2fa/verify        Verify and enable TOTP
POST /api/auth/2fa/disable       Disable TOTP [admin]
```

### Cameras
```
GET    /api/cameras              List all cameras
POST   /api/cameras              Create camera [admin, security_operator]
GET    /api/cameras/:id          Get camera details
PUT    /api/cameras/:id          Update camera [admin, security_operator]
DELETE /api/cameras/:id          Delete camera [admin]
GET    /api/cameras/:id/stream-url        Get RTSP stream URL
POST   /api/cameras/:id/stream/start      Start HLS stream via MediaMTX
POST   /api/cameras/:id/stream/stop       Stop HLS stream [admin, operator]
GET    /api/cameras/:id/stream/status     Get stream status
POST   /api/cameras/:id/status   Update online/offline status
```

### ONVIF
```
POST /api/cameras/:id/onvif/detect        Detect ONVIF capabilities
POST /api/cameras/:id/onvif/test          Test ONVIF connection
POST /api/cameras/:id/onvif/credentials   Update ONVIF credentials [admin, operator]
POST /api/cameras/:id/alarm/trigger       Trigger relay alarm [admin, operator]
POST /api/cameras/:id/alarm/stop          Stop relay alarm [admin, operator]
```

### Events
```
GET  /api/events                  List events (filterable)
POST /api/events                  Create event
POST /api/events/:id/acknowledge  Acknowledge event [admin, operator]
```

### System
```
GET /api/system/mode              Get current system mode (home/away)
PUT /api/system/mode              Set system mode [admin, operator]
GET /api/dashboard/stats          Dashboard statistics
```

### Admin Only
```
GET    /api/users                 List users
PUT    /api/users/:id/role        Update user role
DELETE /api/users/:id             Delete user
GET    /api/settings              Get system settings
PUT    /api/settings              Update system settings
GET    /api/audit-logs            View audit log (filterable)
```

---

## Project Structure

```
sentinel-noc/
├── Makefile                  # All make targets (Docker + Kubernetes)
├── docker-compose.yml        # Base Docker Compose config
├── docker-compose.dev.yml    # Dev overrides (hot-reload, debug)
├── docker-compose.prod.yml   # Production overrides (auth, no mounts)
├── mediamtx.yml              # MediaMTX streaming server config
├── k8s/                      # Kubernetes manifests
│   ├── k3d-config.yaml
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secrets.yaml
│   ├── ingress.yaml
│   ├── network-policy.yaml
│   ├── mongodb/
│   ├── backend/
│   ├── frontend/
│   ├── mediamtx/
│   └── ingress-nginx/
├── backend/                  # Go/Gin API server
│   ├── cmd/api/main.go       # Entry point, router, DI
│   ├── internal/
│   │   ├── config/           # Environment configuration
│   │   ├── handlers/         # HTTP handlers
│   │   ├── middleware/       # JWT auth, CORS, rate limiting, RBAC
│   │   ├── models/           # Request/response DTOs
│   │   ├── repository/       # MongoDB data access layer
│   │   └── services/         # Business logic
│   ├── Dockerfile            # Production multi-stage build
│   └── Dockerfile.dev        # Development build (air hot-reload)
└── frontend/                 # React 19 SPA
    ├── src/
    │   ├── App.js            # Router + protected routes
    │   ├── pages/            # 8 page components
    │   ├── components/       # UI components (shadcn/ui + custom)
    │   └── contexts/         # AuthContext
    ├── Dockerfile            # Development build (CRA dev server)
    ├── Dockerfile.prod       # Production build (multi-stage + Nginx)
    └── nginx.conf            # Nginx configuration for prod
```

---

## Default Credentials

| Field | Value |
|-------|-------|
| Username | `admin` |
| Password | `P@ssw0rd!` |

**Change this immediately in any non-development environment.**

---

## Security Notes

1. **Secrets**: The `k8s/secrets.yaml` contains placeholder values. Always run `make k8s-secrets` before deploying to any shared or production environment.
2. **TLS**: The Ingress definition does not include TLS. Add a `tls:` block and cert-manager (or manual certificate) for HTTPS.
3. **Network Policies**: Camera access requires egress from backend/mediamtx pods to the camera LAN. The network policies allow this explicitly.
4. **TOTP 2FA**: Admins should enable 2FA via the Settings page after first login.
5. **Encryption key rotation**: If you change `ENCRYPTION_KEY`, all stored camera credentials become unreadable. Re-enter credentials after rotation.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.21, Gin, MongoDB Motor |
| Frontend | React 19, React Router 7, Tailwind CSS, shadcn/ui |
| Database | MongoDB 6 |
| Streaming | MediaMTX (RTSP→HLS) |
| Auth | JWT (golang-jwt/jwt v5), TOTP (pquerna/otp) |
| Encryption | AES-256-GCM (crypto/aes) |
| Passwords | bcrypt (golang.org/x/crypto) |
| Containers | Docker, docker-compose |
| Orchestration | Kubernetes (k3d/k3s), ingress-nginx |


New credentials

    Username: admin
    Password: ChangeMe!2026

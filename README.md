# Home Platform

Infrastructure hub for a self-hosted home automation and AI platform. This repository owns the shared CI/CD pipelines, Traefik gateway configuration, and Ansible/AWX automation. The three application stacks are included as git submodules.

## Repository Layout

```
Local-Stuff/  (this repo)
├── awx/                    Ansible automation — AWX deployment, playbooks, roles
├── jenkins/                Shared CI/CD pipelines for all three apps
├── traefik/                Traefik v3 reverse proxy gateway
├── homecam/                submodule → github.com/fntundi/HomeCam
├── local-aistack/          submodule → github.com/fntundi/local-aistack
└── Sentinel-Home/          submodule → github.com/fntundi/Sentinel-Home
```

## Submodules

### homecam — Sentinel NOC Camera Platform

Security camera network-operations-center built on Go and React.

- **Backend**: Go 1.21 / Gin — REST API, JWT + TOTP 2FA, MongoDB, RTSP stream management
- **Frontend**: React 19 / craco / Tailwind CSS — live camera grid, role-based access control
- **Streaming**: MediaMTX — RTSP ingestion → HLS output, ONVIF camera discovery
- **Deployment**: Docker Compose (dev/prod overlays) or k3d cluster `sentinel-noc`
- **Namespace**: `sentinel-noc` | **Ports**: backend `:8001`, HLS `:8888`, RTSP `:8554`
- **Key make targets**: `make up`, `make dev`, `make k3d-create`, `make k8s-deploy`

### local-aistack — Hybrid AI Platform

Production-aligned self-hosted AI stack optimised for CPU-only home servers.

- **LLM inference**: dual Ollama instances (`ollama-chat`, `ollama-dev`) — phi3, nomic-embed-text
- **Gateway / proxy**: LiteLLM `:4000` — provider-agnostic model routing and rate limiting
- **UI**: OpenWebUI `:3000` — full-featured chat interface
- **Notebooks**: JupyterLab `:8888` with scipy stack and platform data mounts
- **Experiment tracking**: MLflow `:5000` backed by PostgreSQL + MinIO (S3-compatible)
- **Observability**: Langfuse `:3100` (LLM traces), OpenSearch Dashboards `:5601`
- **Automation**: Activepieces `:8080` — no-code workflow automation
- **Vector search**: Qdrant `:6333`
- **Custom services**: api-server, code-executor (sandboxed), workflows-runner, general-mcp
- **Infra**: PostgreSQL 16, Redis 7, MinIO — all on `ai-net` (172.22.0.0/16)
- **Deployment**: Docker Compose prod overlay or k3d cluster `ai-platform`
- **Key make targets**: `make up`, `make health`, `make models`, `make backup`

### Sentinel-Home *(in development)*

Future home-security application. Stack not yet defined; skeleton CI pipelines and Traefik routing are in place. Sentinel-Home becomes active once `Sentinel-Home/` submodule content is committed.

---

## Infrastructure Components

### Traefik Gateway (`traefik/`)

Traefik v3.1 reverse proxy that fronts all applications on a shared `home-net` bridge network.

- **Entry points**: HTTP `:80` (→ HTTPS redirect), HTTPS `:443`, RTSP TCP `:8554`
- **TLS**: self-signed via file provider (swap for ACME/Let's Encrypt for public domains)
- **Dynamic config**: `traefik/dynamic/` — one file per app with routers, services, middlewares
- **Dashbard**: `https://traefik.home` (basic auth via `TRAEFIK_DASHBOARD_USERS`)

Start first — it creates the shared `home-net` Docker network:

```bash
docker compose --project-name traefik-gw --file traefik/docker-compose.yml up -d
```

### Authentik SSO (`traefik/`)

Authentik runs alongside Traefik as the shared identity provider for the gatewayed applications.

- **Host**: `https://auth.home`
- **Bootstrap**: initial setup is still available on `http://127.0.0.1:9000/if/flow/initial-setup/`
- **Wiring**: Traefik uses a shared forward-auth middleware so app routes can require Authentik before they reach the upstream service

### Jenkins CI/CD (`jenkins/`)

Declarative pipelines for all three apps, sharing a common library.

- **9 Jenkinsfiles**: frontend, backend, deploy × 3 apps
- **Shared library** (`jenkins/shared-library/vars/`): `deployApp`, `sonarScan`, `zapScan`, `playwright508`, `buildSummary`, `publishToMinio`
- **Artifact storage**: every build uploads to MinIO at `{bucket}/{app}/{component}/{date}/{build#}/{type}/`
- **Reports**: dark-theme HTML dashboard aggregating ZAP DAST, SonarQube, 508 accessibility, and test results
- **Bootstrap**: 3-step guide in [jenkins/README.md](jenkins/README.md)

### AWX Automation (`awx/`)

Ansible automation platform (AWX community edition) for provisioning and operating the host server.

- **Centralized variables**: `awx/inventory/group_vars/all.yml` — single source of truth for all port assignments, network names, domain names, data-directory UIDs, cluster names, and secret references
- **3 Execution Environments**: `base-ee`, `docker-ee`, `k8s-ee`
- **13 Job Templates**: Infra (provision, docker, k3d) + Deploy (per-app + full-stack) + Ops (health, backup, maintenance, update, models)
- **Workflows**: Bootstrap Host, Full Stack Deploy
- **Schedules**: health check every 15 min, daily backup, weekly maintenance
- **AWX as code**: `awx/awx_config/configure_awx.yml` provisions all AWX objects idempotently

---

## Getting Started

### Prerequisites

A Linux host (Ubuntu 22.04+ recommended) with:
- SSH access as a sudoer
- `~/.ssh/id_ed25519` for git operations

### 1 — Clone

```bash
git clone --recurse-submodules git@github.com:fntundi/Local-Stuff.git
cd Local-Stuff
```

### 2 — Deploy AWX

```bash
cd awx
cp .env.example .env   # fill in secrets
docker compose up -d
```

See [awx/README.md](awx/README.md) for the full bootstrap sequence.

### 3 — Run Bootstrap Workflow

In AWX UI → **Templates → Bootstrap Host → Launch**

This provisions the server (user, SSH, firewall), installs Docker, and installs k3d tools.

### 4 — Deploy the Full Stack

In AWX UI → **Templates → Full Stack Deploy → Launch**

Or run playbooks directly:

```bash
ansible-playbook awx/playbooks/platform/00_deploy_all.yml \
  -i awx/inventory/hosts.yml
```

### Local `/etc/hosts`

```
127.0.0.1  homecam.home
127.0.0.1  aistack.home jupyter.aistack.home mlflow.aistack.home
127.0.0.1  minio.aistack.home litellm.aistack.home portainer.aistack.home
127.0.0.1  langfuse.aistack.home opensearch.aistack.home auth.home
127.0.0.1  sentinel.home traefik.home
```

---

## Updating Submodules

```bash
# Pull latest for all submodules
git submodule update --remote --merge

# Pull latest for one submodule
git submodule update --remote --merge local-aistack
```

After updating, commit the new submodule pointer:

```bash
git add local-aistack
git commit -m "chore: bump local-aistack to latest"
```

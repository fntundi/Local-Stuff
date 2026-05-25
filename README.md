# Home Platform

Infrastructure-as-code for a self-hosted home automation and development platform.
All application infrastructure is deployed and managed by **AWX**. **Jenkins** handles
application-level CI/CD (build, test, and Kubernetes deploy). Identity for every service
is federated through **Authentik**.

---

## Architecture

```
Internet
    │  HTTPS / Let's Encrypt
    ▼
┌─────────────────────────────────────────────────────────────┐
│                     home-server  (172.20.0.0/16 home-net)   │
│                                                             │
│  Traefik ──► Authentik (SSO)                                │
│           ├─► AWX (autoans.blackiechan.net)                 │
│           ├─► HomeCam k3d (homecam.blackiechan.net)         │
│           ├─► AI Platform compose (chat/api/lab/mlflow/…)   │
│           └─► Sentinel-Home k3d (sentinel.blackiechan.net)  │
└─────────────────────────────────────────────────────────────┘

LAN-attached Ubuntu VMs (reach home-server via blackiechan.net or LAN IP)
  ├── sonarqube-server  192.168.1.110  sonarqube.blackiechan.net
  ├── jenkins-server    192.168.1.111  leeeroyy.blackiechan.net
  ├── jenkins-runner    192.168.1.112  (agent, no public hostname)
  └── gitlab-server     192.168.4.14   gitlab.blackiechan.net
```

### Deployment responsibility split

| Layer | Tool | What it does |
|---|---|---|
| Infrastructure | **AWX** | Provision VMs, install Docker/k3d, deploy all service stacks |
| Application CI/CD | **Jenkins** | Build images, run tests, `kubectl apply` to k3d clusters |
| Identity | **Authentik** | SSO for every service (OIDC, SAML, or forward-auth proxy) |
| Routing / TLS | **Traefik** | Reverse proxy, Let's Encrypt certs, rate limiting, security headers |

---

## Services

### Home Server (Docker host — `home-net` 172.20.0.0/16)

| Service | URL | Auth | Notes |
|---|---|---|---|
| **Traefik** | `traefik.blackiechan.net` | htpasswd | Reverse proxy + TLS termination |
| **Authentik** | `authentik.blackiechan.net` | native | SSO identity provider |
| **AWX** | `autoans.blackiechan.net` | Authentik forward-auth | Ansible automation |
| **HomeCam** | `homecam.blackiechan.net` | Authentik forward-auth | Camera NOC — Go backend + React frontend, k3d |
| **AI Platform** | `chat.blackiechan.net` + 9 sub-domains | Authentik forward-auth | Ollama, OpenWebUI, LiteLLM, MLflow, Qdrant, Langfuse, MinIO, JupyterLab, Activepieces, Portainer |
| **Sentinel-Home** | `sentinel.blackiechan.net` | Authentik forward-auth | Home security platform — k3d (in development) |

### Isolated Ubuntu VMs (LAN-attached, connected via `blackiechan.net` domain)

| Service | URL | Auth | Host IP |
|---|---|---|---|
| **Jenkins** | `leeeroyy.blackiechan.net` | Authentik OIDC (`oic-auth` plugin) | 192.168.1.111 |
| **SonarQube** | `sonarqube.blackiechan.net` | Authentik SAML | 192.168.1.110 |
| **GitLab CE** | `gitlab.blackiechan.net` | Authentik OIDC (omniauth) | 192.168.4.14 |

> All VM-hosted services have `/etc/hosts` entries injected by AWX so that every
> `blackiechan.net` hostname resolves to the home-server LAN IP. This provides full
> inter-service connectivity (e.g., Jenkins → Authentik, Jenkins → MinIO, Jenkins → SonarQube)
> without requiring public DNS on the LAN.

---

## Repository Layout

```
Local-Stuff/
├── traefik/                  Traefik v3 — static config + dynamic routing per app
│   ├── traefik.yml           Static: entryPoints, ACME resolvers, logging
│   ├── docker-compose.yml    Stack: traefik + whoami (health test)
│   └── dynamic/              One .yml per service (authentik, awx, gitlab, homecam,
│                             jenkins, local-aistack, sentinel-home, sonarqube)
│
├── authentik/                Authentik SSO identity provider
│   ├── docker-compose.yml    Stack: postgresql + server + worker
│   ├── .env.example          Required secrets — copy to .env before deploying
│   └── blueprints/           Authentik Blueprint YAML — one per integrated application
│       ├── homecam.yaml          Proxy provider (Traefik forward-auth)
│       ├── local-aistack.yaml    Proxy provider (Traefik forward-auth)
│       ├── sentinel-home.yaml    Proxy provider (Traefik forward-auth)
│       ├── jenkins.yaml          OAuth2/OIDC provider
│       ├── gitlab.yaml           OAuth2/OIDC provider
│       └── sonarqube.yaml        SAML provider
│
├── awx/                      AWX — all infrastructure automation
│   ├── docker-compose.yml    Stack: awx-web, task, postgres, redis, receptor
│   ├── .env.example          Required secrets — copy to .env before deploying
│   ├── inventory/
│   │   ├── hosts.yml             All managed hosts (home-server + 4 CI VMs)
│   │   └── group_vars/all.yml    SINGLE SOURCE OF TRUTH for all platform variables
│   ├── playbooks/
│   │   ├── infra/                OS hardening + Docker + k3d tools (3 playbooks)
│   │   ├── platform/             Deploy Traefik, Authentik, app stacks (7 playbooks)
│   │   ├── ci/                   Deploy Jenkins, runner, SonarQube, GitLab (5 playbooks)
│   │   └── ops/                  Health check, backup, maintenance, models (7 playbooks)
│   ├── roles/                14 Ansible roles (see awx/README.md for full list)
│   ├── execution-environments/  3 container images used by AWX jobs
│   │   ├── base-ee/          OS provisioning (no Docker socket needed)
│   │   ├── docker-ee/        Docker Compose deployments
│   │   └── k8s-ee/           Kubernetes + k3d + Helm deployments
│   └── awx_config/
│       └── configure_awx.yml    AWX-as-code: idempotent setup of all AWX objects
│
├── jenkins/                  Jenkins CI/CD definitions
│   ├── casc/jenkins.yaml     JCasC: Authentik OIDC realm, role-strategy, credentials,
│   │                         SonarQube server, shared library, views
│   ├── job-dsl/seed.groovy   Seed job — creates all folders and pipeline jobs
│   ├── shared-library/vars/  6 pipeline functions: sonarScan, zapScan, playwright508,
│   │                         buildSummary, publishToMinio, deployApp
│   └── pipelines/            Jenkinsfiles per app × component
│       ├── homecam/           frontend/, backend/, deploy/
│       ├── local-aistack/     frontend/, backend/, deploy/
│       └── sentinel-home/     frontend/, backend/, deploy/
│
├── HomeCam/                  [submodule] fntundi/HomeCam
├── local-aistack/            [submodule] fntundi/local-aistack
└── sentinel-home/            [submodule] fntundi/Sentinel-Home
```

---

## Deployment Guide

### Prerequisites

- Ubuntu 22.04 home server with a static or DDNS public IP
- DNS A-records: all `*.blackiechan.net` subdomains → home-server public IP
- SSH key `~/.ssh/id_ed25519` with access to all managed hosts
- Python 3 + `pip install awxkit` on the control machine

### Step 1 — Clone with submodules

```bash
git clone --recurse-submodules git@github.com:fntundi/Local-Stuff.git
cd Local-Stuff
```

### Step 2 — Deploy AWX (manual, one time)

AWX bootstraps everything else. Run this on the home-server:

```bash
cd awx
cp .env.example .env          # fill in: AWX_ADMIN_PASSWORD, POSTGRES_PASSWORD,
                              #          SECRET_KEY (generate with: openssl rand -hex 32)
docker compose up -d
# Wait ~2 min, then open http://<home-server-ip>:8052
```

### Step 3 — Apply AWX configuration as code

```bash
ansible-playbook awx/awx_config/configure_awx.yml \
  -e "awx_host=http://localhost:8052" \
  -e "awx_username=admin" \
  -e "awx_password=<AWX_ADMIN_PASSWORD>"
```

This creates the **Home Platform Vault** credential type with all secret fields.
Open **AWX UI → Credentials → Home Platform Vault** and fill in every `vault_*` value.
(The Authentik OIDC client secrets for Jenkins/GitLab are filled in after Step 7.)

### Step 4 — Build execution environments

Run on the home-server (requires `ansible-builder`):

```bash
cd awx
bash awx_config/build_execution_environments.sh
```

### Step 5 — Bootstrap the home-server host

In AWX, navigate to **Templates → Workflows → Bootstrap Host**.
Set the limit to `home_servers` and launch. This runs:
`01 Provision Server → 02 Install Docker → 03 Install k3d Tools`

### Step 6 — Deploy the platform stack

Run **Workflows → Full Stack Deploy** (limit: `platform`).

Deployment order enforced by workflow:
```
Traefik (creates home-net)
    └─► Authentik
            ├─► HomeCam
            ├─► AI Platform
            └─► Sentinel-Home
```

### Step 7 — Import Authentik blueprints

Open `https://authentik.blackiechan.net` → **System → Blueprints → Import**.

Import each file from `authentik/blueprints/` in this order:
1. `homecam.yaml`, `local-aistack.yaml`, `sentinel-home.yaml` (proxy providers)
2. `jenkins.yaml`, `gitlab.yaml` (OIDC providers)
3. `sonarqube.yaml` (SAML provider)

After importing `jenkins.yaml` and `gitlab.yaml`, go to each application in Authentik,
open the provider, and **copy the Client ID and Client Secret**. Paste them into the
AWX vault credential (`vault_jenkins_oidc_client_id` / `vault_jenkins_oidc_client_secret`,
and the GitLab equivalents).

For `sonarqube.yaml`: copy the **Signing Certificate** from the SAML provider
into `vault_sonarqube_saml_cert`.

### Step 8 — Deploy CI services (Jenkins, SonarQube, GitLab)

Each CI service runs on its own Ubuntu VM. Use the dedicated AWX bootstrap workflows.
Each workflow runs: `01 Provision Host → 02 Install Docker → 03 Deploy`.

| Workflow | Limit | Deploys |
|---|---|---|
| **Bootstrap Jenkins** | `jenkins_servers` | Jenkins with Authentik OIDC |
| **Bootstrap Jenkins Runner** *(manual steps)* | `jenkins_runners` | SSH build agent |
| **Bootstrap SonarQube** | `sonarqube_servers` | SonarQube with Authentik SAML |
| **Bootstrap GitLab** | `gitlab_servers` | GitLab CE with Authentik OIDC |

> Before running Bootstrap Jenkins: ensure the AWX vault has valid
> `vault_jenkins_oidc_client_id` and `vault_jenkins_oidc_client_secret` values
> (copied in Step 7). The AWX role injects these into Jenkins JCasC at deploy time.

### Step 9 — Seed Jenkins jobs

1. Log into `https://leeeroyy.blackiechan.net` via Authentik
2. Run the **Seed Job** (pre-created by JCasC)
3. All pipeline folders and jobs are created from `jenkins/job-dsl/seed.groovy`

---

## Authentik SSO Integration Reference

| App | Protocol | Group required | How it works |
|---|---|---|---|
| AWX | Forward-auth proxy | *(any authenticated user)* | Traefik calls Authentik outpost before forwarding |
| HomeCam | Forward-auth proxy | `homecam-users` | Same — blueprint policy restricts to group |
| AI Platform | Forward-auth proxy | `aistack-users` | Same |
| Sentinel-Home | Forward-auth proxy | *(any authenticated user)* | Same |
| **Jenkins** | **OIDC** | `jenkins-users` / `jenkins-admins` | `oic-auth` plugin; groups map to role-strategy roles |
| **GitLab** | **OIDC** | `gitlab-users` | omniauth `openid_connect`; auto-links on first login |
| **SonarQube** | **SAML** | `sonarqube-users` | Built-in SAML auth; group attribute `groups` maps to SonarQube groups |

---

## Day-2 Operations

Scheduled jobs in AWX:

| Schedule | Job | Action |
|---|---|---|
| Every 15 min | **Ops \| Health Check** | HTTP probe every service; logs failures |
| Daily 02:00 | **Ops \| Backup All** | `pg_dump` + `mongodump` → MinIO; 30-day retention |
| Sunday 03:00 | **Ops \| Maintenance** | `docker system prune`, log rotation |

Ad-hoc AWX jobs:
- **Ops \| Update Platform** — pull latest images + rolling restart
- **Ops \| Manage Ollama Models** — pull / list / delete Ollama models
- **Ops \| Manage Authentik** — blueprint sync, password rotation

---

## Submodules

| Path | Repo | Description |
|---|---|---|
| `HomeCam/` | fntundi/HomeCam | Camera NOC — Go backend (Gin), React 19 frontend. Deployed to k3d cluster `sentinel-noc` via Jenkins deploy pipeline. NodePorts 31100 (HTTP) / 31101 (API). |
| `local-aistack/` | fntundi/local-aistack | Hybrid AI platform — 35+ services including Ollama, OpenWebUI, LiteLLM, MLflow, Qdrant, Langfuse, MinIO, JupyterLab, Activepieces, Portainer. Docker Compose on home-server. |
| `sentinel-home/` | fntundi/Sentinel-Home | Home security platform (in development). k3d cluster `sentinel-home`. NodePorts 31000 (HTTP) / 31001 (API). |

```bash
# Update all submodules to latest remote HEAD
git submodule update --remote --merge
```

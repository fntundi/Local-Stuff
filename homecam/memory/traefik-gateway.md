---
name: traefik-gateway
description: Traefik v3 reverse proxy gateway shared by all three Home Platform apps — network name, host routing, k3d integration
metadata:
  type: project
---

Traefik gateway stack lives at `HomeCam/traefik/`. Created 2026-05-22.

**Shared Docker network:** `home-net` (172.20.0.0/16) — created by the Traefik stack, all app stacks reference it as external.

**Host routing table:**
| Host | Routes to |
|------|-----------|
| `homecam.home` | HomeCam React frontend (k3d ingress-nginx :80) |
| `homecam.home/api` | HomeCam Go backend (:8001) |
| `homecam.home/hls` | MediaMTX HLS (:8888) |
| `aistack.home` | OpenWebUI (:3000) |
| `jupyter.aistack.home` | JupyterLab (:8888) |
| `mlflow.aistack.home` | MLflow (:5000) |
| `minio.aistack.home` | MinIO console (:9001) |
| `litellm.aistack.home` | LiteLLM (:4000) |
| `portainer.aistack.home` | Portainer (:9000) |
| `langfuse.aistack.home` | Langfuse |
| `opensearch.aistack.home` | OpenSearch Dashboards |
| `sentinel.home` | Sentinel-Home frontend (placeholder) |
| `traefik.home` | Traefik dashboard |

**HomeCam / k3d note:** HomeCam runs in k3d (Kubernetes). Traefik uses `host.docker.internal` to reach the k3d NodePort that maps to ingress-nginx. Routes are in `traefik/dynamic/homecam.yml` (file provider, not Docker labels).

**local-aistack:** Services run in Docker Compose on `home-net`. Routes in `traefik/dynamic/local-aistack.yml`.

**Start gateway first:** `docker compose --project-name traefik-gw --file traefik/docker-compose.yml up -d`

**Why:** User wanted all three apps (different clusters) behind a single Traefik proxy with shared networking.

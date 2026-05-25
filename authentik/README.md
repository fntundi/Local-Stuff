# Authentik

Standalone Authentik stack for the Home Platform.

This folder holds the Authentik deployment artifacts only:
- `docker-compose.yml` for the Authentik server, worker, and PostgreSQL
- `.env.example` for required secrets and runtime settings
- `blueprints/` for blueprints, flows, and application/provider config

Traefik routes that *use* Authentik live under `traefik/dynamic/`.

Bring it up after `traefik/` creates the shared `home-net` network:

```bash
docker compose --project-name authentik --file authentik/docker-compose.yml up -d
```

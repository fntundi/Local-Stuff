# AWX — Ansible Automation Platform

Deploys and manages the Home Platform using AWX (Ansible Tower community edition).
All playbooks, roles, inventory, and AWX configuration live under `awx/`.

## Directory Layout

```
awx/
├── docker-compose.yml              AWX deployment (web + task + postgres + redis + receptor)
├── .env.example                    AWX runtime secrets template
├── settings/local_settings.py     AWX Django settings override
├── receptor/receptor.conf          Receptor worker config
│
├── execution-environments/         Container images for running playbooks
│   ├── base-ee/                    ansible-core + posix + community.general
│   ├── docker-ee/                  base-ee + community.docker + Docker CLI
│   └── k8s-ee/                     docker-ee + kubernetes.core + kubectl/k3d/helm
│
├── inventory/
│   ├── hosts.yml                   Static inventory (home-server + isolated CI VMs)
│   └── group_vars/
│       └── all.yml                 ← Single source of truth for ALL platform variables
│
├── playbooks/
│   ├── infra/                      Server provisioning (run once, in order)
│   │   ├── 01_provision_server.yml  OS hardening, user, SSH, UFW, fail2ban
│   │   ├── 02_install_docker.yml    Docker Engine + Compose plugin
│   │   └── 03_install_k3d_tools.yml kubectl, k3d, Helm
│   ├── ci/                         Standalone SonarQube / Jenkins / runner deployments
│   │   ├── 10_deploy_sonarqube.yml  SonarQube + PostgreSQL on isolated Ubuntu VM
│   │   ├── 11_deploy_jenkins.yml    Jenkins controller with plugin/image packaging
│   │   └── 12_deploy_jenkins_runner.yml SSH build runner host
│   ├── platform/                   Application deployment
│   │   ├── 00_deploy_all.yml        Full stack (Traefik + Authentik → HomeCam → AI → BlackieFi)
│   │   ├── 10_deploy_traefik.yml    Traefik gateway + Authentik SSO
│   │   ├── 11_deploy_homecam.yml    HomeCam / Sentinel NOC
│   │   ├── 12_deploy_local_aistack.yml  Hybrid AI Platform
│   │   ├── 13_deploy_blackiefi.yml  BlackieFi (k3d only)
│   │   └── 14_deploy_sentinel_home.yml  Sentinel-Home (skeleton)
│   └── ops/                        Day-2 operations
│       ├── 20_health_check.yml      HTTP + Docker health probes
│       ├── 21_backup.yml            PostgreSQL + MongoDB backup + rotation
│       ├── 22_maintenance.yml       Docker prune + log rotation
│       ├── 23_update_platform.yml   Pull latest images + rolling restart
│       └── 24_manage_models.yml     Ollama model pull/list
│
├── roles/                          Reusable Ansible roles
│   ├── common/                     OS setup (user, SSH, UFW, fail2ban, sysctl)
│   ├── docker/                     Docker Engine installation
│   ├── k3d_tools/                  kubectl + k3d + Helm
│   ├── platform_dirs/              Data directory creation with correct UIDs
│   ├── traefik/                    Traefik gateway + Authentik deployment
│   ├── homecam/                    HomeCam (compose + k3d tasks)
│   ├── local_aistack/              AI Platform (compose + k3d tasks)
│   ├── blackiefi/                  BlackieFi (k3d only)
│   ├── sentinel_home/              Sentinel-Home (skeleton)
│   ├── jenkins/                    Jenkins controller image + JCasC bootstrap
│   ├── jenkins_runner/             SSH-based Jenkins build agent host
│   └── sonarqube/                  SonarQube + PostgreSQL deployment
│
└── awx_config/
    ├── configure_awx.yml           AWX-as-code: creates all orgs, credentials, projects, templates
    └── build_execution_environments.sh  Build + tag all EE images
```

## Quick Start

### 1 — Deploy AWX

```bash
cd awx
cp .env.example .env
# Fill in AWX_ADMIN_PASSWORD, AWX_PG_PASSWORD, AWX_REDIS_PASSWORD, AWX_SECRET_KEY
docker compose up -d
# Wait ~2 minutes for first-run initialization
docker compose exec awx-task awx-manage createsuperuser  # only if not using AWX_ADMIN_*
```

Access: `http://localhost:8052` (or map port in `.env`)

### 2 — Build Execution Environments

```bash
pip install ansible-builder
bash awx_config/build_execution_environments.sh
```

### 3 — Configure AWX (AWX as Code)

```bash
pip install awxkit
ansible-galaxy collection install awx.awx
ansible-playbook awx_config/configure_awx.yml \
  -e "awx_host=http://localhost:8052" \
  -e "awx_password=<your-admin-password>"
```

This creates the organization, credential type, all 22 job templates, 2 workflow templates, and 3 schedules in one run. Re-run any time to reconcile state (idempotent).

### 4 — Set Secrets

In the AWX UI: **Credentials → Home Platform Vault** → fill in all `vault_*` fields.
Include the Authentik secrets as well: `vault_authentik_secret_key` and `vault_authentik_pg_password`.
See `inventory/group_vars/all.yml` for the full list of `vault_*` references.

### 5 — Run Bootstrap Workflow

In AWX: **Templates → Bootstrap Host → Launch**

Or ad-hoc:
```bash
ansible-playbook playbooks/infra/01_provision_server.yml \
  -i inventory/hosts.yml \
  --become
```

## Centralized Variables

All platform variables live in `inventory/group_vars/all.yml`. Key sections:

| Section | What it controls |
|---------|-----------------|
| Platform identity | `platform_user`, home dir, repo root, env, deploy strategy |
| Git repos | All GitHub repo URLs |
| Docker networks | Network names and subnets for each stack |
| Domains | Local `.home` and public `blackiechan.net` hostnames |
| Ports | Host-level port assignments |
| Kubernetes | Cluster names, namespaces, NodePort offsets |
| Data directories | Paths and UIDs for persistent storage |
| App secrets | `vault_*` references resolved by AWX credential |

Override any variable at the job template, inventory, or host level in AWX without touching the file.

## Job Templates Reference

| Template | Playbook | EE | `become` |
|----------|----------|----|---------|
| Infra \| 01 Provision Server | `infra/01_provision_server.yml` | base-ee | yes |
| Infra \| 02 Install Docker | `infra/02_install_docker.yml` | base-ee | yes |
| Infra \| 03 Install k3d Tools | `infra/03_install_k3d_tools.yml` | base-ee | yes |
| Deploy \| Traefik Gateway | `platform/10_deploy_traefik.yml` | docker-ee | no |
| Deploy \| HomeCam | `platform/11_deploy_homecam.yml` | docker-ee | no |
| Deploy \| AI Platform | `platform/12_deploy_local_aistack.yml` | docker-ee | no |
| Deploy \| BlackieFi | `platform/13_deploy_blackiefi.yml` | k8s-ee | no |
| Deploy \| Full Stack | `platform/00_deploy_all.yml` | k8s-ee | no |
| SonarQube \| 01 Provision Host | `ci/01_provision_server.yml` | base-ee | yes |
| SonarQube \| 02 Install Docker | `ci/02_install_docker.yml` | base-ee | yes |
| SonarQube \| 03 Deploy | `ci/10_deploy_sonarqube.yml` | docker-ee | no |
| Jenkins \| 01 Provision Host | `ci/01_provision_server.yml` | base-ee | yes |
| Jenkins \| 02 Install Docker | `ci/02_install_docker.yml` | base-ee | yes |
| Jenkins \| 03 Deploy | `ci/11_deploy_jenkins.yml` | docker-ee | no |
| Jenkins Runner \| 01 Provision Host | `ci/01_provision_server.yml` | base-ee | yes |
| Jenkins Runner \| 02 Install Docker | `ci/02_install_docker.yml` | base-ee | yes |
| Jenkins Runner \| 03 Deploy | `ci/12_deploy_jenkins_runner.yml` | base-ee | yes |
| Ops \| Health Check | `ops/20_health_check.yml` | docker-ee | no |
| Ops \| Backup All | `ops/21_backup.yml` | docker-ee | no |
| Ops \| Maintenance | `ops/22_maintenance.yml` | docker-ee | no |
| Ops \| Update Platform | `ops/23_update_platform.yml` | docker-ee | no |
| Ops \| Manage Ollama Models | `ops/24_manage_models.yml` | docker-ee | no |

## Workflow Templates

| Workflow | Sequence |
|----------|---------|
| Bootstrap Host | Provision → Install Docker → Install k3d Tools |
| Full Stack Deploy | Traefik → (HomeCam \|\| AI Platform \|\| BlackieFi) in parallel |

## Schedules

| Schedule | Template | Frequency |
|----------|----------|-----------|
| Health Check | Ops \| Health Check | Every 15 minutes |
| Backup | Ops \| Backup All | Daily at 02:00 |
| Maintenance | Ops \| Maintenance | Sundays at 03:00 |

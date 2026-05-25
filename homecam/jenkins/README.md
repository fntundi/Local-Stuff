# Jenkins CI/CD — Home Platform

Centralized Jenkins pipeline definitions for **HomeCam**, **local-aistack**, and **Sentinel-Home**.

## Directory Layout

```
jenkins/
├── casc/                        # Jenkins Configuration as Code (JCasC)
│   └── jenkins.yaml             #   Full Jenkins config: credentials stubs, views, plugins
├── job-dsl/
│   └── seed.groovy              # Seed job: creates all folder/pipeline jobs in Jenkins
├── shared-library/
│   └── vars/                    # Reusable pipeline steps
│       ├── deployApp.groovy     #   kubectl/docker-compose deploy helper
│       ├── sonarScan.groovy     #   SonarQube analysis wrapper
│       ├── zapScan.groovy       #   OWASP ZAP DAST wrapper
│       └── playwright508.groovy #   Axe-core 508/WCAG accessibility wrapper
├── pipelines/
│   ├── global.env               # Shared env vars (registry, sonar URL, etc.)
│   ├── homecam/
│   │   ├── homecam.env          #   App-specific overrides
│   │   ├── frontend/Jenkinsfile #   React build → lint → test → sonar → zap → 508
│   │   ├── backend/Jenkinsfile  #   Go build → vet → test → sonar → zap
│   │   └── deploy/Jenkinsfile   #   k8s deploy (dev/staging/prod gated)
│   ├── local-aistack/
│   │   ├── local-aistack.env
│   │   ├── backend/Jenkinsfile  #   Python lint → pytest → sonar → zap
│   │   ├── frontend/Jenkinsfile #   WebUI smoke → zap → 508
│   │   └── deploy/Jenkinsfile   #   docker-compose deploy
│   └── sentinel-home/
│       ├── sentinel-home.env
│       ├── frontend/Jenkinsfile #   Skeleton (populate when stack is defined)
│       ├── backend/Jenkinsfile
│       └── deploy/Jenkinsfile
└── traefik/
    ├── docker-compose.yml       # Traefik gateway stack
    ├── traefik.yml              # Static config (entrypoints, dashboard, providers)
    └── dynamic/
        ├── homecam.yml          # Routing rules → HomeCam services
        ├── local-aistack.yml    # Routing rules → AI platform services
        └── sentinel-home.yml    # Routing rules → Sentinel-Home services
```

## Quick Start — Bootstrap Jenkins

### 1. Install required plugins
Apply `casc/jenkins.yaml` via **Manage Jenkins → Configuration as Code**.
The file declares all required plugin IDs — use the Plugin Manager to install them first,
or bake them into your Jenkins Docker image with `jenkins-plugin-cli`.

### 2. Register the shared library
In **Manage Jenkins → System → Global Pipeline Libraries** add:
- Name: `home-platform`
- Default version: `main`
- Source: this repo (SCM path `jenkins/shared-library`)

### 3. Run the seed job
Create a **freestyle** job named `_seed` that runs the Job DSL script at
`jenkins/job-dsl/seed.groovy`. Run it once — it builds the full folder tree and
all pipeline jobs automatically.

### 4. Configure credentials
Set these Jenkins credentials (IDs must match exactly):

| Credential ID            | Type            | Used by                  |
|--------------------------|-----------------|--------------------------|
| `sonarqube-token`        | Secret text     | All SonarQube stages     |
| `docker-registry-creds`  | Username/password| All build/push stages   |
| `kubeconfig-homecam`     | Secret file     | HomeCam deploy pipeline  |
| `kubeconfig-aistack`     | Secret file     | local-aistack deploy     |
| `kubeconfig-sentinel`    | Secret file     | Sentinel-Home deploy     |
| `git-credentials`        | Username/password| SCM checkout            |

## Environment Variable Reference

All pipelines load **`jenkins/pipelines/global.env`** first, then the app-specific
`.env` file, then Jenkins-level env vars (highest precedence).

Key variables tunable from the Jenkins UI (Build with Parameters):

| Parameter        | Default  | Effect                                       |
|------------------|----------|----------------------------------------------|
| `ENVIRONMENT`    | `dev`    | Target deploy environment                    |
| `SKIP_SONAR`     | `false`  | Skip SonarQube scan                          |
| `SKIP_ZAP`       | `false`  | Skip OWASP ZAP DAST scan                     |
| `SKIP_508`       | `false`  | Skip Playwright 508 accessibility tests      |
| `SKIP_DEPLOY`    | `false`  | Skip deploy stage                            |
| `DOCKER_TAG`     | `commit` | Override image tag                           |
| `ZAP_TARGET_URL` | env file | Override ZAP target (useful for ad-hoc runs) |

## Traefik Gateway

All three apps sit behind a single Traefik reverse proxy defined in `traefik/`.
Apps connect via the shared external Docker network `home-net`.

| Host                  | Routes to                           |
|-----------------------|-------------------------------------|
| `homecam.home`        | HomeCam frontend (React)            |
| `homecam.home/api`    | HomeCam backend (Go/Gin :8001)      |
| `homecam.home/hls`    | MediaMTX HLS streams                |
| `aistack.home`        | OpenWebUI (Ollama chat UI)          |
| `aistack.home/jupyter`| JupyterLab                          |
| `aistack.home/mlflow` | MLflow tracking server              |
| `aistack.home/minio`  | MinIO object storage console        |
| `aistack.home/api`    | local-aistack API server            |
| `sentinel.home`       | Sentinel-Home frontend (TBD)        |
| `traefik.home`        | Traefik dashboard                   |

Add these to `/etc/hosts` or a local DNS resolver:
```
127.0.0.1  homecam.home aistack.home sentinel.home traefik.home
```

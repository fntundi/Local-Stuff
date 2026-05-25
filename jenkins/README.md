# Jenkins CI/CD — Home Platform

Centralized pipeline definitions for **HomeCam**, **local-aistack**, and **Sentinel-Home**.

## Directory Layout

```
jenkins/
├── casc/
│   └── jenkins.yaml             # JCasC: views, credential stubs, SonarQube, shared library
├── job-dsl/
│   └── seed.groovy              # Run once to create all folders + pipeline jobs
├── shared-library/
│   └── vars/
│       ├── deployApp.groovy     # kubectl/compose deploy helper
│       ├── sonarScan.groovy     # SonarQube scan + API metrics → sonar/summary.json
│       ├── zapScan.groovy       # OWASP ZAP + alert parser → zap/summary.json
│       ├── playwright508.groovy # axe-core 508 tests → 508/summary.json
│       ├── buildSummary.groovy  # Reads all summaries → build-summary.html (dark theme)
│       └── publishToMinio.groovy# Uploads build-artifacts/ tree to MinIO
├── pipelines/
│   ├── global.env               # Shared vars (MINIO_ENDPOINT, MINIO_BUCKET, ZAP_IMAGE…)
│   ├── homecam/
│   │   ├── homecam.env
│   │   ├── frontend/Jenkinsfile
│   │   ├── backend/Jenkinsfile
│   │   └── deploy/Jenkinsfile
│   ├── local-aistack/
│   │   ├── local-aistack.env
│   │   ├── backend/Jenkinsfile
│   │   ├── frontend/Jenkinsfile
│   │   └── deploy/Jenkinsfile
│   └── sentinel-home/
│       ├── sentinel-home.env
│       ├── frontend/Jenkinsfile
│       ├── backend/Jenkinsfile
│       └── deploy/Jenkinsfile
└── traefik/
    ├── docker-compose.yml
    ├── traefik.yml
    └── dynamic/
        ├── homecam.yml
        ├── local-aistack.yml
        └── sentinel-home.yml
```

## Bootstrap (3 steps)

### 1. Install plugins
Apply `casc/jenkins.yaml` after installing (or baking into your Jenkins image):
`blueocean, configuration-as-code, job-dsl, workflow-aggregator, pipeline-stage-view,
git, docker-workflow, docker-plugin, sonar, htmlpublisher, junit, build-monitor-plugin,
dashboard-view, cloudbees-folder, credentials-binding, timestamper, ansicolor`

### 2. Set credentials (IDs must match exactly)

| Credential ID          | Type             | Purpose                          |
|------------------------|------------------|----------------------------------|
| `sonarqube-token`      | Secret text      | SonarQube analysis               |
| `docker-registry-creds`| Username/password| Registry push/pull               |
| `minio-credentials`    | Username/password| MinIO artifact storage (access key / secret key) |
| `kubeconfig-homecam`   | Secret file      | HomeCam k3d cluster              |
| `kubeconfig-aistack`   | Secret file      | local-aistack k8s                |
| `kubeconfig-sentinel`  | Secret file      | Sentinel-Home cluster            |
| `git-credentials`      | Username/password| SCM checkout                     |

### 3. Run the seed job
Create a freestyle `_seed` job pointing at `jenkins/job-dsl/seed.groovy`.
Run it once — all 9 folders and pipeline jobs appear automatically.

## Artifact Storage Layout (MinIO)

Every build uploads its full artifact tree to:
```
{MINIO_BUCKET}/{app}/{component}/{YYYY-MM-DD}/{BUILD_NUMBER}/
    ├── build-summary.html    ← dark-theme HTML dashboard (also published in Jenkins UI)
    ├── unit/
    │   ├── results.xml       JUnit XML
    │   └── coverage.*        Coverage report
    ├── smoke/
    │   └── results.txt       SMOKE_STATUS=PASS|FAIL key-value
    ├── functional/
    │   └── results.xml       JUnit XML
    ├── regression/
    │   └── results.xml       JUnit XML
    ├── zap/
    │   ├── report.html       Full OWASP ZAP HTML report
    │   ├── report.json       Raw ZAP JSON
    │   └── summary.json      { high, medium, low, informational, total, status }
    ├── sonar/
    │   └── summary.json      { gateStatus, bugs, vulnerabilities, smells, coverage, … }
    └── 508/
        ├── playwright-report/  Full Playwright HTML
        └── summary.json      { violations, passes, incomplete, status }
```

Default bucket: `ci-artifacts`  
Override with `MINIO_BUCKET` env var or pipeline parameter.

## Key Environment Variables

All overridable at the Jenkins job level or via pipeline parameters:

| Variable           | Default                  | Scope    |
|--------------------|--------------------------|----------|
| `MINIO_ENDPOINT`   | `http://minio:9000`      | global   |
| `MINIO_BUCKET`     | `ci-artifacts`           | global   |
| `SONARQUBE_URL`    | `http://sonarqube:9000`  | global   |
| `ZAP_IMAGE`        | `ghcr.io/zaproxy/zaproxy:stable` | global |
| `ZAP_SCAN_MODE`    | `baseline`               | global   |
| `WCAG_LEVEL`       | `wcag21aa`               | global   |
| `TRAEFIK_NETWORK`  | `home-net`               | global   |
| `ENVIRONMENT`      | `dev`                    | per-run param |
| `SKIP_SONAR`       | `false`                  | per-run param |
| `SKIP_ZAP`         | `false`                  | per-run param |
| `SKIP_508`         | `false`                  | per-run param |
| `SKIP_DEPLOY`      | `true`                   | per-run param |
| `DOCKER_TAG`       | commit SHA               | per-run param |

## Traefik Gateway

Start the gateway first (creates `home-net`):
```bash
docker compose --project-name traefik-gw --file traefik/docker-compose.yml up -d
```

Add to `/etc/hosts`:
```
127.0.0.1  homecam.home aistack.home jupyter.aistack.home mlflow.aistack.home
127.0.0.1  minio.aistack.home litellm.aistack.home portainer.aistack.home
127.0.0.1  langfuse.aistack.home opensearch.aistack.home
127.0.0.1  sentinel.home traefik.home whoami.home
```

---
name: jenkins-cicd
description: Jenkins CI/CD pipeline infrastructure for HomeCam, local-aistack, Sentinel-Home — location, structure, bootstrap steps, credential IDs
metadata:
  type: project
---

Full Jenkins pipeline suite lives at `HomeCam/jenkins/`. Created 2026-05-22.

**Structure:**
- `casc/jenkins.yaml` — JCasC: views, credentials stubs, SonarQube server, shared library registration
- `job-dsl/seed.groovy` — Run once as a freestyle "seed" job to create all folders + pipeline jobs
- `shared-library/vars/` — 4 reusable steps: `deployApp`, `sonarScan`, `zapScan`, `playwright508`
- `pipelines/global.env` + per-app `.env` files — all tunable env vars
- `pipelines/{homecam,local-aistack,sentinel-home}/{frontend,backend,deploy}/Jenkinsfile`

**Required Jenkins credential IDs (must match exactly):**
- `sonarqube-token` — SonarQube secret text
- `docker-registry-creds` — username/password
- `kubeconfig-homecam`, `kubeconfig-aistack`, `kubeconfig-sentinel` — Secret files
- `git-credentials` — username/password

**Bootstrap order:**
1. Install plugins listed in `casc/jenkins.yaml` comments
2. Apply `casc/jenkins.yaml` via JCasC
3. Register shared library `home-platform` pointing to `jenkins/shared-library`
4. Create freestyle `_seed` job running `jenkins/job-dsl/seed.groovy`
5. Run `_seed` — creates all folders and pipeline jobs

**Why:** User requested multi-app CI/CD with frontend/backend segregation, smoke/functional/regression/ZAP/SonarQube/Playwright-508 stages, env-var management, and single-pane-of-glass Jenkins UI.

**How to apply:** When asked about CI pipelines, tests, or Jenkins for this platform, start from this structure.

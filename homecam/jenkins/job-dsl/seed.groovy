// ============================================================================
// Home Platform — Jenkins Job DSL Seed Script
// Run once from a freestyle "seed" job to bootstrap the full folder + job tree.
// Job DSL plugin: https://plugins.jenkins.io/job-dsl/
// ============================================================================

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

def SCM_URL       = binding.variables['GIT_REPO_URL']    ?: 'https://github.com/your-org/HomeCam.git'
def SCM_CREDS     = 'git-credentials'
def LIB_BRANCH    = 'main'
def JENKINS_BASE  = 'jenkins/pipelines'

// Returns a pipelineJob config block pointing to a Jenkinsfile in the repo
def pipelineFromScm(String jobPath, String displayName, String description, String jenkinsfileRelPath) {
    pipelineJob(jobPath) {
        displayName(displayName)
        description(description)

        logRotator {
            numToKeep(50)
            daysToKeep(30)
            artifactNumToKeep(20)
        }

        definition {
            cpsScm {
                scm {
                    git {
                        remote {
                            url(SCM_URL)
                            credentials(SCM_CREDS)
                        }
                        branch("*/${LIB_BRANCH}")
                    }
                }
                scriptPath(jenkinsfileRelPath)
                lightweight(true)
            }
        }
    }
}

// ---------------------------------------------------------------------------
// Root-level description view helper (Build Monitor)
// ---------------------------------------------------------------------------
buildMonitorView('Platform Monitor') {
    description('Real-time build status for all Home Platform pipelines')
    jobs {
        regex(/.*/)
    }
    recurse(true)
}

// ============================================================================
// HomeCam
// ============================================================================
folder('HomeCam') {
    displayName('HomeCam — Security NOC')
    description('Go/Gin backend + React 19 frontend | MongoDB | MediaMTX streaming')
}

folder('HomeCam/Frontend') {
    displayName('Frontend')
    description('React 19 pipeline stages — lint, build, test, security, accessibility')
}

folder('HomeCam/Backend') {
    displayName('Backend')
    description('Go/Gin pipeline stages — build, vet, test, security')
}

folder('HomeCam/Deploy') {
    displayName('Deploy')
    description('HomeCam Kubernetes (k3d / sentinel-noc) deployments')
}

// HomeCam / Frontend jobs
pipelineFromScm(
    'HomeCam/Frontend/CI',
    '🔵 Frontend CI',
    'Lint → Build → Unit Tests → Smoke → Functional → Regression → SonarQube → ZAP → 508',
    "${JENKINS_BASE}/homecam/frontend/Jenkinsfile"
)

// HomeCam / Backend jobs
pipelineFromScm(
    'HomeCam/Backend/CI',
    '🟢 Backend CI',
    'go vet → Build → Unit Tests → Smoke → Functional → Regression → SonarQube → ZAP',
    "${JENKINS_BASE}/homecam/backend/Jenkinsfile"
)

// HomeCam / Deploy
pipelineFromScm(
    'HomeCam/Deploy/Deploy',
    '🚀 Deploy',
    'Deploy HomeCam to dev / staging / prod k8s cluster (approval-gated for prod)',
    "${JENKINS_BASE}/homecam/deploy/Jenkinsfile"
)

// ============================================================================
// local-aistack
// ============================================================================
folder('local-aistack') {
    displayName('local-aistack — AI Platform')
    description('Hybrid AI platform: Ollama, OpenWebUI, MLflow, Qdrant, JupyterLab, LiteLLM, and more')
}

folder('local-aistack/Backend') {
    displayName('Backend')
    description('Python API server, code-executor, workflows-runner, general-mcp — lint, pytest, security')
}

folder('local-aistack/Frontend') {
    displayName('Frontend (Web UIs)')
    description('OpenWebUI, JupyterLab, MLflow dashboards — ZAP and 508 accessibility scans')
}

folder('local-aistack/Deploy') {
    displayName('Deploy')
    description('local-aistack Docker Compose and Kubernetes deployments')
}

pipelineFromScm(
    'local-aistack/Backend/CI',
    '🟢 Backend CI',
    'flake8 → black → mypy → pytest → SonarQube → ZAP',
    "${JENKINS_BASE}/local-aistack/backend/Jenkinsfile"
)

pipelineFromScm(
    'local-aistack/Frontend/CI',
    '🔵 Frontend CI (Web UIs)',
    'Service health → ZAP → Playwright 508 accessibility on OpenWebUI / JupyterLab',
    "${JENKINS_BASE}/local-aistack/frontend/Jenkinsfile"
)

pipelineFromScm(
    'local-aistack/Deploy/Deploy',
    '🚀 Deploy',
    'Deploy local-aistack (docker-compose or k8s) with env-file selection',
    "${JENKINS_BASE}/local-aistack/deploy/Jenkinsfile"
)

// ============================================================================
// Sentinel-Home
// ============================================================================
folder('Sentinel-Home') {
    displayName('Sentinel-Home')
    description('Sentinel-Home application — skeleton pipelines, ready to populate')
}

folder('Sentinel-Home/Frontend') {
    displayName('Frontend')
    description('Frontend CI pipeline (populate Jenkinsfile when stack is defined)')
}

folder('Sentinel-Home/Backend') {
    displayName('Backend')
    description('Backend CI pipeline (populate Jenkinsfile when stack is defined)')
}

folder('Sentinel-Home/Deploy') {
    displayName('Deploy')
    description('Sentinel-Home deployment pipeline')
}

pipelineFromScm(
    'Sentinel-Home/Frontend/CI',
    '🔵 Frontend CI',
    'Skeleton — lint, test, sonar, zap, 508 stages ready to configure',
    "${JENKINS_BASE}/sentinel-home/frontend/Jenkinsfile"
)

pipelineFromScm(
    'Sentinel-Home/Backend/CI',
    '🟢 Backend CI',
    'Skeleton — lint, test, sonar, zap stages ready to configure',
    "${JENKINS_BASE}/sentinel-home/backend/Jenkinsfile"
)

pipelineFromScm(
    'Sentinel-Home/Deploy/Deploy',
    '🚀 Deploy',
    'Sentinel-Home deployment pipeline',
    "${JENKINS_BASE}/sentinel-home/deploy/Jenkinsfile"
)

// ============================================================================
// Top-level list views (complement the Build Monitor)
// ============================================================================
listView('All Pipelines') {
    description('Every pipeline across all applications')
    recurse(true)
    jobs { regex(/.*/) }
    columns {
        status()
        weather()
        name()
        lastSuccess()
        lastFailure()
        lastDuration()
        buildButton()
    }
}

listView('Security Scans') {
    description('ZAP DAST and SonarQube jobs only')
    recurse(true)
    jobs { regex(/.*(sonar|zap|security|508).*/) }
    columns {
        status()
        weather()
        name()
        lastSuccess()
        lastFailure()
        lastDuration()
        buildButton()
    }
}

listView('Deployments') {
    description('All deployment pipelines')
    recurse(true)
    jobs { regex(/.*/Deploy/.*/) }
    columns {
        status()
        weather()
        name()
        lastSuccess()
        lastFailure()
        lastDuration()
        buildButton()
    }
}

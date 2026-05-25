// Home Platform — Jenkins Job DSL Seed Script
// Run once from a freestyle "_seed" job to bootstrap the full folder + job tree.

def SCM_URL      = binding.variables['GIT_REPO_URL'] ?: 'https://github.com/your-org/Local-Stuff.git'
def SCM_CREDS    = 'git-credentials'
def LIB_BRANCH   = 'main'
def JENKINS_BASE = 'jenkins/pipelines'

def pipelineFromScm(String jobPath, String displayName, String description, String scriptPath) {
    pipelineJob(jobPath) {
        displayName(displayName)
        description(description)
        logRotator { numToKeep(50); daysToKeep(30); artifactNumToKeep(20) }
        definition {
            cpsScm {
                scm {
                    git {
                        remote { url(SCM_URL); credentials(SCM_CREDS) }
                        branch("*/${LIB_BRANCH}")
                    }
                }
                scriptPath(scriptPath)
                lightweight(true)
            }
        }
    }
}

buildMonitorView('Platform Monitor') {
    description('Real-time build status for all Home Platform pipelines')
    jobs { regex(/.*/) }
    recurse(true)
}

// ── HomeCam ──────────────────────────────────────────────────────────────
folder('HomeCam') { displayName('HomeCam — Security NOC')
    description('Go/Gin backend + React 19 frontend | MongoDB | MediaMTX') }
folder('HomeCam/Frontend') { displayName('Frontend') }
folder('HomeCam/Backend')  { displayName('Backend')  }
folder('HomeCam/Deploy')   { displayName('Deploy')   }

pipelineFromScm('HomeCam/Frontend/CI', '🔵 Frontend CI',
    'Lint → Build → Unit → Smoke → Functional → Regression → SonarQube → ZAP → 508',
    "${JENKINS_BASE}/homecam/frontend/Jenkinsfile")
pipelineFromScm('HomeCam/Backend/CI',  '🟢 Backend CI',
    'go vet → Build → Unit → Smoke → Functional → Regression → SonarQube → ZAP',
    "${JENKINS_BASE}/homecam/backend/Jenkinsfile")
pipelineFromScm('HomeCam/Deploy/Deploy', '🚀 Deploy',
    'Deploy HomeCam to dev/staging/prod k8s (approval-gated for prod)',
    "${JENKINS_BASE}/homecam/deploy/Jenkinsfile")

// ── local-aistack ─────────────────────────────────────────────────────────
folder('local-aistack') { displayName('local-aistack — AI Platform')
    description('Ollama · OpenWebUI · MLflow · Qdrant · JupyterLab · LiteLLM') }
folder('local-aistack/Backend')  { displayName('Backend')  }
folder('local-aistack/Frontend') { displayName('Frontend (Web UIs)') }
folder('local-aistack/Deploy')   { displayName('Deploy')   }

pipelineFromScm('local-aistack/Backend/CI',  '🟢 Backend CI',
    'flake8 → black → mypy → pytest → SonarQube → ZAP',
    "${JENKINS_BASE}/local-aistack/backend/Jenkinsfile")
pipelineFromScm('local-aistack/Frontend/CI', '🔵 Frontend CI (Web UIs)',
    'Health checks → ZAP scans → Playwright 508 on OpenWebUI / JupyterLab',
    "${JENKINS_BASE}/local-aistack/frontend/Jenkinsfile")
pipelineFromScm('local-aistack/Deploy/Deploy', '🚀 Deploy',
    'Docker Compose or k8s deploy with env-file selection',
    "${JENKINS_BASE}/local-aistack/deploy/Jenkinsfile")

// ── Sentinel-Home ────────────────────────────────────────────────────────
folder('Sentinel-Home') { displayName('Sentinel-Home')
    description('Skeleton pipelines — ready to populate when stack is defined') }
folder('Sentinel-Home/Frontend') { displayName('Frontend') }
folder('Sentinel-Home/Backend')  { displayName('Backend')  }
folder('Sentinel-Home/Deploy')   { displayName('Deploy')   }

pipelineFromScm('Sentinel-Home/Frontend/CI', '🔵 Frontend CI',
    'Skeleton — all stages pre-wired, populate TODO sections',
    "${JENKINS_BASE}/sentinel-home/frontend/Jenkinsfile")
pipelineFromScm('Sentinel-Home/Backend/CI',  '🟢 Backend CI',
    'Skeleton — all stages pre-wired, populate TODO sections',
    "${JENKINS_BASE}/sentinel-home/backend/Jenkinsfile")
pipelineFromScm('Sentinel-Home/Deploy/Deploy', '🚀 Deploy',
    'Sentinel-Home deployment pipeline',
    "${JENKINS_BASE}/sentinel-home/deploy/Jenkinsfile")

// ── Views ─────────────────────────────────────────────────────────────────
listView('All Pipelines') {
    description('Every pipeline'); recurse(true); jobs { regex(/.*/) }
    columns { status(); weather(); name(); lastSuccess(); lastFailure(); lastDuration(); buildButton() }
}
listView('Security Scans') {
    description('ZAP DAST and SonarQube jobs'); recurse(true); jobs { regex(/.*(sonar|zap|508).*/) }
    columns { status(); weather(); name(); lastSuccess(); lastFailure(); lastDuration(); buildButton() }
}
listView('Deployments') {
    description('All deploy pipelines'); recurse(true); jobs { regex(/.*/Deploy/.*/) }
    columns { status(); weather(); name(); lastSuccess(); lastFailure(); lastDuration(); buildButton() }
}

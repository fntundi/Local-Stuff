// Shared library step: sonarScan
// Runs sonar-scanner and waits for the quality gate.
// Queries the SonarQube API for key metrics and writes sonar/summary.json:
//   { gateStatus, bugs, vulnerabilities, smells, coverage, duplications,
//     securityHotspots, ncloc, projectKey, sonarUrl }
//
// Usage:
//   sonarScan projectKey: 'homecam-backend', sources: './backend', language: 'go'
//   sonarScan projectKey: 'aistack-backend',  sources: './src',     language: 'python'

def call(Map config = [:]) {
    def projectKey     = config.projectKey     ?: error('sonarScan: projectKey is required')
    def projectName    = config.projectName    ?: projectKey
    def sources        = config.sources        ?: '.'
    def language       = config.language       ?: 'generic'
    def coverageReport = config.coverageReport ?: ''
    def exclusions     = config.exclusions     ?: '**/node_modules/**,**/.git/**,**/vendor/**,**/__pycache__/**'
    def sonarServer    = config.sonarServer    ?: 'SonarQube'
    def qualityGate    = config.qualityGate    ?: true
    def extraProps     = config.extraProps     ?: ''
    def artifactDir    = config.artifactDir    ?: env.ARTIFACT_DIR ?: 'build-artifacts'

    def sonarDir = "${artifactDir}/sonar"
    sh "mkdir -p ${sonarDir}"

    // ── Run scanner ────────────────────────────────────────────────────────
    withSonarQubeEnv(sonarServer) {
        def args = [
            "-Dsonar.projectKey=${projectKey}",
            "-Dsonar.projectName='${projectName}'",
            "-Dsonar.sources=${sources}",
            "-Dsonar.exclusions=${exclusions}",
        ]
        if (language == 'go') {
            args += ["-Dsonar.language=go",
                     "-Dsonar.go.coverage.reportPaths=${coverageReport ?: 'coverage.out'}"]
        } else if (language == 'python') {
            args += ["-Dsonar.python.version=3",
                     "-Dsonar.python.coverage.reportPaths=${coverageReport ?: 'coverage.xml'}"]
        } else if (language in ['js', 'javascript', 'typescript']) {
            args += ["-Dsonar.javascript.lcov.reportPaths=${coverageReport ?: 'coverage/lcov.info'}",
                     "-Dsonar.typescript.lcov.reportPaths=${coverageReport ?: 'coverage/lcov.info'}"]
        }
        if (extraProps) {
            args += extraProps.split('\n').collect { it.trim() }.findAll { it }
        }
        sh "sonar-scanner ${args.join(' \\\n    ')}"
    }

    // ── Quality gate ───────────────────────────────────────────────────────
    def gateStatus = 'NOT_RUN'
    if (qualityGate) {
        timeout(time: 10, unit: 'MINUTES') {
            def qg = waitForQualityGate()
            gateStatus = qg.status
            if (qg.status != 'OK') {
                unstable("SonarQube Quality Gate failed: ${qg.status}")
            }
        }
    }

    // ── Fetch metrics from SonarQube API ───────────────────────────────────
    withSonarQubeEnv(sonarServer) {
        sh """
            python3 - <<'PYEOF'
import json, os, sys, urllib.request, urllib.error

sonar_url = os.environ.get('SONAR_HOST_URL', os.environ.get('JENKINS_SONARQUBE_URL', 'http://sonarqube:9000')).rstrip('/')
token     = os.environ.get('SONAR_AUTH_TOKEN', os.environ.get('SONAR_TOKEN', ''))
key       = '${projectKey}'

metrics = 'bugs,vulnerabilities,code_smells,coverage,duplicated_lines_density,security_hotspots,ncloc'
url = f'{sonar_url}/api/measures/component?component={key}&metricKeys={metrics}'

headers = {}
if token:
    import base64
    creds = base64.b64encode(f'{token}:'.encode()).decode()
    headers['Authorization'] = f'Basic {creds}'

result = {
    'projectKey':      key,
    'sonarUrl':        sonar_url,
    'gateStatus':      '${gateStatus}',
    'bugs':            '—',
    'vulnerabilities': '—',
    'smells':          '—',
    'coverage':        '—',
    'duplications':    '—',
    'securityHotspots':'—',
    'ncloc':           '—',
}

try:
    req = urllib.request.Request(url, headers=headers)
    with urllib.request.urlopen(req, timeout=15) as resp:
        data = json.loads(resp.read())
    for m in data.get('component', {}).get('measures', []):
        k = m['metric']
        v = m.get('value', '—')
        if k == 'bugs':                        result['bugs'] = v
        elif k == 'vulnerabilities':           result['vulnerabilities'] = v
        elif k == 'code_smells':               result['smells'] = v
        elif k == 'coverage':                  result['coverage'] = v
        elif k == 'duplicated_lines_density':  result['duplications'] = v
        elif k == 'security_hotspots':         result['securityHotspots'] = v
        elif k == 'ncloc':                     result['ncloc'] = v
except Exception as e:
    print(f'Warning: could not fetch SonarQube metrics: {e}', file=sys.stderr)

output_path = '${sonarDir}/summary.json'
with open(output_path, 'w') as f:
    json.dump(result, f, indent=2)

print(f"SonarQube Summary: gate={result['gateStatus']} bugs={result['bugs']} "
      f"vulns={result['vulnerabilities']} coverage={result['coverage']}%")
PYEOF
        """
    }

    archiveArtifacts artifacts: "${sonarDir}/summary.json", allowEmptyArchive: true
}

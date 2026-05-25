// Shared library step: playwright508
// Runs Playwright + @axe-core/playwright WCAG 2.1 AA / Section 508 tests.
// Writes 508/summary.json:
//   { violations, passes, incomplete, inapplicable, status, wcagLevel, baseUrl }
//
// Usage:
//   playwright508 dir: './frontend', baseUrl: 'http://homecam-frontend:3000'
//   playwright508 dir: './tests/508', baseUrl: 'http://openwebui:3000', browsers: 'chromium'

def call(Map config = [:]) {
    def testDir      = config.dir          ?: './frontend'
    def baseUrl      = config.baseUrl      ?: error('playwright508: baseUrl is required')
    def browsers     = config.browsers     ?: 'chromium'
    def wcagLevel    = config.wcagLevel    ?: 'wcag21aa'
    def nodeImage    = config.nodeImage    ?: 'mcr.microsoft.com/playwright:v1.44.0-jammy'
    def failOnViolation = config.failOnViolation ?: true
    def artifactDir  = config.artifactDir  ?: env.ARTIFACT_DIR ?: 'build-artifacts'

    def reportDir    = "${artifactDir}/508"
    sh "mkdir -p ${reportDir}"

    def exitCode = sh(returnStatus: true, script: """
        docker run --rm \\
            --network ${env.TRAEFIK_NETWORK ?: env.JENKINS_TRAEFIK_NETWORK ?: 'home-net'} \\
            -v \${PWD}:/work \\
            -w /work/${testDir} \\
            -e BASE_URL=${baseUrl} \\
            -e WCAG_LEVEL=${wcagLevel} \\
            -e BROWSERS=${browsers} \\
            -e PW_REPORT_DIR=/work/${reportDir} \\
            -e PLAYWRIGHT_HTML_REPORT=/work/${reportDir}/playwright-report \\
            -e AXE_RESULTS_FILE=/work/${reportDir}/axe-results.json \\
            ${nodeImage} \\
            bash -c "
                npm ci --prefer-offline 2>/dev/null || yarn install --frozen-lockfile 2>/dev/null || true
                npx playwright test --project=${browsers} \\
                    --reporter=html,json \\
                    --output=/work/${reportDir}/test-results \\
                    2>&1 | tee /work/${reportDir}/playwright-output.txt
            "
    """)

    // ── Parse axe-results.json (if produced) to write summary.json ─────────
    sh """
        python3 - <<'PYEOF'
import json, os, sys, glob

report_dir = '${reportDir}'
summary_path = os.path.join(report_dir, 'summary.json')
exit_code = ${exitCode}

summary = {
    'baseUrl':       '${baseUrl}',
    'wcagLevel':     '${wcagLevel}',
    'browsers':      '${browsers}',
    'status':        'PASS' if exit_code == 0 else 'FAIL',
    'violations':    0,
    'passes':        0,
    'incomplete':    0,
    'inapplicable':  0,
    'topViolations': []
}

# Try to read axe-results.json
axe_path = os.path.join(report_dir, 'axe-results.json')
if os.path.exists(axe_path):
    try:
        with open(axe_path) as f:
            data = json.load(f)
        if isinstance(data, list):
            for page in data:
                summary['violations']   += len(page.get('violations',   []))
                summary['passes']       += len(page.get('passes',       []))
                summary['incomplete']   += len(page.get('incomplete',   []))
                summary['inapplicable'] += len(page.get('inapplicable', []))
                for v in page.get('violations', []):
                    summary['topViolations'].append({
                        'id':     v.get('id', ''),
                        'impact': v.get('impact', ''),
                        'description': v.get('description', '')[:120],
                        'count': len(v.get('nodes', []))
                    })
    except Exception as e:
        print(f'Warning: could not parse axe-results.json: {e}', file=sys.stderr)
else:
    # Fall back to exit code only
    summary['violations'] = 'unknown' if exit_code != 0 else 0

summary['topViolations'] = summary['topViolations'][:10]

with open(summary_path, 'w') as f:
    json.dump(summary, f, indent=2)

print(f"508 Summary: violations={summary['violations']} passes={summary['passes']} status={summary['status']}")
PYEOF
    """

    // ── Publish Playwright HTML report ─────────────────────────────────────
    publishHTML(target: [
        allowMissing         : true,
        alwaysLinkToLastBuild: true,
        keepAll              : true,
        reportDir            : "${reportDir}/playwright-report",
        reportFiles          : 'index.html',
        reportName           : '508 Accessibility',
        reportTitles         : 'WCAG / Section 508',
    ])

    archiveArtifacts artifacts: "${reportDir}/**/*.json", allowEmptyArchive: true

    if (exitCode != 0 && failOnViolation) {
        error("Playwright 508 tests found violations — see report")
    } else if (exitCode != 0) {
        unstable("Playwright 508 tests found violations — see report")
    }
}

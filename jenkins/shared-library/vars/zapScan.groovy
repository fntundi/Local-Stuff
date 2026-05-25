// Shared library step: zapScan
// Runs OWASP ZAP in baseline or full scan mode.
// After the scan, parses report.json to produce summary.json:
//   { high, medium, low, informational, total, status, target, mode }
// Stashes both report files for the pipeline post-block to collect.
//
// Usage:
//   zapScan target: 'http://homecam-backend:8001', mode: 'baseline', reportName: 'zap-backend'
//   zapScan target: 'http://homecam-frontend:3000', mode: 'full', artifactDir: 'build-artifacts'

def call(Map config = [:]) {
    def target       = config.target      ?: error('zapScan: target URL is required')
    def mode         = config.mode        ?: 'baseline'
    def reportName   = config.reportName  ?: 'zap-report'
    def zapImage     = config.zapImage    ?: env.ZAP_IMAGE ?: 'ghcr.io/zaproxy/zaproxy:stable'
    def rulesFile    = config.rulesFile   ?: ''
    def contextFile  = config.contextFile ?: ''
    def network      = config.network     ?: env.TRAEFIK_NETWORK ?: env.JENKINS_TRAEFIK_NETWORK ?: 'home-net'
    def failOnWarn   = config.failOnWarn  ?: false
    def artifactDir  = config.artifactDir ?: env.ARTIFACT_DIR ?: 'build-artifacts'

    def reportDir    = "${artifactDir}/zap"
    sh "mkdir -p ${reportDir}"

    def rulesArg   = rulesFile   ? "-c /zap/wrk/${rulesFile}"   : ''
    def contextArg = contextFile ? "-n /zap/wrk/${contextFile}" : ''
    def scanScript = mode == 'full' ? 'zap-full-scan.py' : 'zap-baseline.py'

    def zapCmd = """
        docker run --rm \\
            --network ${network} \\
            -v \${PWD}/${reportDir}:/zap/wrk/:rw \\
            ${zapImage} \\
            ${scanScript} \\
            -t ${target} \\
            -r report.html \\
            -J report.json \\
            ${rulesArg} ${contextArg} \\
            -I
    """.stripIndent().trim()

    def exitCode = sh(script: zapCmd, returnStatus: true)

    // ── Parse JSON report and write summary.json ───────────────────────────
    def summaryStatus = exitCode == 0 ? 'PASS' : exitCode == 1 ? 'WARN' : 'FAIL'
    sh """
        python3 - <<'PYEOF'
import json, os, sys

report_path = '${reportDir}/report.json'
summary_path = '${reportDir}/summary.json'

counts = {'high': 0, 'medium': 0, 'low': 0, 'informational': 0}
details = []

if os.path.exists(report_path):
    try:
        with open(report_path) as f:
            data = json.load(f)
        sites = data.get('site', [])
        if isinstance(sites, dict):
            sites = [sites]
        for site in sites:
            for alert in site.get('alerts', []):
                risk = alert.get('riskdesc', '').lower()
                count_val = int(alert.get('count', 0))
                if 'high' in risk:
                    counts['high'] += count_val
                elif 'medium' in risk:
                    counts['medium'] += count_val
                elif 'low' in risk:
                    counts['low'] += count_val
                else:
                    counts['informational'] += count_val
                if 'high' in risk or 'medium' in risk:
                    details.append({
                        'name': alert.get('name', ''),
                        'risk': alert.get('riskdesc', ''),
                        'count': count_val,
                        'url': alert.get('instances', [{}])[0].get('uri', '') if alert.get('instances') else ''
                    })
    except Exception as e:
        print(f'Warning: could not parse ZAP report: {e}', file=sys.stderr)

total = sum(counts.values())
summary = {
    'target': '${target}',
    'mode': '${mode}',
    'status': '${summaryStatus}',
    'high': counts['high'],
    'medium': counts['medium'],
    'low': counts['low'],
    'informational': counts['informational'],
    'total': total,
    'topAlerts': details[:5]
}

with open(summary_path, 'w') as f:
    json.dump(summary, f, indent=2)

print(f"ZAP Summary: {counts} | Total: {total} | Status: ${summaryStatus}")
PYEOF
    """

    // ── Publish HTML report in Jenkins UI ──────────────────────────────────
    publishHTML(target: [
        allowMissing         : true,
        alwaysLinkToLastBuild: true,
        keepAll              : true,
        reportDir            : reportDir,
        reportFiles          : 'report.html',
        reportName           : "ZAP — ${reportName}",
        reportTitles         : "OWASP ZAP: ${reportName}",
    ])

    archiveArtifacts artifacts: "${reportDir}/*.json", allowEmptyArchive: true

    // ── Evaluate exit code ─────────────────────────────────────────────────
    if (exitCode == 2) {
        error("ZAP scan FAILED (FAIL-level alerts) on ${target}")
    } else if (exitCode == 1 && failOnWarn) {
        error("ZAP scan found WARN-level alerts on ${target} (failOnWarn=true)")
    } else if (exitCode == 1) {
        unstable("ZAP scan found WARN-level alerts on ${target} — review report")
    }
}

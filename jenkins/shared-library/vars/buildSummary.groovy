// Shared library step: buildSummary
// Reads summary JSON files produced by zapScan, sonarScan, playwright508,
// and JUnit XML results, then generates a single HTML build-summary.html
// and publishes it as a Jenkins HTML report.
//
// Expected artifact layout (build-artifacts/):
//   unit/results.xml         JUnit XML
//   smoke/results.txt        Key=Value pairs (SMOKE_STATUS=PASS)
//   functional/results.xml   JUnit XML
//   regression/results.xml   JUnit XML
//   zap/summary.json         { high, medium, low, informational, total, status }
//   sonar/summary.json       { gateStatus, bugs, vulnerabilities, smells, coverage, duplications }
//   508/summary.json         { violations, passes, incomplete, inapplicable, status }
//
// Usage:
//   buildSummary(
//       app:           'homecam',
//       component:     'backend',
//       artifactDir:   'build-artifacts',
//       minioEndpoint: 'http://minio:9000',
//       minioBucket:   'ci-artifacts',
//       minioPath:     'homecam/backend/2026-05-22/42',
//   )

def call(Map config = [:]) {
    def app          = config.app          ?: env.APP        ?: 'unknown'
    def component    = config.component    ?: env.COMPONENT  ?: 'unknown'
    def artifactDir  = config.artifactDir  ?: 'build-artifacts'
    def buildNum     = config.buildNumber  ?: env.BUILD_NUMBER
    def buildResult  = config.buildResult  ?: (currentBuild.currentResult ?: 'UNKNOWN')
    def buildUrl     = config.buildUrl     ?: env.BUILD_URL  ?: '#'
    def minioEndpoint = config.minioEndpoint ?: env.MINIO_ENDPOINT ?: env.JENKINS_MINIO_ENDPOINT ?: 'http://minio:9000'
    def minioBucket  = config.minioBucket  ?: env.MINIO_BUCKET  ?: env.JENKINS_MINIO_BUCKET ?: 'ci-artifacts'
    def minioPath    = config.minioPath    ?: env.MINIO_REMOTE_PATH ?: "${app}/${component}"
    def buildDate    = env.BUILD_DATE ?: sh(returnStdout: true, script: 'date +%Y-%m-%d').trim()

    sh "mkdir -p ${artifactDir}"

    // ── Parse ZAP summary ──────────────────────────────────────────────────
    def zapData  = _readJsonSafe("${artifactDir}/zap/summary.json",
        [high: '—', medium: '—', low: '—', informational: '—', total: '—', status: 'NOT_RUN'])

    // ── Parse SonarQube summary ────────────────────────────────────────────
    def sonarData = _readJsonSafe("${artifactDir}/sonar/summary.json",
        [gateStatus: 'NOT_RUN', bugs: '—', vulnerabilities: '—', smells: '—',
         coverage: '—', duplications: '—', securityHotspots: '—'])

    // ── Parse 508 summary ─────────────────────────────────────────────────
    def a508Data = _readJsonSafe("${artifactDir}/508/summary.json",
        [violations: '—', passes: '—', incomplete: '—', status: 'NOT_RUN'])

    // ── Parse JUnit results ────────────────────────────────────────────────
    def unitData   = _parseJunitXml("${artifactDir}/unit/results.xml")
    def funcData   = _parseJunitXml("${artifactDir}/functional/results.xml")
    def regrData   = _parseJunitXml("${artifactDir}/regression/results.xml")

    // ── Parse smoke results ────────────────────────────────────────────────
    def smokeStatus = _readKeyValue("${artifactDir}/smoke/results.txt", 'SMOKE_STATUS', 'NOT_RUN')

    // ── Status colour helpers ──────────────────────────────────────────────
    def resultColor = buildResult == 'SUCCESS' ? '#2ea44f' : buildResult == 'UNSTABLE' ? '#dbab09' : '#d73a49'
    def gateColor   = sonarData.gateStatus == 'OK' ? '#2ea44f' : sonarData.gateStatus == 'NOT_RUN' ? '#6e7681' : '#d73a49'
    def zapColor    = zapData.status == 'PASS' ? '#2ea44f' : zapData.status == 'WARN' ? '#dbab09' : zapData.status == 'NOT_RUN' ? '#6e7681' : '#d73a49'
    def a508Color   = a508Data.status == 'PASS' ? '#2ea44f' : a508Data.status == 'NOT_RUN' ? '#6e7681' : '#d73a49'
    def smokeColor  = smokeStatus == 'PASS' ? '#2ea44f' : smokeStatus == 'NOT_RUN' ? '#6e7681' : '#d73a49'

    def html = """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Build Summary — ${app}/${component} #${buildNum}</title>
  <style>
    body { font-family: -apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;
           background:#0d1117; color:#c9d1d9; margin:0; padding:24px; }
    h1   { color:#e6edf3; font-size:1.4rem; margin-bottom:4px; }
    h2   { color:#e6edf3; font-size:1.1rem; margin:20px 0 8px; border-bottom:1px solid #30363d; padding-bottom:4px; }
    h3   { color:#8b949e; font-size:.95rem; margin:14px 0 6px; }
    .meta { color:#8b949e; font-size:.85rem; margin-bottom:20px; }
    .meta a { color:#58a6ff; text-decoration:none; }
    table { border-collapse:collapse; width:100%; margin-bottom:12px; }
    th,td { padding:6px 12px; text-align:left; border:1px solid #30363d; font-size:.88rem; }
    th    { background:#161b22; color:#8b949e; font-weight:600; }
    tr:nth-child(even) td { background:#0d1117; }
    tr:nth-child(odd)  td { background:#161b22; }
    .badge { display:inline-block; padding:2px 8px; border-radius:12px;
             font-size:.78rem; font-weight:600; color:#fff; }
    .section { background:#161b22; border:1px solid #30363d; border-radius:6px;
               padding:16px; margin-bottom:16px; }
    .grid2 { display:grid; grid-template-columns:1fr 1fr; gap:16px; }
    .grid3 { display:grid; grid-template-columns:1fr 1fr 1fr; gap:16px; }
    .kv    { display:flex; justify-content:space-between; padding:4px 0;
             border-bottom:1px solid #21262d; }
    .kv:last-child { border-bottom:none; }
    .kv-key   { color:#8b949e; font-size:.85rem; }
    .kv-val   { font-size:.85rem; font-weight:500; }
    .pill-red    { background:#d73a49; }
    .pill-yellow { background:#dbab09; }
    .pill-green  { background:#2ea44f; }
    .pill-grey   { background:#6e7681; }
    .high-risk   { color:#ff7b72; font-weight:bold; }
    .med-risk    { color:#ffa657; }
    .low-risk    { color:#f0e68c; }
    .artifacts   { font-size:.82rem; color:#8b949e; word-break:break-all; }
    .artifacts a { color:#58a6ff; }
  </style>
</head>
<body>

<h1>${app} / ${component} &nbsp;<span class="badge" style="background:${resultColor}">${buildResult}</span></h1>
<p class="meta">
  Build <a href="${buildUrl}">#${buildNum}</a> &nbsp;·&nbsp; ${buildDate} &nbsp;·&nbsp;
  <a href="${buildUrl}console">Console Output</a>
</p>

<!-- ── Overview ─────────────────────────────────────────────────── -->
<div class="section">
  <h2>Overview</h2>
  <div class="grid3">
    <div>
      <div class="kv"><span class="kv-key">Build Result</span>
        <span class="kv-val"><span class="badge" style="background:${resultColor}">${buildResult}</span></span></div>
      <div class="kv"><span class="kv-key">Smoke Tests</span>
        <span class="kv-val"><span class="badge" style="background:${smokeColor}">${smokeStatus}</span></span></div>
    </div>
    <div>
      <div class="kv"><span class="kv-key">SonarQube Gate</span>
        <span class="kv-val"><span class="badge" style="background:${gateColor}">${sonarData.gateStatus}</span></span></div>
      <div class="kv"><span class="kv-key">ZAP Scan</span>
        <span class="kv-val"><span class="badge" style="background:${zapColor}">${zapData.status}</span></span></div>
    </div>
    <div>
      <div class="kv"><span class="kv-key">508 Accessibility</span>
        <span class="kv-val"><span class="badge" style="background:${a508Color}">${a508Data.status}</span></span></div>
      <div class="kv"><span class="kv-key">Branch</span>
        <span class="kv-val">${env.GIT_BRANCH ?: env.BRANCH_NAME ?: '—'}</span></div>
    </div>
  </div>
</div>

<!-- ── Test Results ──────────────────────────────────────────────── -->
<div class="section">
  <h2>Test Results</h2>
  <table>
    <tr><th>Suite</th><th>Tests</th><th>Passed</th><th>Failed</th><th>Skipped</th><th>Duration</th><th>Status</th></tr>
    ${_testRow('Unit',       unitData)}
    ${_testRow('Functional', funcData)}
    ${_testRow('Regression', regrData)}
  </table>
</div>

<!-- ── Security Scans ────────────────────────────────────────────── -->
<div class="section">
  <h2>Security Scans</h2>
  <div class="grid2">

    <div>
      <h3>OWASP ZAP DAST &nbsp;<span class="badge" style="background:${zapColor}">${zapData.status}</span></h3>
      <table>
        <tr><th>Risk Level</th><th>Count</th></tr>
        <tr><td class="high-risk">High</td><td>${zapData.high}</td></tr>
        <tr><td class="med-risk">Medium</td><td>${zapData.medium}</td></tr>
        <tr><td class="low-risk">Low</td><td>${zapData.low}</td></tr>
        <tr><td>Informational</td><td>${zapData.informational}</td></tr>
        <tr><td><strong>Total</strong></td><td><strong>${zapData.total}</strong></td></tr>
      </table>
    </div>

    <div>
      <h3>SonarQube &nbsp;<span class="badge" style="background:${gateColor}">${sonarData.gateStatus}</span></h3>
      <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>Bugs</td><td>${sonarData.bugs}</td></tr>
        <tr><td>Vulnerabilities</td><td>${sonarData.vulnerabilities}</td></tr>
        <tr><td>Code Smells</td><td>${sonarData.smells}</td></tr>
        <tr><td>Coverage</td><td>${sonarData.coverage}%</td></tr>
        <tr><td>Duplications</td><td>${sonarData.duplications}%</td></tr>
        <tr><td>Security Hotspots</td><td>${sonarData.securityHotspots}</td></tr>
      </table>
    </div>

  </div>
</div>

<!-- ── 508 Accessibility ─────────────────────────────────────────── -->
<div class="section">
  <h2>Section 508 / WCAG 2.1 AA &nbsp;<span class="badge" style="background:${a508Color}">${a508Data.status}</span></h2>
  <table>
    <tr><th>Metric</th><th>Count</th></tr>
    <tr><td class="high-risk">Violations</td><td>${a508Data.violations}</td></tr>
    <tr><td class="low-risk">Incomplete (Needs Review)</td><td>${a508Data.incomplete}</td></tr>
    <tr><td style="color:#2ea44f">Passes</td><td>${a508Data.passes}</td></tr>
  </table>
</div>

<!-- ── Artifacts ────────────────────────────────────────────────── -->
<div class="section">
  <h2>Artifacts</h2>
  <p class="artifacts">
    MinIO: <a href="${minioEndpoint}/browser/${minioBucket}/${minioPath}">${minioEndpoint}/${minioBucket}/${minioPath}/</a>
  </p>
  <table>
    <tr><th>Type</th><th>MinIO Path</th></tr>
    <tr><td>Unit Tests</td><td class="artifacts">${minioPath}/unit/</td></tr>
    <tr><td>Functional Tests</td><td class="artifacts">${minioPath}/functional/</td></tr>
    <tr><td>Regression Tests</td><td class="artifacts">${minioPath}/regression/</td></tr>
    <tr><td>ZAP Report</td><td class="artifacts">${minioPath}/zap/</td></tr>
    <tr><td>SonarQube Summary</td><td class="artifacts">${minioPath}/sonar/</td></tr>
    <tr><td>508 Report</td><td class="artifacts">${minioPath}/508/</td></tr>
    <tr><td>This Summary</td><td class="artifacts">${minioPath}/build-summary.html</td></tr>
  </table>
</div>

</body>
</html>"""

    writeFile file: "${artifactDir}/build-summary.html", text: html

    publishHTML(target: [
        allowMissing        : false,
        alwaysLinkToLastBuild: true,
        keepAll             : true,
        reportDir           : artifactDir,
        reportFiles         : 'build-summary.html',
        reportName          : "Build Summary — ${app}/${component}",
        reportTitles        : "Build Summary",
    ])

    echo "Build summary written to ${artifactDir}/build-summary.html"
}

// ── Private helpers ────────────────────────────────────────────────────────

private Map _readJsonSafe(String filePath, Map defaults) {
    try {
        if (fileExists(filePath)) {
            def data = readJSON(file: filePath)
            return defaults + data   // merge: data wins over defaults
        }
    } catch (e) {
        echo "buildSummary: could not parse ${filePath}: ${e.message}"
    }
    return defaults
}

private String _readKeyValue(String filePath, String key, String defaultVal) {
    try {
        if (fileExists(filePath)) {
            def content = readFile(filePath)
            def match = content.readLines().find { it.startsWith("${key}=") }
            if (match) return match.split('=', 2)[1].trim()
        }
    } catch (e) { /* ignore */ }
    return defaultVal
}

private Map _parseJunitXml(String xmlPath) {
    def result = [tests: 0, passed: 0, failed: 0, skipped: 0, duration: '—']
    try {
        if (!fileExists(xmlPath)) return result
        def xml = readFile(xmlPath)
        // Parse totals from <testsuite> or <testsuites> attributes
        def tsMatch  = xml =~ /tests="(\d+)"/
        def failMatch = xml =~ /failures="(\d+)"/
        def errMatch  = xml =~ /errors="(\d+)"/
        def skipMatch = xml =~ /skipped="(\d+)"/
        def timeMatch = xml =~ /time="([\d.]+)"/
        if (tsMatch.find())  result.tests   = tsMatch.group(1).toInteger()
        if (failMatch.find()) result.failed  = failMatch.group(1).toInteger()
        if (errMatch.find())  result.failed += errMatch.group(1).toInteger()
        if (skipMatch.find()) result.skipped = skipMatch.group(1).toInteger()
        if (timeMatch.find()) result.duration = "${String.format('%.1f', timeMatch.group(1).toFloat())}s"
        result.passed = result.tests - result.failed - result.skipped
    } catch (e) {
        echo "buildSummary: could not parse ${xmlPath}: ${e.message}"
    }
    return result
}

private String _testRow(String suite, Map d) {
    if (d.tests == 0) {
        return "<tr><td>${suite}</td><td colspan='5' style='color:#6e7681'>Not run</td><td><span class='badge pill-grey'>NOT_RUN</span></td></tr>"
    }
    def statusColor = d.failed > 0 ? 'pill-red' : 'pill-green'
    def statusLabel = d.failed > 0 ? 'FAIL' : 'PASS'
    return """<tr>
      <td>${suite}</td>
      <td>${d.tests}</td>
      <td style='color:#2ea44f'>${d.passed}</td>
      <td style='color:${d.failed > 0 ? '#d73a49' : 'inherit'}'>${d.failed}</td>
      <td style='color:#8b949e'>${d.skipped}</td>
      <td>${d.duration}</td>
      <td><span class='badge ${statusColor}'>${statusLabel}</span></td>
    </tr>"""
}

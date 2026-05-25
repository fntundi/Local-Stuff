// Shared library step: zapScan
// Runs OWASP ZAP in baseline or full scan mode against a target URL.
// Usage:
//   zapScan target: 'http://homecam-backend:8001', mode: 'baseline', reportName: 'zap-backend'
//   zapScan target: 'http://homecam-frontend:3000', mode: 'full', rules: '/zap/wrk/rules.tsv'

def call(Map config = [:]) {
    def target      = config.target     ?: error('zapScan: target URL is required')
    def mode        = config.mode       ?: 'baseline'   // 'baseline' or 'full'
    def reportName  = config.reportName ?: 'zap-report'
    def zapImage    = config.zapImage   ?: env.ZAP_IMAGE ?: 'ghcr.io/zaproxy/zaproxy:stable'
    def rulesFile   = config.rulesFile  ?: ''
    def network     = config.network    ?: env.TRAEFIK_NETWORK ?: 'home-net'
    def failOnWarn  = config.failOnWarn ?: false
    def contextFile = config.contextFile ?: ''

    def reportDir  = "zap-reports/${reportName}"
    def htmlReport = "${reportDir}/report.html"
    def jsonReport = "${reportDir}/report.json"

    sh "mkdir -p ${reportDir}"

    def rulesArg    = rulesFile   ? "-c /zap/wrk/${rulesFile}" : ''
    def contextArg  = contextFile ? "-n /zap/wrk/${contextFile}" : ''

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
            ${rulesArg} \\
            ${contextArg} \\
            -I
    """.stripIndent().trim()

    def exitCode = sh(script: zapCmd, returnStatus: true)

    // ZAP exit codes: 0=pass, 1=warnings, 2=fail
    if (exitCode == 2) {
        error("ZAP scan FAILED (FAIL-level alerts) on ${target}")
    } else if (exitCode == 1 && failOnWarn) {
        error("ZAP scan found WARN-level alerts on ${target} (failOnWarn=true)")
    } else if (exitCode == 1) {
        unstable("ZAP scan found WARN-level alerts on ${target} — review report")
    }

    publishHTML(target: [
        allowMissing        : false,
        alwaysLinkToLastBuild: true,
        keepAll             : true,
        reportDir           : reportDir,
        reportFiles         : 'report.html',
        reportName          : "ZAP — ${reportName}",
        reportTitles        : "OWASP ZAP: ${reportName}",
    ])

    archiveArtifacts artifacts: "${reportDir}/*.json", allowEmptyArchive: true
}

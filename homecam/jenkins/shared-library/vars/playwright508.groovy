// Shared library step: playwright508
// Runs Playwright + axe-core WCAG 2.1 AA / Section 508 accessibility tests.
// Expects a playwright project under the given directory with:
//   - package.json with "test:508" script that uses @axe-core/playwright
//   - Results output to playwright-report/ and accessibility-results/
//
// Usage:
//   playwright508 dir: './frontend', baseUrl: 'http://homecam-frontend:3000'
//   playwright508 dir: './tests/508', baseUrl: 'http://openwebui:3000', browsers: 'chromium'

def call(Map config = [:]) {
    def testDir     = config.dir        ?: './frontend'
    def baseUrl     = config.baseUrl    ?: error('playwright508: baseUrl is required')
    def browsers    = config.browsers   ?: 'chromium'
    def wcagLevel   = config.wcagLevel  ?: 'wcag2aa'    // wcag2a, wcag2aa, wcag21aa, section508
    def reportDir   = config.reportDir  ?: 'accessibility-results'
    def nodeImage   = config.nodeImage  ?: 'mcr.microsoft.com/playwright:v1.44.0-jammy'
    def failOnViolation = config.failOnViolation ?: true

    sh "mkdir -p ${reportDir}"

    def exitCode = sh(returnStatus: true, script: """
        docker run --rm \\
            --network ${env.TRAEFIK_NETWORK ?: 'home-net'} \\
            -v \${PWD}:/work \\
            -w /work/${testDir} \\
            -e BASE_URL=${baseUrl} \\
            -e WCAG_LEVEL=${wcagLevel} \\
            -e BROWSERS=${browsers} \\
            -e PLAYWRIGHT_HTML_REPORT=/work/${reportDir}/report.html \\
            ${nodeImage} \\
            bash -c "
                npm ci --prefer-offline 2>/dev/null || yarn install --frozen-lockfile
                npx playwright test --project=${browsers} --reporter=html,json 2>&1
            "
    """)

    publishHTML(target: [
        allowMissing         : true,
        alwaysLinkToLastBuild: true,
        keepAll              : true,
        reportDir            : reportDir,
        reportFiles          : 'report.html',
        reportName           : '508 Accessibility Report',
        reportTitles         : 'WCAG / Section 508',
    ])

    archiveArtifacts artifacts: "${reportDir}/**/*.json", allowEmptyArchive: true

    if (exitCode != 0 && failOnViolation) {
        error("Playwright 508 accessibility tests found violations — see report")
    } else if (exitCode != 0) {
        unstable("Playwright 508 accessibility tests found violations — see report")
    }
}

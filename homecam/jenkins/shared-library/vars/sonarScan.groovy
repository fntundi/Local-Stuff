// Shared library step: sonarScan
// Usage:
//   sonarScan projectKey: 'homecam-backend', sources: './backend', language: 'go'
//   sonarScan projectKey: 'homecam-frontend', sources: './frontend/src', language: 'js'
//   sonarScan projectKey: 'aistack-backend', sources: './src', language: 'python', coverageReport: 'coverage.xml'

def call(Map config = [:]) {
    def projectKey      = config.projectKey      ?: error('sonarScan: projectKey is required')
    def projectName     = config.projectName     ?: projectKey
    def sources         = config.sources         ?: '.'
    def language        = config.language        ?: 'generic'
    def coverageReport  = config.coverageReport  ?: ''
    def exclusions      = config.exclusions      ?: '**/node_modules/**,**/.git/**,**/vendor/**,**/__pycache__/**'
    def sonarServer     = config.sonarServer     ?: 'SonarQube'
    def qualityGate     = config.qualityGate     ?: true
    def extraProps      = config.extraProps      ?: ''

    withSonarQubeEnv(sonarServer) {
        def scannerArgs = [
            "-Dsonar.projectKey=${projectKey}",
            "-Dsonar.projectName='${projectName}'",
            "-Dsonar.sources=${sources}",
            "-Dsonar.exclusions=${exclusions}",
        ]

        if (language == 'go') {
            scannerArgs += [
                "-Dsonar.language=go",
                "-Dsonar.go.coverage.reportPaths=${coverageReport ?: 'coverage.out'}",
            ]
        } else if (language == 'python') {
            scannerArgs += [
                "-Dsonar.python.version=3",
                "-Dsonar.python.coverage.reportPaths=${coverageReport ?: 'coverage.xml'}",
            ]
        } else if (language == 'js' || language == 'javascript' || language == 'typescript') {
            scannerArgs += [
                "-Dsonar.javascript.lcov.reportPaths=${coverageReport ?: 'coverage/lcov.info'}",
                "-Dsonar.typescript.lcov.reportPaths=${coverageReport ?: 'coverage/lcov.info'}",
            ]
        }

        if (extraProps) {
            scannerArgs += extraProps.split('\n').collect { it.trim() }.findAll { it }
        }

        sh "sonar-scanner ${scannerArgs.join(' \\\n    ')}"
    }

    if (qualityGate) {
        timeout(time: 10, unit: 'MINUTES') {
            def qg = waitForQualityGate()
            if (qg.status != 'OK') {
                unstable("SonarQube Quality Gate failed: ${qg.status}")
            }
        }
    }
}

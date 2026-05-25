// Shared library step: publishToMinio
// Uploads a local directory tree to MinIO (S3-compatible) using the mc client.
//
// MinIO path convention enforced by this step:
//   {bucket}/{app}/{component}/{YYYY-MM-DD}/{build_number}/{artifact_type}/
//
// Usage:
//   publishToMinio(
//       app:          'homecam',
//       component:    'backend',
//       buildDate:    '2026-05-22',
//       buildNumber:  env.BUILD_NUMBER,
//       localDir:     'build-artifacts',
//       artifactType: 'unit',           // optional: upload a single sub-dir
//   )
//
// Or upload a flat set of files:
//   publishToMinio(app: 'homecam', component: 'frontend', localDir: 'build-artifacts')
//   // uploads everything under build-artifacts/ preserving sub-directory names as artifact type

def call(Map config = [:]) {
    def app         = config.app         ?: env.APP         ?: error('publishToMinio: app is required')
    def component   = config.component   ?: env.COMPONENT   ?: error('publishToMinio: component is required')
    def buildDate   = config.buildDate   ?: env.BUILD_DATE  ?: sh(returnStdout: true, script: 'date +%Y-%m-%d').trim()
    def buildNumber = config.buildNumber ?: env.BUILD_NUMBER
    def bucket      = config.bucket      ?: env.MINIO_BUCKET    ?: env.JENKINS_MINIO_BUCKET ?: 'ci-artifacts'
    def endpoint    = config.endpoint    ?: env.MINIO_ENDPOINT  ?: env.JENKINS_MINIO_ENDPOINT ?: 'http://minio:9000'
    def credId      = config.credentialsId ?: 'minio-credentials'
    def localDir    = config.localDir    ?: 'build-artifacts'
    def artifactType = config.artifactType  // if set, only upload that sub-dir
    def mcImage     = config.mcImage     ?: 'minio/mc:RELEASE.2024-04-18T16-45-29Z'
    def network     = config.network     ?: env.TRAEFIK_NETWORK ?: env.JENKINS_TRAEFIK_NETWORK ?: 'home-net'

    def remoteBase  = "${app}/${component}/${buildDate}/${buildNumber}"
    def localSrc    = artifactType ? "${localDir}/${artifactType}" : localDir
    def remoteDest  = artifactType ? "${remoteBase}/${artifactType}" : remoteBase

    // Skip silently if the source directory is empty / doesn't exist
    def exists = sh(returnStatus: true, script: "test -d '${localSrc}' && find '${localSrc}' -type f | grep -q .")
    if (exists != 0) {
        echo "publishToMinio: '${localSrc}' is empty or missing — skipping upload"
        return
    }

    withCredentials([usernamePassword(
            credentialsId: credId,
            usernameVariable: 'MC_ACCESS_KEY',
            passwordVariable: 'MC_SECRET_KEY')]) {
        sh """
            docker run --rm \\
                --network ${network} \\
                -v \${PWD}/${localSrc}:/upload:ro \\
                -e MC_HOST_homeminio="http://\${MC_ACCESS_KEY}:\${MC_SECRET_KEY}@${endpoint.replaceAll('https?://', '')}" \\
                ${mcImage} \\
                sh -c "
                    mc mb --ignore-existing homeminio/${bucket} || true
                    mc cp --recursive /upload/ homeminio/${bucket}/${remoteDest}/
                    echo 'Upload complete: ${bucket}/${remoteDest}/'
                    mc ls homeminio/${bucket}/${remoteDest}/ 2>/dev/null || true
                "
        """
    }

    echo "Artifacts available at: ${endpoint}/${bucket}/${remoteDest}/"
}

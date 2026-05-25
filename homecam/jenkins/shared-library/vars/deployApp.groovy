// Shared library step: deployApp
// Usage:
//   deployApp app: 'homecam', environment: 'dev', kubeconfig: 'kubeconfig-homecam'
//   deployApp app: 'local-aistack', environment: 'prod', composeFile: 'compose/docker-compose.prod.yml'

def call(Map config = [:]) {
    def app         = config.app         ?: error('deployApp: app is required')
    def environment = config.environment ?: 'dev'
    def strategy    = config.strategy    ?: 'k8s'   // 'k8s' or 'compose'
    def namespace   = config.namespace   ?: app
    def imageTag    = config.imageTag    ?: env.DOCKER_TAG ?: env.GIT_COMMIT?.take(8) ?: 'latest'

    echo "Deploying ${app} to ${environment} (${strategy}) — image tag: ${imageTag}"

    if (strategy == 'k8s') {
        _deployK8s(app, environment, namespace, imageTag, config)
    } else if (strategy == 'compose') {
        _deployCompose(app, environment, imageTag, config)
    } else {
        error("deployApp: unknown strategy '${strategy}'. Use 'k8s' or 'compose'.")
    }
}

private void _deployK8s(String app, String env, String namespace, String imageTag, Map config) {
    def kubeconfigId = config.kubeconfigCredId ?: "kubeconfig-${app}"
    def k8sDir       = config.k8sDir ?: './k8s'

    withCredentials([file(credentialsId: kubeconfigId, variable: 'KUBECONFIG')]) {
        sh """
            export KUBECONFIG=\${KUBECONFIG}
            kubectl config current-context

            # Update image tags in all deployments
            for dep in \$(kubectl get deployments -n ${namespace} -o jsonpath='{.items[*].metadata.name}'); do
                kubectl set image deployment/\${dep} \${dep}=\${DOCKER_REGISTRY}/\${dep}:${imageTag} \
                    -n ${namespace} --record || true
            done

            # Apply any manifest changes
            kubectl apply -f ${k8sDir}/ -n ${namespace} --dry-run=client
            kubectl apply -f ${k8sDir}/ -n ${namespace}

            # Wait for rollout
            kubectl rollout status deployment --timeout=5m -n ${namespace}
        """
    }
}

private void _deployCompose(String app, String envName, String imageTag, Map config) {
    def composeFile  = config.composeFile ?: 'docker-compose.yml'
    def envFile      = config.envFile     ?: "env/.env.${envName}"
    def projectName  = config.projectName ?: app

    sh """
        export DOCKER_TAG=${imageTag}
        docker compose \\
            --project-name ${projectName} \\
            --file ${composeFile} \\
            --env-file ${envFile} \\
            pull

        docker compose \\
            --project-name ${projectName} \\
            --file ${composeFile} \\
            --env-file ${envFile} \\
            up -d --remove-orphans

        docker compose \\
            --project-name ${projectName} \\
            --file ${composeFile} \\
            --env-file ${envFile} \\
            ps
    """
}

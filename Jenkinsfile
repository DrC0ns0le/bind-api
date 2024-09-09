pipeline {
    agent {
        kubernetes {
            yamlFile 'kaniko-builder.yml'
        }
    }

    environment {
        DOCKER_REGISTRY = "registry.internal.leejacksonz.com"
        DOCKER_IMAGE = "bind-api"
        GIT_BRANCH_NAME = "${env.GIT_BRANCH.replaceAll('^origin/', '')}"
        DOCKER_TAG = "${GIT_BRANCH_NAME.replaceAll('/', '-')}-${env.GIT_COMMIT.take(7)}"
    }
    stages {
        stage('Build and Push Docker Image') {
            steps {
                container('kaniko') {
                    script {
                        def kanikoCmd = "/kaniko/executor --context . " +
                                        "--destination ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG}"
                        
                        if (GIT_BRANCH_NAME == 'main' || GIT_BRANCH_NAME == 'master') {
                            kanikoCmd += " --destination ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest"
                        }
                        
                        sh kanikoCmd
                    }
                }
            }
        }
    }
    post {
        always {
            cleanWs()
        }
    }
}
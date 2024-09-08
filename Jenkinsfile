pipeline {
    agent {
        kubernetes {
            yamlFile 'kaniko-builder.yml'
        }
    }

    environment {
        DOCKER_REGISTRY = "registry.internal.leejacksonz.com"
        DOCKER_IMAGE = "bind-api"
        DOCKER_TAG = "${env.GIT_BRANCH.replaceAll('/', '-')}-${env.GIT_COMMIT.take(7)}"
    }

    stages {
        stage('Build Docker Image') {
            steps {
                sh 'docker build -t ${DOCKER_REGISTRY}/${DOCKER_IMAGE_NAME}:${DOCKER_TAG} .'
            }
        }
        stage('Push to Registry') {
            steps {
                script {
                    sh "docker push ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG}"
                    
                    if (env.GIT_BRANCH == 'main' || env.GIT_BRANCH == 'master') {
                        sh "docker tag ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest"
                        sh "docker push ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest"
                    }
                }
            }
        }
    }

    post {
        always {
            cleanWs()
            sh "docker rmi ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG}"
            sh "docker rmi ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest || true"
        }
    }
}
podTemplate(yaml: '''
    apiVersion: v1
    kind: Pod
    spec:
      containers:
      - name: maven
        image: maven:3.8-openjdk-11
        command:
        - cat
        tty: true
      - name: kaniko
        image: gcr.io/kaniko-project/executor:debug
        command:
        - cat
        tty: true
        volumeMounts:
        - name: kaniko-secret
          mountPath: /kaniko/.docker
      restartPolicy: Never
      volumes:
      - name: kaniko-secret
        secret:
            secretName: dockercred
            items:
            - key: .dockerconfigjson
              path: config.json
''') {

    pipeline {
        agent {
            kubernetes {
                yaml podTemplate.yaml
            }
        }

        environment {
            DOCKER_REGISTRY = "registry.internal.leejacksonz.com"
            DOCKER_IMAGE = "bind-api"
            DOCKER_TAG = "${env.GIT_BRANCH.replaceAll('/', '-')}-${env.GIT_COMMIT.take(7)}"
        }

        stages {
            stage('Build and Test') {
                steps {
                    container('maven') {
                        sh 'mvn clean package'
                    }
                }
            }

            stage('Build and Push Docker Image') {
                steps {
                    container('kaniko') {
                        script {
                            def destinations = [
                                "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG}"
                            ]
                            
                            if (env.GIT_BRANCH == 'main' || env.GIT_BRANCH == 'master') {
                                destinations.add("${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest")
                            }
                            
                            sh """
                                /kaniko/executor --context . \
                                    --destination ${destinations.join(' --destination ')}
                            """
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
}
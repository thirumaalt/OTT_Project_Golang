// Root Jenkinsfile — myflix OTT monorepo
//
// Upgrades the per-service admin-dashboard pipeline into one monorepo pipeline:
//   - SERVICES = 'all'      -> build every service that has a Dockerfile
//   - SERVICES = 'changed'  -> build only services touched since the last commit
//   - SERVICES = 'a,b,c'    -> build an explicit comma-separated list
//
// Jenkins credentials required:
//   docker-hub-creds  (Username/Password) — Docker Hub login
//   sonar-token       (Secret text)       — SonarQube/SonarCloud token

pipeline {
    agent any

    parameters {
        string(name: 'SERVICES', defaultValue: 'changed',
               description: "'all', 'changed', or a comma-separated list e.g. auth-service,user-service")
        string(name: 'DOCKERFILE', defaultValue: 'Dockerfile',
               description: "Dockerfile to build (use Dockerfile.eks for ECR/distroless builds)")
        booleanParam(name: 'RUN_SECURITY_SCAN', defaultValue: true,  description: 'Run Trivy image scan')
        booleanParam(name: 'RUN_SONAR',         defaultValue: false, description: 'Run SonarQube analysis')
        booleanParam(name: 'PUSH_TO_REGISTRY',  defaultValue: false, description: 'Push images to Docker Hub')
    }

    environment {
        DOCKER_REGISTRY = 'thiru98'
        DOCKER_CREDS_ID = 'docker-hub-creds'
        SONAR_TOKEN     = credentials('sonar-token')
        GIT_REPO        = 'https://github.com/thirumaalt/OTT_Project_Golang.git'
    }

    stages {
        stage('Checkout') {
            steps {
                git branch: 'main', url: "${GIT_REPO}"
            }
        }

        stage('Resolve services') {
            steps {
                script {
                    def services = []
                    if (params.SERVICES == 'all') {
                        // every top-level dir that has a Dockerfile
                        services = sh(returnStdout: true, script:
                            "for d in */; do [ -f \"\${d}Dockerfile\" ] && echo \${d%/}; done"
                        ).trim().split('\n').findAll { it }
                    } else if (params.SERVICES == 'changed') {
                        def diff = sh(returnStdout: true, script:
                            "git diff --name-only HEAD~1 HEAD 2>/dev/null || true"
                        ).trim()
                        def changedDirs = diff ? diff.split('\n').collect { it.split('/')[0] }.unique() : []
                        services = changedDirs.findAll { fileExists("${it}/Dockerfile") }
                        if (services.isEmpty()) {
                            echo "No service Dockerfiles touched in the last commit — nothing to build."
                        }
                    } else {
                        services = params.SERVICES.split(',').collect { it.trim() }.findAll { it }
                    }
                    env.RESOLVED_SERVICES = services.join(' ')
                    echo "Services to build: ${env.RESOLVED_SERVICES ?: '(none)'}"
                }
            }
        }

        stage('Build · Scan · Push') {
            when { expression { return env.RESOLVED_SERVICES?.trim() } }
            steps {
                script {
                    def shortSha = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()
                    for (svc in env.RESOLVED_SERVICES.split(' ')) {
                        def image = "${DOCKER_REGISTRY}/${svc}"
                        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                        echo "🔨 ${svc}"
                        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

                        if (!fileExists("${svc}/${params.DOCKERFILE}")) {
                            error("Dockerfile not found: ${svc}/${params.DOCKERFILE}")
                        }

                        sh """
                            docker build -f ${svc}/${params.DOCKERFILE} \
                                -t ${image}:${BUILD_NUMBER} \
                                -t ${image}:${shortSha} \
                                -t ${image}:latest \
                                ${svc}
                        """

                        if (params.RUN_SECURITY_SCAN) {
                            echo "🔍 Trivy scan — ${svc}"
                            sh """
                                docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
                                    aquasecurity/trivy image \
                                    --severity HIGH,CRITICAL \
                                    --ignore-unfixed \
                                    --exit-code 0 \
                                    --format table \
                                    ${image}:${shortSha} || true
                            """
                        }

                        if (params.PUSH_TO_REGISTRY) {
                            withCredentials([usernamePassword(credentialsId: DOCKER_CREDS_ID,
                                usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
                                sh """
                                    echo \$PASSWORD | docker login -u \$USERNAME --password-stdin
                                    docker push ${image}:${BUILD_NUMBER}
                                    docker push ${image}:${shortSha}
                                    docker push ${image}:latest
                                """
                            }
                            echo "📤 Pushed ${image} (${shortSha}, ${BUILD_NUMBER}, latest)"
                        }
                    }
                }
            }
        }

        stage('SonarQube') {
            when { expression { return params.RUN_SONAR } }
            steps {
                echo "📊 SonarQube analysis (whole repo)"
                sh """
                    docker run --rm \
                        -e SONAR_HOST_URL=\"https://sonarcloud.io\" \
                        -e SONAR_TOKEN=${SONAR_TOKEN} \
                        -v ${WORKSPACE}:/usr/src \
                        sonarsource/sonar-scanner-cli
                """
            }
        }
    }

    post {
        always {
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "📋 BUILD SUMMARY"
            echo "   Build:    ${BUILD_NUMBER}"
            echo "   Services: ${env.RESOLVED_SERVICES ?: '(none)'}"
            echo "   Scan:     ${params.RUN_SECURITY_SCAN}   Sonar: ${params.RUN_SONAR}   Push: ${params.PUSH_TO_REGISTRY}"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            cleanWs()
        }
        success { echo "✅ Pipeline succeeded 🚀" }
        failure { echo "❌ Pipeline failed" }
    }
}
pipeline {
    agent { label 'pr' }

    options {
        timestamps()
        timeout(time: 4, unit: 'HOURS')
        disableConcurrentBuilds(abortPrevious: true)
    }

    stages {
        stage('Validate commit') {
            steps {
                script {
                    def CHANGE_REPO = sh(script: 'basename -s .git `git config --get remote.origin.url`', returnStdout: true).trim()
                    build job: '/Utils/Validate-Git-Commit', parameters: [
                        string(name: 'Repo', value: "${CHANGE_REPO}"),
                        string(name: 'Branch', value: "${env.CHANGE_BRANCH}"),
                        string(name: 'Commit', value: "${GIT_COMMIT}")
                    ]
                }
            }
        }

        stage('Build') {
            steps {
		sh 'git submodule update --init --recursive'
                sh 'make'
            }
        }

        stage('Run tests') {
            steps {
                sh 'go test ./...'
            }
        }

        stage('Clean up') {
            steps {
                sh 'make clean'
            }
        }
    }
}

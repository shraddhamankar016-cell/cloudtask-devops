// Jenkins declarative pipeline for CloudTask
// Runs on a Jenkins agent with Docker, AWS CLI, and kubectl installed
// (see scripts/setup-ec2-jenkins.sh for a ready-made provisioning script).

pipeline {
    agent any

    environment {
        AWS_REGION       = 'ap-south-1'
        AWS_ACCOUNT_ID   = credentials('aws-account-id')
        ECR_BACKEND_REPO = "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/cloudtask-backend"
        ECR_FRONTEND_REPO = "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/cloudtask-frontend"
        IMAGE_TAG        = "${env.BUILD_NUMBER}"
        EKS_CLUSTER_NAME = 'cloudtask-eks'
    }

    options {
        timestamps()
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '15'))
    }

    stages {

        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Backend: Test & Build') {
            steps {
                dir('backend') {
                    sh 'go test ./... -v'
                    sh 'go build -o /tmp/cloudtask-backend .'
                }
            }
        }

        stage('Build Docker Images') {
            steps {
                sh "docker build -t ${ECR_BACKEND_REPO}:${IMAGE_TAG} -t ${ECR_BACKEND_REPO}:latest ./backend"
                sh "docker build -t ${ECR_FRONTEND_REPO}:${IMAGE_TAG} -t ${ECR_FRONTEND_REPO}:latest ./frontend"
            }
        }

        stage('Push to Amazon ECR') {
            steps {
                withCredentials([[
                    $class: 'AmazonWebServicesCredentialsBinding',
                    credentialsId: 'aws-jenkins-creds'
                ]]) {
                    sh '''
                        aws ecr get-login-password --region $AWS_REGION | \
                        docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

                        docker push ${ECR_BACKEND_REPO}:${IMAGE_TAG}
                        docker push ${ECR_BACKEND_REPO}:latest
                        docker push ${ECR_FRONTEND_REPO}:${IMAGE_TAG}
                        docker push ${ECR_FRONTEND_REPO}:latest
                    '''
                }
            }
        }

        stage('Deploy to EKS') {
            when { branch 'main' }
            steps {
                withCredentials([[
                    $class: 'AmazonWebServicesCredentialsBinding',
                    credentialsId: 'aws-jenkins-creds'
                ]]) {
                    sh '''
                        aws eks update-kubeconfig --region $AWS_REGION --name $EKS_CLUSTER_NAME

                        kubectl apply -f k8s/namespace.yaml
                        kubectl apply -f k8s/configmap.yaml
                        kubectl apply -f k8s/secret.yaml
                        kubectl apply -f k8s/backend-deployment.yaml
                        kubectl apply -f k8s/backend-service.yaml
                        kubectl apply -f k8s/frontend-deployment.yaml
                        kubectl apply -f k8s/frontend-service.yaml
                        kubectl apply -f k8s/hpa.yaml
                        kubectl apply -f k8s/ingress.yaml

                        kubectl set image deployment/cloudtask-backend backend=${ECR_BACKEND_REPO}:${IMAGE_TAG} -n cloudtask
                        kubectl set image deployment/cloudtask-frontend frontend=${ECR_FRONTEND_REPO}:${IMAGE_TAG} -n cloudtask

                        kubectl rollout status deployment/cloudtask-backend -n cloudtask --timeout=120s
                        kubectl rollout status deployment/cloudtask-frontend -n cloudtask --timeout=120s
                    '''
                }
            }
        }
    }

    post {
        success {
            echo "✅ Build #${env.BUILD_NUMBER} deployed successfully."
        }
        failure {
            echo "❌ Build #${env.BUILD_NUMBER} failed. Check the stage logs above."
        }
        always {
            sh 'docker image prune -f || true'
        }
    }
}

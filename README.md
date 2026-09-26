# ☁️ CloudTask — End-to-End DevOps Project on AWS

[![CI/CD](https://github.com/YOUR_USERNAME/cloudtask-devops/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/YOUR_USERNAME/cloudtask-devops/actions/workflows/ci-cd.yml)
![Docker](https://img.shields.io/badge/Docker-multi--stage-2496ED?logo=docker&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-EKS-326CE5?logo=kubernetes&logoColor=white)
![Terraform](https://img.shields.io/badge/Terraform-IaC-7B42BC?logo=terraform&logoColor=white)
![Jenkins](https://img.shields.io/badge/Jenkins-Pipeline-D24939?logo=jenkins&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-Cloud-232F3E?logo=amazon-aws&logoColor=white)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)

**CloudTask** is a full-stack Task Management web application, built to showcase a **complete, production-style DevOps pipeline**: from a Linux-provisioned CI server, through containerization, infrastructure-as-code on AWS, orchestration with Kubernetes, and fully automated CI/CD with both **Jenkins** and **GitHub Actions**.

> Built as a portfolio project to demonstrate hands-on DevOps engineering skills: Linux, AWS, Docker, Kubernetes, Terraform, Jenkins and CI/CD — end to end, not just in theory.

---

## 🖼️ What it is

A Kanban-style task manager with:
- A **vanilla JS frontend** (To Do / In Progress / Done board)
- A **Go REST API** backend (standard library only — `net/http`, no external frameworks)
- Full **DevOps tooling** wrapped around it so the same app can be built, tested, containerized, provisioned, and deployed to AWS with a single `git push`

---

## 🏗️ Architecture

```
                         ┌─────────────────────────────────────────┐
                         │                Developer                 │
                         └───────────────────┬───────────────────────┘
                                              │ git push
                                              ▼
                 ┌────────────────────────────────────────────────────┐
                 │        CI/CD  (GitHub Actions  /  Jenkins)          │
                 │  1. go test (backend)                                │
                 │  2. docker build (frontend + backend)                │
                 │  3. docker push  → Amazon ECR / Docker Hub            │
                 │  4. kubectl apply + rollout → Amazon EKS               │
                 └───────────────────────┬────────────────────────────┘
                                          ▼
        ┌───────────────────────────── AWS Cloud (via Terraform) ─────────────────────────────┐
        │                                                                                        │
        │   VPC (public + private subnets, 2 AZs)                                                │
        │        │                                                                                │
        │        ├── Amazon EKS Cluster                                                          │
        │        │      ├── Deployment: cloudtask-frontend (Nginx, 2+ pods, HPA)                  │
        │        │      └── Deployment: cloudtask-backend  (Go API, 2+ pods, HPA)                 │
        │        │              ▲ liveness/readiness probes, resource limits                      │
        │        ├── Application Load Balancer (Ingress)  → routes / and /api                     │
        │        ├── Amazon ECR  (backend & frontend image repositories)                          │
        │        └── Amazon RDS (Postgres, private subnet — ready for persistence)                │
        │                                                                                          │
        └──────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🧰 Tech stack

| Layer                | Tools / Services                                              |
|-----------------------|----------------------------------------------------------------|
| Frontend              | HTML5, CSS3, vanilla JavaScript (fetch API), Nginx             |
| Backend               | Go (standard library `net/http`), built-in `testing` + `httptest` |
| Containerization      | Docker (multi-stage builds), Docker Compose                    |
| Orchestration         | Kubernetes (Amazon EKS) — Deployments, Services, Ingress, HPA  |
| Infrastructure as Code| Terraform (VPC, EKS, ECR, RDS, IAM)                            |
| CI/CD                 | Jenkins (declarative pipeline) + GitHub Actions                |
| Cloud                 | AWS (EKS, ECR, RDS, VPC, IAM, ALB)                              |
| OS / Provisioning     | Linux (Amazon Linux / Ubuntu) bash provisioning scripts         |

---

## 📁 Project structure

```
cloudtask-devops/
├── frontend/                    # Static UI served by Nginx
│   ├── index.html
│   ├── style.css
│   ├── script.js
│   ├── config.template.js       # runtime-injected API config
│   ├── nginx.conf.template
│   ├── docker-entrypoint.sh
│   └── Dockerfile
├── backend/                     # Go REST API (standard library only)
│   ├── main.go                  # router, middleware, server startup
│   ├── handlers.go              # HTTP handlers for the tasks API
│   ├── store.go                 # thread-safe in-memory task store
│   ├── main_test.go             # unit tests (net/http/httptest)
│   ├── go.mod
│   └── Dockerfile
├── k8s/                         # Kubernetes manifests
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── backend-deployment.yaml
│   ├── backend-service.yaml
│   ├── frontend-deployment.yaml
│   ├── frontend-service.yaml
│   ├── ingress.yaml
│   └── hpa.yaml
├── terraform/                   # AWS infrastructure as code
│   ├── provider.tf
│   ├── variables.tf
│   ├── vpc.tf
│   ├── eks.tf
│   ├── ecr.tf
│   ├── rds.tf
│   ├── outputs.tf
│   └── terraform.tfvars.example
├── scripts/                      # Linux automation
│   ├── setup-ec2-jenkins.sh      # provisions a Jenkins CI server on EC2
│   ├── bootstrap-tf-backend.sh   # creates S3 + DynamoDB for TF remote state
│   └── deploy.sh                 # manual build → push → deploy helper
├── .github/workflows/ci-cd.yml   # GitHub Actions pipeline
├── Jenkinsfile                   # Jenkins declarative pipeline
├── docker-compose.yml            # local full-stack dev environment
└── README.md
```

---

## 🚀 Run it locally

### Option 1 — Docker Compose (recommended, one command)

```bash
git clone https://github.com/YOUR_USERNAME/cloudtask-devops.git
cd cloudtask-devops
docker compose up --build
```

- Frontend → http://localhost:3000
- Backend API → http://localhost:8080/api/tasks

### Option 2 — Run services individually

```bash
# Backend (requires Go 1.22+)
cd backend
go run .            # http://localhost:8080

# Frontend (in a new terminal)
cd frontend
npx serve .         # or open index.html directly
```

### Run backend tests

```bash
cd backend
go test ./... -v
```

---

## ☸️ Deploy to Kubernetes (Amazon EKS)

### 1. Provision AWS infrastructure with Terraform

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars   # fill in db_password etc.
terraform init
terraform plan
terraform apply
```

This creates: a VPC with public/private subnets across 2 AZs, an EKS cluster with a managed node group, two ECR repositories, and an RDS Postgres instance.

### 2. Point kubectl at the new cluster

```bash
aws eks update-kubeconfig --region ap-south-1 --name cloudtask-eks
```

### 3. Build, push and deploy

```bash
./scripts/deploy.sh <your-aws-account-id> ap-south-1 v1.0.0
```

This builds both Docker images, pushes them to ECR, applies every manifest in `k8s/`, and rolls out the new versions with zero-downtime rolling updates.

### 4. Check it's running

```bash
kubectl get pods -n cloudtask
kubectl get ingress -n cloudtask
```

---

## 🔁 CI/CD pipelines

Two equivalent pipelines are included so the project demonstrates both a **cloud-native** and a **self-hosted / traditional** DevOps workflow:

### GitHub Actions (`.github/workflows/ci-cd.yml`)
Runs automatically on every push: run Go unit tests → build both Docker images → push to Docker Hub → (on `main`) deploy to EKS.
Add these repository secrets under **Settings → Secrets and variables → Actions**:
`DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`.

### Jenkins (`Jenkinsfile`)
A self-hosted pipeline for a more traditional enterprise DevOps setup: checkout → test → build → push to Amazon ECR → deploy to EKS.
Spin up a Jenkins server in one command using the provided Linux provisioning script:

```bash
./scripts/setup-ec2-jenkins.sh
```

This installs Docker, Go, AWS CLI v2, kubectl, Terraform, Java and Jenkins itself on a fresh EC2 instance. Then create a Pipeline job in Jenkins pointing at this repo — it will pick up the `Jenkinsfile` automatically.

---

## ✅ What this project demonstrates

- **Linux**: bash scripting for server provisioning (Jenkins/CI server setup, Terraform backend bootstrap)
- **AWS**: VPC design, EKS, ECR, RDS, IAM roles/policies, Application Load Balancer — all defined in Terraform
- **Docker**: multi-stage builds, non-root containers, health checks, Docker Compose for local dev
- **Kubernetes**: Deployments with rolling updates, readiness/liveness probes, resource requests/limits, ConfigMaps/Secrets, HPA autoscaling, Ingress
- **Terraform**: modular, reusable infrastructure-as-code with remote state support
- **Jenkins & GitHub Actions**: two parallel, production-style CI/CD pipelines
- **Full-stack development**: a working REST API and UI wired together, not just static boilerplate

---

## 🗺️ Possible next steps

- Swap the in-memory store (`backend/store.go`) for the already-provisioned RDS Postgres instance
- Add Prometheus + Grafana for cluster/app monitoring
- Add Helm charts as an alternative to raw manifests
- Add SonarQube / Trivy stages to the pipelines for code quality & container scanning

---

## 👩‍💻 Author

**Shraddha Mankar**
This project was built to demonstrate practical, end-to-end DevOps engineering — feel free to fork it, star it, or reach out with questions.

## 📄 License

Licensed under the [MIT License](LICENSE).

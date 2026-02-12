# Thunder Deployment Patterns

This directory contains comprehensive guides for deploying WSO2 Thunder across different environments and platforms. Choose the deployment pattern that best fits your infrastructure and requirements.

## Available Deployment Options

### 🐳 [Docker Deployment](docker/docker-deployment.md)

Deploy Thunder using Docker containers for development and production environments.

**Best for:**
- Local development and testing
- Containerized applications
- Simple production setups
- CI/CD pipelines

**Key Features:**
- Quick setup with pre-built images
- PostgreSQL integration
- Custom configuration mounting
- Container orchestration ready

---

### ☸️ [Kubernetes Deployment](kubernetes/kubernetes-deployment.md)

Deploy Thunder on Kubernetes clusters using Helm charts for scalable, production-ready deployments.

**Best for:**
- Production environments
- Auto-scaling requirements
- High availability setups
- Cloud-native architectures

**Key Features:**
- Helm chart deployment
- Multi-replica support
- Ingress configuration
- Database flexibility (PostgreSQL/SQLite)
- Rolling updates and rollbacks

---

### 🔄 [OpenChoreo Deployment](openchoreo/openchoreo-deployment.md)

Deploy Thunder on OpenChoreo platform for advanced orchestration and cell-based architecture.

**Best for:**
- Microservices architectures
- Cell-based deployments
- Advanced orchestration needs
- Platform abstraction

**Key Features:**
- Cell-based deployment model
- Integrated platform services
- Advanced networking
- Service mesh integration

## Getting Started

### For Docker
Start with **[Docker Deployment](docker/docker-deployment.md)** for the quickest way to get Thunder running locally.

### For Kubernetes
Choose **[Kubernetes Deployment](kubernetes/kubernetes-deployment.md)** for robust, scalable production deployments.

### For OpenChoreo
Explore **[OpenChoreo Deployment](openchoreo/openchoreo-deployment.md)** for sophisticated platform-based deployments.

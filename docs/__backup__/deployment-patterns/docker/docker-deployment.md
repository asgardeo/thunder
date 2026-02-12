# Docker Deployment Guide

This guide provides comprehensive instructions for deploying Thunder using Docker, covering everything from local development to production deployments with Docker Compose.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Database Setup](#database-setup)

## Prerequisites

### Infrastructure Requirements

- **Docker Engine**: Version 20.10+ (Check with `docker --version`)
- **Docker Compose**: Version 2.0+ (Check with `docker-compose --version`)
- **Network**: Access to required ports (8090 for Thunder, 5432 for PostgreSQL)

### Required Tools

| Tool          | Installation Guide | Version Check Command|
|---------------|--------------------|-----------------------|
| Docker        | [Install Docker](https://docs.docker.com/engine/install/) | `docker --version` |
| Docker Compose| [Install Docker Compose](https://docs.docker.com/compose/install/) | `docker-compose --version` |
| Git           | [Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) | `git --version` |

### Verify Prerequisites

```bash
# Check Docker installation
docker --version
docker-compose --version

# Test Docker daemon
docker run hello-world

# Check available system resources
docker system df
```

## Quick Start

Deploy Thunder with Docker in under 5 minutes:

### Option 1: Run with Docker

Follow these steps to run Thunder using Docker.

1. **Pull the Docker image**

    ```bash
    docker pull ghcr.io/asgardeo/thunder:latest
    ```

2. **Setup the product**

    You need to setup the server with the initial configurations and data before starting the server for the first time.

    ```bash
        docker run -it --rm \
            ghcr.io/asgardeo/thunder:latest \
            ./setup.sh
    ```

    > [!NOTE]
    > This will shut down the container after the setup is complete. You need to start the container again using the command in step 3. If you are using sqlite as the database, then you need to mount a volume to persist the database file and share it between the setup and server run containers.

3. **Run the container**

    ```bash
    docker run --rm \
      -p 8090:8090 \
      ghcr.io/asgardeo/thunder:latest
    ```

    Optionally if you want to modify the server configurations, you can mount a custom `deployment.yaml` file. Create a `deployment.yaml` file in your working directory similar to the [deployment.yaml](https://github.com/asgardeo/thunder/blob/main/backend/cmd/server/repository/conf/deployment.yaml), and mount it as below:

    ```bash
    docker run --rm \
      -p 8090:8090 \
      -v $(pwd)/deployment.yaml:/opt/thunder/repository/conf/deployment.yaml \
      ghcr.io/asgardeo/thunder:latest
    ```

    Optionally if you want to use custom configurations or certificates, you can mount them as follows:

    ```bash
    docker run --rm \
      -p 8090:8090 \
      -v $(pwd)/deployment.yaml:/opt/thunder/repository/conf/deployment.yaml \
      -v $(pwd)/certs/server.cert:/opt/thunder/repository/resources/security/server.cert \
      -v $(pwd)/certs/server.key:/opt/thunder/repository/resources/security/server.key \
      ghcr.io/asgardeo/thunder:latest
    ```

### Access Thunder

- **Application**: https://localhost:8090
- **Gate (Login/Register)**: https://localhost:8090/signin
- **Develop (Admin Console)**: https://localhost:8090/develop

## Database Setup

### Option 1: Embedded SQLite (Default)

Thunder uses SQLite by default for development - no additional setup needed.

### Option 2: External PostgreSQL

For production deployments with PostgreSQL:

#### Step 1: Start and Initialize PostgreSQL

1. Navigate to local-development directory

```bash
cd install/local-development
```

2. Start PostgreSQL Database in background

```bash
docker compose up -d 
```

3. View PostgreSQL Database logs

```bash
docker compose logs -f
```

4. Stop PostgreSQL Database

```bash
docker compose down
```

- Stop PostgreSQL Database and delete all data 

```bash
docker compose down -v
```

# Thunder Quick Start

Run Thunder locally using Docker Compose. This is the fastest way to get Thunder up and running with all dependencies configured.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose installed
- Terminal access

---

## Running Thunder

Start all services with a single command:

```bash
docker compose up
```

This will automatically:
1. **Initialize** the database from the image
2. **Run setup** — bootstraps default resources (admin user, sample apps, etc.)
3. **Start the server** — Thunder is ready to serve requests

Once running, Thunder is available at:

| URL | Description |
|---|---|
| `https://localhost:8090` | Thunder Server |
| `https://localhost:8090/develop` | Developer Console |

> **Default credentials:** `admin` / `admin`

---

## Services

The Compose file defines three services:

| Service | Description |
|---|---|
| `thunder-db-init` | One-shot container that copies the initial database files to a shared volume |
| `thunder-setup` | One-shot container that bootstraps default resources via `setup.sh` |
| `thunder` | The Thunder server — starts after setup completes |

The `thunder-db-init` and `thunder-setup` services run once and exit. Only `thunder` stays running.

---

## Stopping Thunder

```bash
# Stop and keep data
docker compose down

# Stop and remove all data (fresh start)
docker compose down -v
```

---

## Custom Host and Port

By default, Thunder runs on `localhost:8090`. To run it on a different hostname or port — for example a custom domain, a server IP, or a local alias like `thunder.local` — you need to override three configuration files.

### How It Works

The Docker image bakes in default configuration files. You can override them with volume mounts — no image rebuild required.

| File in container | Purpose |
|---|---|
| `/opt/thunder/repository/conf/deployment.yaml` | Backend server — bind address, public URL, CORS, Gate client redirect |
| `/opt/thunder/apps/develop/config.js` | Developer Console frontend |
| `/opt/thunder/apps/gate/config.js` | Gate login app frontend |

### Step 1: Create Your Configuration Files

Create the following three files in the same directory as `docker-compose.yml`:

```text
.
├── docker-compose.yml
├── deployment.yaml       ← backend configuration
├── develop-config.js     ← Developer Console configuration
└── gate-config.js        ← Gate login app configuration
```

#### `deployment.yaml`

```yaml
server:
  hostname: "0.0.0.0"                            # Keep as-is — binds to all interfaces
  port: <your-port>                              # e.g. 8090
  public_url: "https://<your-host>:<your-port>" # e.g. https://thunder.local:8090

gate_client:
  hostname: "<your-host>"
  port: <your-port>
  scheme: "https"
  path: "/gate"

cors:
  allowed_origins:
    - "https://<your-host>:<your-port>"  # e.g. https://thunder.local:8090

passkey:
  allowed_origins:
    - "https://<your-host>:<your-port>"  # e.g. https://thunder.local:8090

# Other configurations...
```

#### `develop-config.js`

```js
window.__THUNDER_RUNTIME_CONFIG__ = {
  client: {
    base: '/develop',
    client_id: 'DEVELOP',
    scopes: ['openid', 'profile', 'email', 'system'],
  },
  server: {
    public_url: 'https://<your-host>:<your-port>', // e.g. https://thunder.local:8090
  },
};
```

#### `gate-config.js`

```js
window.__THUNDER_RUNTIME_CONFIG__ = {
  client: {
    base: '/gate',
  },
  server: {
    public_url: 'https://<your-host>:<your-port>', // e.g. https://thunder.local:8090
  },
};
```

### Step 2: Add Volume Mounts to `docker-compose.yml`

Add the following volume mounts to the `thunder-setup` and `thunder` services:

```yaml
services:
  thunder-setup:
    # ...
    volumes:
      # ...
      - ./deployment.yaml:/opt/thunder/repository/conf/deployment.yaml:ro

  thunder:
    # ...
    ports:
      - "<your-port>:<your-port>"  # Update if changing the port, e.g. 9090:9090
    volumes:
      # ...
      - ./deployment.yaml:/opt/thunder/repository/conf/deployment.yaml:ro
      - ./develop-config.js:/opt/thunder/apps/develop/config.js:ro
      - ./gate-config.js:/opt/thunder/apps/gate/config.js:ro
```

> **Note:** `deployment.yaml` must be mounted into `thunder-setup` too, because the setup process starts a temporary server to bootstrap resources. The frontend `config.js` files only need to be in the `thunder` service.
>
> The `ports` mapping only needs updating if you change the port number. If you are only changing the hostname, leave it as `8090:8090`.

### Step 3: Start Thunder

```bash
docker compose up
```

### Example: Using `thunder.local`

First, add the alias to your hosts file:

**macOS / Linux:**
```bash
echo "127.0.0.1 thunder.local" | sudo tee -a /etc/hosts
```

**Windows (run as Administrator):**
```powershell
Add-Content -Path "C:\Windows\System32\drivers\etc\hosts" -Value "127.0.0.1 thunder.local"
```

Then replace `<your-host>` with `thunder.local` and `<your-port>` with `8090` (or your chosen port) in all three configuration files.

---

## Troubleshooting

**`yaml: unmarshal errors` on startup**
Your `deployment.yaml` contains an unrecognized field. Ensure the config schema matches the image version you are running.

**Frontend still redirects to `localhost` or the wrong port**
Make sure all three files are mounted correctly. A hard refresh (`Ctrl+Shift+R`) may be needed to clear the browser cache.

**CORS errors in the browser**
Ensure your full origin (host + port) is listed under `cors.allowed_origins` in `deployment.yaml`.

**Connection refused on the new port**
Ensure the `ports` mapping in `docker-compose.yml` matches the port set in `deployment.yaml`.

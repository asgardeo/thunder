# Thunder - Gate App

This is the gate app for project thunder. Which serves UIs for Login, Registration and Recovery. 

### ✅ Prerequisites

- Node.js 20+
- PNPM 10+

---

### 🛠 Step 1: Install Dependencies

```bash
pnpm i
```

### 🔐 Step 2: Generate SSL Certificates

```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.cert -days 365 -nodes -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
```

### ▶️ Step 3: Run the Application

```bash
pnpm --filter gate dev
```

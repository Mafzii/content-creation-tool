# Content Creation Tool

A local tool for drafting content with LLM assistance. Manage topics, sources, and writing styles, then generate and tweak draft variants before saving.

## Stack

- **Backend** — Go, SQLite
- **Frontend** — Single-page HTML/JS (no build step)
- **LLM** — Pluggable: Gemini, Claude, or OpenAI

## Getting Started

**Prerequisites:** Go 1.22+

```bash
cd backend
go build -o server ./cmd/api
./server
```

Open `http://localhost:8080` in your browser.

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DATABASE_PATH` | `data/app.db` | SQLite database path |

Set them in a `.env` file in the `backend/` directory or export them directly.

### LLM Setup

Configure your provider in the **Settings** tab after launching. Paste your API key, choose a provider (Gemini / Claude / OpenAI), and set the model name.

## Usage

1. **Topics** — Create a topic with a description and keywords to scope generation.
2. **Sources** — Add reference material (paste text or a URL).
3. **Styles** — Define a writing style with a tone and prompt.
4. **Drafts** — Pick a topic, style, and sources, then click **Generate 3 Drafts** to get variants. Tweak a selected variant in the modal, then use it as your draft content.

## MCP Server

The `mcp/` directory contains a Model Context Protocol server that wraps the REST API, letting AI assistants (like Claude Code) manage topics, sources, styles, and drafts directly through tool calls.

**Prerequisites:** [uv](https://docs.astral.sh/uv/) (Python package manager)

```bash
cd mcp
uv sync
```

The server runs over stdio and is configured via `.mcp.json` at the project root. To use it with Claude Code, the backend must be running first.

| Variable | Default | Description |
|---|---|---|
| `CONTENT_TOOL_URL` | `http://localhost:8080` | Backend API base URL |

### Available Tools

| Tool | Description |
|---|---|
| `list/create/update/delete_topic` | Manage content topics |
| `list/create/update/delete_source` | Manage reference sources (text or URL) |
| `refetch_source` | Re-fetch content for a URL source |
| `list/create/update/delete_style` | Manage writing styles |
| `list/create/update/delete_draft` | Manage drafts |
| `generate_drafts` | Generate 3 draft variants from a topic + style |
| `tweak_draft` | Revise content with a natural language instruction |
| `get/save_settings` | Configure LLM provider, model, and API key |

## Deploying to Your Server (Tailscale + CI/CD)

This section explains how to run the tool on a home server or VPS and access it from your phone — even when you're away from home — using [Tailscale](https://tailscale.com).

### Prerequisites

| Tool | Why |
|---|---|
| [Docker](https://docs.docker.com/engine/install/) + [Docker Compose v2](https://docs.docker.com/compose/install/) | Container runtime |
| [Tailscale](https://tailscale.com/download) | Secure overlay network (phone ↔ server) |
| Git | Code checkout on the server |

### 1 — One-time server setup

```bash
# On your server
git clone https://github.com/Mafzii/content-creation-tool.git
cd content-creation-tool

# Create a backend/.env from the template (edit values as needed)
cp .env.example backend/.env

# Start the stack
./scripts/deploy.sh
```

The app will be available at `http://<server-tailscale-ip>:8080` from any device on your Tailscale network.

> **Tip:** Enable [MagicDNS](https://tailscale.com/kb/1081/magicdns) in your Tailscale admin so you can reach the server by name (e.g. `http://myserver:8080`) instead of an IP.

### 2 — Automated deploys via GitHub Actions (CI/CD)

Every push to `main` automatically deploys to your server. Add these secrets in **GitHub → Settings → Secrets → Actions**:

| Secret | Description |
|---|---|
| `DEPLOY_HOST` | Tailscale IP or MagicDNS hostname of your server |
| `DEPLOY_USER` | SSH username on the server |
| `DEPLOY_SSH_KEY` | Private SSH key (the server must have the public key in `~/.ssh/authorized_keys`) |
| `DEPLOY_PORT` | SSH port (default `22`) |
| `DEPLOY_PATH` | Absolute path to the cloned repo on the server (default `~/content-creation-tool`) |

**Workflow:**
1. You push a commit (from your laptop, phone via GitHub, or any editor).
2. GitHub Actions runs tests → builds the Docker image → SSHes into your server → pulls & restarts the container.
3. Open the app on your phone; the new version is live.

### 3 — Managing the running stack

```bash
# On the server
./scripts/deploy.sh          # pull latest & restart
./scripts/deploy.sh logs     # follow live logs
./scripts/deploy.sh status   # show container status
./scripts/deploy.sh stop     # shut everything down
```

### 4 — Accessing the app on your phone

1. Install [Tailscale](https://tailscale.com/download) on your phone and sign in.
2. Open `http://<server-name>:8080` in your phone's browser.
3. Bookmark it — the URL never changes because Tailscale keeps the overlay IP stable.

The app is a responsive SPA and works on mobile browsers without any native app required.

---

## Project Structure

```
backend/
  cmd/api/        # Entry point
  internal/
    config/       # Env config
    db/           # SQLite connection + migrations
    handlers/     # HTTP handlers (generic CRUD + generate)
    llm/          # LLM provider adapters
    models/       # Shared types
    store/        # SQLite store implementations
  migrations/     # SQL migration files
frontend/
  index.html      # Single-file SPA
  style.css       # Styles
  js/             # Modular JS (app, tabs, forms, api, etc.)
mcp/
  content_tool_mcp.py  # MCP server (FastMCP + httpx)
  pyproject.toml       # Python project config
  uv.lock              # Dependency lockfile
scripts/
  deploy.sh            # Server-side deploy helper
.github/workflows/
  ci.yml               # Build & test on every PR
  deploy.yml           # Auto-deploy to server on push to main
Dockerfile             # Multi-stage Go + frontend image
docker-compose.yml     # Local development
docker-compose.prod.yml # Production overrides
.env.example           # Environment variable template
```

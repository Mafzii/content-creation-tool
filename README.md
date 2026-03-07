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

## Running with Docker

```bash
# Create a local .env from the template
cp .env.example backend/.env

# Build and start
docker compose up --build
```

The app will be available at `http://localhost:8080`.

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
.github/workflows/
  ci.yml               # Build & test on every PR
Dockerfile             # Multi-stage Go + frontend image
docker-compose.yml     # Local development
.env.example           # Environment variable template
```

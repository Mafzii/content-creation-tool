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
```

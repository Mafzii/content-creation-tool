"""MCP server wrapping the content creation tool REST API."""

import os
from typing import Optional

import httpx
from mcp.server.fastmcp import FastMCP

BASE_URL = os.environ.get("CONTENT_TOOL_URL", "http://localhost:8080")

mcp = FastMCP(
    "content-tool",
    instructions=(
        "Content creation tool for managing topics, sources, writing styles, "
        "and drafts. Use these tools to create blog posts, articles, and other "
        "content. Typical workflow: create a topic, add sources, pick a style, "
        "then generate drafts. You can tweak drafts with natural language instructions."
    ),
)

client = httpx.AsyncClient(base_url=BASE_URL, timeout=120.0)


async def _request(method: str, path: str, **kwargs) -> dict | list | str:
    resp = await client.request(method, path, **kwargs)
    if resp.status_code >= 400:
        return f"Error {resp.status_code}: {resp.text}"
    if resp.status_code == 204 or not resp.content:
        return "OK"
    return resp.json()


# ── Topics ──────────────────────────────────────────────────────────────────


@mcp.tool()
async def list_topics() -> dict | list | str:
    """List all content topics.

    Returns all topics with their id, name, description, and keywords.
    Use this to see what topics exist before creating or generating drafts.
    """
    return await _request("GET", "/topics")


@mcp.tool()
async def create_topic(
    name: str,
    description: str = "",
    keywords: str = "",
) -> dict | list | str:
    """Create a new content topic.

    A topic defines what the content is about. Give it a clear name and
    optionally a description and comma-separated keywords for context.

    Args:
        name: Topic name, e.g. "AI Safety" or "Python Best Practices"
        description: Longer description of what the topic covers
        keywords: Comma-separated keywords for additional context
    """
    return await _request("POST", "/topics", json={
        "name": name,
        "description": description,
        "keywords": keywords,
    })


@mcp.tool()
async def update_topic(
    topic_id: int,
    name: Optional[str] = None,
    description: Optional[str] = None,
    keywords: Optional[str] = None,
) -> dict | list | str:
    """Update an existing topic.

    Args:
        topic_id: ID of the topic to update
        name: New topic name
        description: New description
        keywords: New comma-separated keywords
    """
    current = await _request("GET", f"/topics/{topic_id}")
    if isinstance(current, str):
        return current
    if name is not None:
        current["name"] = name
    if description is not None:
        current["description"] = description
    if keywords is not None:
        current["keywords"] = keywords
    return await _request("PUT", f"/topics/{topic_id}", json=current)


@mcp.tool()
async def delete_topic(topic_id: int) -> dict | list | str:
    """Delete a topic by ID.

    Args:
        topic_id: ID of the topic to delete
    """
    return await _request("DELETE", f"/topics/{topic_id}")


# ── Sources ─────────────────────────────────────────────────────────────────


@mcp.tool()
async def list_sources() -> dict | list | str:
    """List all content sources.

    Returns all sources with their id, name, type (text/url/file), raw input,
    extracted content, and status. Sources provide reference material for
    draft generation.
    """
    return await _request("GET", "/sources")


@mcp.tool()
async def create_source(
    name: str,
    type: str,
    raw: str,
    content: str = "",
) -> dict | list | str:
    """Create a new content source (text or URL).

    Sources provide reference material that gets fed into draft generation.
    - For type "text": raw is the text content itself
    - For type "url": raw is the URL to fetch content from (fetched async)

    Args:
        name: Display name for the source
        type: Either "text" or "url"
        raw: The text content or URL
        content: Pre-extracted content (optional, mainly for text sources)
    """
    return await _request("POST", "/sources", json={
        "name": name,
        "type": type,
        "raw": raw,
        "content": content,
    })


@mcp.tool()
async def update_source(
    source_id: int,
    name: Optional[str] = None,
    type: Optional[str] = None,
    raw: Optional[str] = None,
    content: Optional[str] = None,
    status: Optional[str] = None,
) -> dict | list | str:
    """Update an existing source.

    Args:
        source_id: ID of the source to update
        name: New display name
        type: New type (text/url)
        raw: New raw content or URL
        content: New extracted content
        status: New status (ready/pending/error)
    """
    current = await _request("GET", f"/sources/{source_id}")
    if isinstance(current, str):
        return current
    if name is not None:
        current["name"] = name
    if type is not None:
        current["type"] = type
    if raw is not None:
        current["raw"] = raw
    if content is not None:
        current["content"] = content
    if status is not None:
        current["status"] = status
    return await _request("PUT", f"/sources/{source_id}", json=current)


@mcp.tool()
async def delete_source(source_id: int) -> dict | list | str:
    """Delete a source by ID.

    Args:
        source_id: ID of the source to delete
    """
    return await _request("DELETE", f"/sources/{source_id}")


@mcp.tool()
async def refetch_source(source_id: int) -> dict | list | str:
    """Re-fetch content for a URL source.

    Triggers an async re-fetch of the URL. The source status will change to
    "pending" and then "ready" (or "error") once the fetch completes.
    Use list_sources or check back to see the updated status.

    Args:
        source_id: ID of the URL source to re-fetch
    """
    return await _request("POST", f"/sources/{source_id}/fetch")


# ── Styles ──────────────────────────────────────────────────────────────────


@mcp.tool()
async def list_styles() -> dict | list | str:
    """List all writing styles.

    Returns all styles with their id, name, prompt, tone, and example.
    Styles control the voice and format of generated drafts.
    """
    return await _request("GET", "/styles")


@mcp.tool()
async def create_style(
    name: str,
    prompt: str,
    tone: str = "",
    example: str = "",
) -> dict | list | str:
    """Create a new writing style.

    A style controls how generated content sounds and is structured.

    Args:
        name: Style name, e.g. "Casual Blog" or "Technical Documentation"
        prompt: Instructions for the LLM on how to write, e.g. "Write in a
            conversational, friendly tone with short paragraphs"
        tone: Short tone descriptor, e.g. "casual", "formal", "technical"
        example: Example text showing the desired writing style
    """
    return await _request("POST", "/styles", json={
        "name": name,
        "prompt": prompt,
        "tone": tone,
        "example": example,
    })


@mcp.tool()
async def update_style(
    style_id: int,
    name: Optional[str] = None,
    prompt: Optional[str] = None,
    tone: Optional[str] = None,
    example: Optional[str] = None,
) -> dict | list | str:
    """Update an existing writing style.

    Args:
        style_id: ID of the style to update
        name: New style name
        prompt: New writing instructions
        tone: New tone descriptor
        example: New example text
    """
    current = await _request("GET", f"/styles/{style_id}")
    if isinstance(current, str):
        return current
    if name is not None:
        current["name"] = name
    if prompt is not None:
        current["prompt"] = prompt
    if tone is not None:
        current["tone"] = tone
    if example is not None:
        current["example"] = example
    return await _request("PUT", f"/styles/{style_id}", json=current)


@mcp.tool()
async def delete_style(style_id: int) -> dict | list | str:
    """Delete a writing style by ID.

    Args:
        style_id: ID of the style to delete
    """
    return await _request("DELETE", f"/styles/{style_id}")


# ── Drafts ──────────────────────────────────────────────────────────────────


@mcp.tool()
async def list_drafts() -> dict | list | str:
    """List all drafts.

    Returns all drafts with their id, title, content, topic_id, style_id,
    status, notes, and source_ids. Drafts are the final content pieces.
    """
    return await _request("GET", "/drafts")


@mcp.tool()
async def create_draft(
    title: str,
    content: str,
    topic_id: int,
    style_id: int,
    status: str = "draft",
    notes: str = "",
    source_ids: Optional[list[int]] = None,
) -> dict | list | str:
    """Save a new draft.

    Use this after generating content to save a chosen variant as a draft.

    Args:
        title: Draft title
        content: The full draft content/body
        topic_id: ID of the topic this draft is about
        style_id: ID of the style used
        status: Draft status, e.g. "draft", "review", "published"
        notes: Internal notes about this draft
        source_ids: List of source IDs used as references
    """
    return await _request("POST", "/drafts", json={
        "title": title,
        "content": content,
        "topic_id": topic_id,
        "style_id": style_id,
        "status": status,
        "notes": notes,
        "source_ids": source_ids or [],
    })


@mcp.tool()
async def update_draft(
    draft_id: int,
    title: Optional[str] = None,
    content: Optional[str] = None,
    topic_id: Optional[int] = None,
    style_id: Optional[int] = None,
    status: Optional[str] = None,
    notes: Optional[str] = None,
    source_ids: Optional[list[int]] = None,
) -> dict | list | str:
    """Update an existing draft.

    Args:
        draft_id: ID of the draft to update
        title: New title
        content: New content body
        topic_id: New topic ID
        style_id: New style ID
        status: New status (draft/review/published)
        notes: New notes
        source_ids: New list of source IDs
    """
    current = await _request("GET", f"/drafts/{draft_id}")
    if isinstance(current, str):
        return current
    if title is not None:
        current["title"] = title
    if content is not None:
        current["content"] = content
    if topic_id is not None:
        current["topic_id"] = topic_id
    if style_id is not None:
        current["style_id"] = style_id
    if status is not None:
        current["status"] = status
    if notes is not None:
        current["notes"] = notes
    if source_ids is not None:
        current["source_ids"] = source_ids
    return await _request("PUT", f"/drafts/{draft_id}", json=current)


@mcp.tool()
async def delete_draft(draft_id: int) -> dict | list | str:
    """Delete a draft by ID.

    Args:
        draft_id: ID of the draft to delete
    """
    return await _request("DELETE", f"/drafts/{draft_id}")


# ── Generation ──────────────────────────────────────────────────────────────


@mcp.tool()
async def generate_drafts(
    topic_id: int,
    style_id: int,
    source_ids: Optional[list[int]] = None,
    notes: str = "",
) -> dict | list | str:
    """Generate 3 draft variants from a topic, style, and optional sources.

    This calls the LLM to produce 3 different content variants based on the
    specified topic, writing style, and source material. After reviewing the
    variants, use create_draft to save your preferred one.

    Requires LLM settings to be configured (use get_settings / save_settings).

    Args:
        topic_id: ID of the topic to write about
        style_id: ID of the writing style to use
        source_ids: Optional list of source IDs for reference material
        notes: Additional instructions or context for generation
    """
    return await _request("POST", "/generate", json={
        "topic_id": topic_id,
        "style_id": style_id,
        "source_ids": source_ids or [],
        "notes": notes,
    })


@mcp.tool()
async def tweak_draft(content: str, instruction: str) -> dict | list | str:
    """Revise content with a natural language instruction.

    Takes existing content and an instruction describing how to change it,
    then returns the revised content. Useful for iterating on drafts.

    Args:
        content: The current content to revise
        instruction: What to change, e.g. "make it shorter", "add more examples",
            "change the tone to be more formal"
    """
    return await _request("POST", "/tweak", json={
        "content": content,
        "instruction": instruction,
    })


# ── Settings ────────────────────────────────────────────────────────────────


@mcp.tool()
async def get_settings() -> dict | list | str:
    """Get current LLM settings.

    Returns the configured LLM provider, model, and API key.
    These must be set before using generate_drafts or tweak_draft.
    """
    return await _request("GET", "/settings")


@mcp.tool()
async def save_settings(
    provider: Optional[str] = None,
    model: Optional[str] = None,
    api_key: Optional[str] = None,
) -> dict | list | str:
    """Save LLM settings (provider, model, API key).

    Configure the LLM backend used for draft generation and tweaking.
    Each parameter is saved individually — only provide the ones you want to change.

    Args:
        provider: LLM provider, e.g. "openai", "anthropic"
        model: Model name, e.g. "gpt-4", "claude-3-opus-20240229"
        api_key: API key for the provider
    """
    results = []
    if provider is not None:
        r = await _request("POST", "/settings", json={"key": "llm_provider", "value": provider})
        results.append(f"provider: {r}")
    if model is not None:
        r = await _request("POST", "/settings", json={"key": "llm_model", "value": model})
        results.append(f"model: {r}")
    if api_key is not None:
        r = await _request("POST", "/settings", json={"key": "llm_api_key", "value": api_key})
        results.append(f"api_key: {r}")
    if not results:
        return "No settings provided to save."
    return "; ".join(results)


if __name__ == "__main__":
    mcp.run(transport="stdio")

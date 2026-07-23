# 🔷 Gorenel MCP Bridge & Roadmap

This document outlines the technical implementation, usage guide, and future roadmap of the **Gorenel Model Context Protocol (MCP) Bridge**. It allows developers to securely expose their local stdio-based MCP servers to remote AI models (like Claude Code, Cursor, or custom OpenAI/Anthropic agents) over an encrypted HTTP/SSE tunnel.

---

## 🚀 How It Works (Architecture)

Traditional MCP servers communicate locally using standard input/output (`stdin`/`stdout`). If you want to connect a remote AI agent to your local environment, you have to bridge that gap. 

Gorenel spawns your local MCP server as a subprocess, intercepts its standard I/O streams, and exposes them as a single, multiplexed HTTP/SSE endpoint at a public URL (e.g., `https://xxxx.gorenel.site/sse`).

```
┌─────────────┐           ┌──────────────┐           ┌──────────────┐           ┌──────────────┐
│  AI Client  │ ──HTTP──> │   Gorenel    │ ──YAMUX──>│ Gorenel CLI  │ ──StdIn──>│  Local MCP   │
│ (Claude/AI) │ <──SSE─── │ Cloud Relay  │ <──SSH─── │   (Bridge)   │ <──StdOut─│ Subprocess   │
└─────────────┘           └──────────────┘           └──────────────┘           └──────────────┘
```

---

## 🛠️ Quick Start

### 1. Start the Tunnel
To start the bridge, run the Gorenel CLI with the `--command` parameter pointing to your local MCP server startup script:

```bash
go run cmd/client/main.go mcp --command "node mcp_server.js"
```
*Note: Make sure to wrap the inner command in double/single quotes if it contains spaces or arguments.*

### 2. Configure Your AI Client (e.g., Claude Code)
Use the allocated tunnel URL in your client configuration. For Claude Code:

```bash
claude mcp add --transport http gorenel-test https://<YOUR_SUBDOMAIN>.gorenel.site/sse
```

### 3. Clear Caches & Run
Restart the Claude daemon to ensure it picks up the new config:
```bash
claude daemon stop --any
claude
```
Now, you can ask Claude to run tools from your local server!

---

## 🗺️ Feature Roadmap

### 🟩 Phase 1: Core Stdio Bridge (Completed)
- [x] **Subprocess Management:** Clean execution of subprocesses with stdin/stdout piping.
- [x] **Streamable HTTP Support:** Bridge modern MCP senkron request-response protocols (`POST /sse` matching JSON-RPC `id` fields) and legacy asenkron HTTP+SSE (`GET /sse` streams).
- [x] **Windows Stdout Buffering Fix:** Resolved Node.js stdout pipe buffering using synchronous file descriptor write locks.
- [x] **CORS & Options Support:** Native pre-flight handling for cross-origin integration with web-based agent environments.

### 🟨 Phase 2: Security & Authentication (Q3 2026)
- [ ] **Edge Tokens:** Require auth headers (`Authorization: Bearer <key>`) at the global edge proxy to prevent unauthenticated access to your local tools.
- [ ] **IP Pinning:** Restrict tunnel access only to official IP blocks of AI providers (e.g., Anthropic, OpenAI, or Cursor IPs).
- [ ] **Request Inspector:** Real-time logging of incoming JSON-RPC payloads in the Gorenel Sniffer Web Dashboard.

### 🟨 Phase 3: Multi-Process Orchestration (Q4 2026)
- [ ] **ConfigFile Deployment:** Expose multiple local MCP servers simultaneously through a single tunnel configuration (e.g., a `.gorenel-mcp.json` file config).
- [ ] **Process Auto-Reloading:** Watch files in the local MCP workspace and hot-reload/restart subprocesses upon changes.
- [ ] **Health Checks & Recovery:** Auto-restart crashed subprocesses with back-off retry logic.

### 🟦 Phase 4: Developer Ecosystem (Q1 2027)
- [ ] **Local Sandbox Environments:** Integrated containerized execution environments to securely test MCP tools.
- [ ] **Official Python/Go SDKs:** Native packages to integrate Gorenel directly inside orchestrators like LangChain, AutoGen, or LlamaIndex.
- [ ] **Pre-built Tool Hub:** Easy ingestion of popular local MCP tools (PostgreSQL, Git, Docker, Puppeteer) through simple commands like `gorenel install mcp-postgres`.

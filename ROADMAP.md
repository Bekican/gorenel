# Gorenel Product Roadmap 🚀

This document outlines the short-term, mid-term, and long-term roadmap for Gorenel. We are building the next-generation, developer-first, open-source secure reverse tunneling platform.

---

## 🟢 Phase 1: Core Foundation & Lightweight ML (Completed)
* [x] **High-Performance Relay Server:** Go-based control-plane and proxy engine using `yamux` multiplexing.
* [x] **Lightweight Streaming ML Engine:** Replaced heavy Keras/TensorFlow model with `river` HalfSpaceTrees anomaly detection (~30MB RAM footprint).
* [x] **Robust GeoIP Parsing:** Real client IP detection and dynamic region naming (e.g. "Istanbul", "Frankfurt").
* [x] **Smart Client-Side Hydration:** Eliminated React SSG DOM mismatches on the dashboard interface.

---

## 🟡 Phase 2: Open Source Visibility & DX Refinements (In Progress)
* [x] **Self-Hosted Template:** Standardized `docker-compose.self-hosted.yml` configurations for 1-click private hosting.
* [x] **CLI DX Shortcuts:** Introduce standard shortcuts: `gorenel http <port>` and `gorenel tcp <port>`.
* [ ] **Automated Signed Releases:** Publish signed client binaries with checksums on GitHub Releases.
* [ ] **Detailed Benchmarks:** Publish CPU, memory, and latency comparison graphs against ngrok and Cloudflare Tunnels.

---

## 🔵 Phase 3: Advanced Access Policies & Security Controls (Q3 2026)
* [ ] **JWT Authentication Policy:** Authenticate external client requests directly at the Gorenel Edge before routing traffic to the localhost.
* [ ] **Wildcard Custom Domains:** Add simple support for mapping entire wildcard domain networks to your local tunnels.
* [ ] **Advanced Anomaly Active Mitigation:** Option to automatically prompt and block anomalous clients using Google ReCAPTCHA or cloud-level firewalls.

---

## 🟣 Phase 4: Enterprise Scalability & Developer Ecosystem (Q4 2026)
* [ ] **Kubernetes Operator:** Run self-hosted multi-tenant tunnel clusters on Kubernetes natively.
* [ ] **MCP (Model Context Protocol) Gateway Support:** Native secure tunnels configured specifically for AI Agents and local LLM services.
* [ ] **Technical Blog:** Launch `blog.gorenel.site` with detailed engineering write-ups:
  - *Websocket Multiplexing in Go*
  - *Streaming ML at the Edge*
  - *How we achieved sub-20ms latency inside the Turkish / European cloud ecosystem.*

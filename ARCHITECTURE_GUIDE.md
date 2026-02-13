# Gorenel: Modern Reverse Tunneling & Proxy Architecture

This document provides a comprehensive guide to the Gorenel project, explaining how different components and files are interconnected.

## 🌟 Overview
Gorenel is a high-performance reverse tunneling system built with Go. It allows developers to expose local servers (HTTP, TCP, or UDP) to the internet through a public gateway, similar to Ngrok. It includes advanced features like ML-based anomaly detection, geolocation, and real-time dashboarding.

---

## 🏗️ Backend Architecture (Go)

The backend is structured using modern Go practices, separating entry points (`cmd/`) from internal logic (`internal/`) and shared packages (`pkg/`).

### 1. Entry Points (`cmd/`)
*   **[server/main.go](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/cmd/server/main.go):** The heart of the system. It initializes the Tunnel Manager, Proxy Servers (HTTP/TCP/UDP), ML clients, and the Monitoring Server.
*   **[client/main.go](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/cmd/client/main.go):** Minimal entry point that calls the Cobra-based CLI.
*   **[client/cmd/](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/cmd/client/cmd/):** Defines CLI commands. `start.go` contains the logic for establishing a connection to the server and multiplexing local traffic.

### 2. Core Logic (`internal/`)
*   **[server/tunnel_manager.go](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/internal/server/tunnel_manager.go):** Manages active tunnels, subdomain assignments, and port allocations.
*   **[server/http_proxy.go](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/internal/server/http_proxy.go):** Handles incoming HTTP traffic, routes it to the correct tunnel, and integrates with the ML engine for traffic inspection.
*   **[protocol/](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/internal/protocol/):** Defines the custom wire protocol used between the client and server (Registration, Messages, Error handling).
*   **[ml/](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/internal/ml/):** Client for the Python-based Machine Learning service that detects anomalous traffic patterns.

### 3. Shared Utilities (`pkg/`)
*   **[auth/](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/pkg/auth/):** JWT-based authentication logic for both the dashboard and client connections.
*   **[logger/](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/pkg/logger/):** Standardized Zap-based logging used across the entire system.

---

## 🖥️ Frontend Architecture (React)

The dashboard is built with React, Vite, and Tailwind CSS, located in `web-dashboard/`.

*   **[src/main.tsx](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/web-dashboard/src/main.tsx):** Application entry point.
*   **[src/App.tsx](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/web-dashboard/src/App.tsx):** Main layout and routing logic.
*   **[src/components/](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/web-dashboard/src/components/):** Contains reusable UI elements like `TunnelsList.tsx` and `Analytics.tsx`.
*   **[src/api/client.ts](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/web-dashboard/src/api/client.ts):** Centralized Axios client for communicating with the Go Monitoring Server.

---

## 🌐 Infrastructure & DevOps

*   **[docker-compose.yml](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/docker-compose.yml):** Orchestrates the Go server, Redis (for pub/sub), ClickHouse (for analytics), and ML services.
*   **[helm/](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/helm/):** Kubernetes deployment configurations for production environments.
*   **[Makefile](file:///c:/Users/Bekir%20Can/Desktop/advancedbackend/gorenel/Makefile):** Shortcuts for building, testing, and running the various components of the project.

---

## 🔄 Data Flow: How a Request Travels
1.  **User Request:** A user visits `abc.gorenel.io:8080`.
2.  **Server Proxy:** `http_proxy.go` intercepts the request, checks `tunnel_manager.go` for the session matching `abc`.
3.  **ML Check:** The request is sent to the ML service (`ml/client.go`) to check for anomalies.
4.  **Multiplexing:** The request is packaged into a `yamux` stream and sent over the persistent TCP connection to the Gorenel Client.
5.  **Local Delivery:** The Gorenel Client (`cmd/client/cmd/start.go`) receives the stream and forwards it to the local port (e.g., `localhost:3000`).
6.  **Response:** The response travels back through the same tunnel to the user.

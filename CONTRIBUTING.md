# Contributing to Gorenel

Thank you for your interest in contributing to Gorenel! We are building a modern, open-source, developer-first tunneling platform, and we welcome contributions from the community.

---

## 🛠️ Local Development Setup

To build and run Gorenel locally on your system, ensure you have the following installed:
* **Go** 1.21+
* **Python** 3.10+ (for the ML anomaly service)
* **Node.js** 18+ & **npm** (for the web dashboard)

### 1. Clone the Repository
```bash
git clone https://github.com/Bekican/gorenel.git
cd gorenel
```

### 2. Run the Go Server (Relay & Control Plane)
Copy the example environment file and run:
```bash
cp .env.example .env
# Edit .env with your local secrets
go run cmd/server/main.go
```

### 3. Run the ML Service
Instantiate a Python virtual environment and run the API:
```bash
cd services/ml
python -m venv venv
source venv/bin/activate  # On Windows: .\venv\Scripts\activate
pip install -r requirements.txt
python app.py
```

### 4. Build and Launch the Dashboard
```bash
cd web-dashboard
npm install
npm run dev
```

### 5. Build the CLI Client
To compile the CLI executable locally:
```bash
go build -o bin/gorenel cmd/client/main.go
```
Now you can start a tunnel using:
```bash
./bin/gorenel http 3000
```

---

## 🤝 Contribution Guidelines

### Branching Policy
* Create feature branches off the `main` branch.
* Use descriptive branch names: `feature/new-command` or `fix/ws-keepalive`.

### Commit Style
We follow [Conventional Commits](https://www.conventionalcommits.org/):
* `feat(cli): add http command shortcut`
* `fix(server): fix client IP resolution header chain`
* `docs(readme): update quickstart guidelines`

### Submitting a Pull Request
1. Keep pull requests focused on a single change.
2. Run standard tests and linters before submitting:
   ```bash
   go test -v -race ./...
   ```
3. Open a Pull Request detailing the changes, and link any corresponding issues.

# Security Policy

We take security seriously at Gorenel. As a reverse tunneling tool that exposes local development ports to the public internet, security, encryption, and privacy are core to everything we do.

## Supported Versions

Only the latest release versions are actively supported and patched.

| Version | Supported |
| ------- | --------- |
| v1.2.x  | ✅ Yes    |
| < v1.2.0 | ❌ No     |

## Reporting a Vulnerability

If you discover a security vulnerability within Gorenel, please report it immediately. **Do not open a public GitHub issue.** Instead, follow these steps:

1. Send an email to **security@gorenel.site** (or email the repository maintainers directly).
2. Include a detailed description of the vulnerability, along with:
   - Steps to reproduce (Proof of Concept).
   - The potential impact of the issue.
   - Any proposed mitigations or fixes.
3. We will acknowledge your report within **48 hours** and work with you to analyze, fix, and release a patched version in a coordinated manner.

## Secure Usage Practices

* **Always keep your CLI client updated** to prevent local security exploits.
* **Never share your Gorenel API key** or commit config files (`config.yaml`) containing active keys into public repositories.
* Use our built-in **Edge Policies** (e.g. basic auth and IP whitelists) when exposing sensitive local tools like databases or internal dashboards:
  ```bash
  # Expose port with basic password protection:
  gorenel http 3000 --auth "admin:mysecretpassword"

  # Expose port only allowed to your specific client IP:
  gorenel http 3000 --ip-whitelist 203.0.113.50
  ```

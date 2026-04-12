# Security Policy

## Supported Versions

This project is maintained on a rolling release basis.

| Version | Status |
| --- | --- |
| Latest tagged release | Supported |
| `main` | Best effort for contributors and evaluators |
| Older tags | Not a supported security baseline |

## Reporting A Vulnerability

Do not open a public GitHub issue for a security problem.

Please use GitHub's private vulnerability reporting for this repository. If private reporting is unavailable, contact the maintainer privately through GitHub and mark the report as security-sensitive.

Please include:

- affected version, tag, or commit
- impact summary
- reproduction steps or proof of concept
- whether the issue exposes cluster data, credentials, or report artifacts
- whether the issue affects the CLI, desktop app, generated artifacts, or release pipeline

We aim to acknowledge reports promptly and keep follow-up communication private until a fix or mitigation is ready.

## Scope Notes

This tool reads Kubernetes metadata and may emit operationally sensitive names, endpoints, storage locations, and backup context in its artifacts.

Before sharing scan output outside your team:

- prefer `--redact` for exported reports
- review `recovery-scan.json` and `recovery-enriched.json` for environment-specific identifiers
- use least-privilege RBAC from [`deploy/rbac`](deploy/rbac) where possible
- avoid enabling Secret metadata collection unless the security tradeoff is explicitly approved

## Security Posture

- The scanner is designed to be read-only against the Kubernetes API.
- Backup conclusions are intentionally conservative and do not claim recoverability without evidence.
- Release assets are built with checksums and SBOM generation in GitHub Actions.
- Desktop and report outputs are designed to stay offline-friendly and CDN-independent.

For general help, use [SUPPORT.md](SUPPORT.md). For contributor workflow details, see [CONTRIBUTING.md](CONTRIBUTING.md).

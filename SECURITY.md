# Security Policy

## Supported Versions

This project is maintained on a rolling basis.

- The most recent tagged release is the supported release line for operators.
- `main` is supported on a best-effort basis for contributors and early adopters.
- Older tags may continue to work, but they should not be treated as a supported security baseline.

## Reporting a Vulnerability

Do not open a public GitHub issue for a security problem.

Please report vulnerabilities privately by using GitHub's private vulnerability reporting for this repository when available, or by contacting the repository owner directly through GitHub.

Include:

- affected version or commit
- impact summary
- reproduction steps or proof of concept
- whether the issue exposes cluster data, credentials, or report artifacts

## Scope Notes

This tool reads Kubernetes metadata and may emit operationally sensitive names, endpoints, or storage locations in its artifacts.

Before sharing scan output outside your team:

- prefer `--redact` for exported reports
- review `recovery-scan.json` and `recovery-enriched.json` for environment-specific identifiers
- use least-privilege RBAC from [`deploy/rbac`](deploy/rbac) where possible

## Security Posture

- The scanner is designed to be read-only against the Kubernetes API.
- Release assets are built with checksums and SBOM generation in GitHub Actions.
- Backup conclusions are intentionally conservative and do not claim recoverability without evidence.

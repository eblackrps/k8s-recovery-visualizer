# Privacy Policy

Effective date: April 13, 2026

This Privacy Policy applies to the `k8s-recovery-visualizer` repository, the `K8V` desktop application, and the source-built CLI distributed from this project.

## Summary

`K8V` is a local-first desktop application and `k8s-recovery-visualizer` also ships a source-built CLI. The project does not operate a hosted application service for scans, reports, or bundle review. In normal use, scan processing and artifact generation happen on your machine.

## Information The Software Processes

Depending on how you use the project, the software may process:

- Kubernetes connection details you choose to supply, such as kubeconfig paths, selected contexts, and namespace scope
- cluster metadata and recovery-readiness data retrieved during a live scan
- bundle files and archives you choose to open
- generated output files such as JSON bundles, HTML reports, Markdown reports, summaries, runbooks, CSV exports, and redacted exports
- desktop settings such as workspace root, default output directory, selected defaults, and export preferences
- local startup diagnostics written by the desktop app when it launches

## Local Storage

The desktop app stores some local data on the machine where you run it.

Examples include:

- desktop settings in the user config directory, such as `settings.json`
- a startup log in the user config directory, such as `k8v-startup.log`
- Windows WebView2 local browser/runtime data used by the embedded desktop shell
- scan outputs and exported artifacts in the workspace or output directory you choose

The project does not require you to create an account to run the desktop app or CLI.

## What The Project Does Not Do

Based on the current application behavior in this repository:

- the desktop app and CLI do not include built-in telemetry or product analytics
- the project does not include a project-operated backend that receives scan results automatically
- the desktop app does not include a built-in advertising SDK
- the project does not sell or rent your personal information
- the project does not perform automatic cloud upload of scan bundles, reports, or settings on your behalf

## When Information May Leave Your Machine

Information may leave your machine in a few cases:

- when you run a live scan, the software connects from your machine to the Kubernetes API endpoints and related infrastructure you configure
- when you download releases, browse the repository, or open issues on GitHub, GitHub may process information under GitHub's own policies
- when you voluntarily share logs, screenshots, reports, bundles, or issue details with the maintainer or in public issues

The project maintainer does not receive cluster data automatically just because you run the app.

## Sensitive Infrastructure Data

Scan outputs can include environment-specific information such as namespace names, workload names, storage details, network endpoints, image references, backup inventory, and other operational metadata.

Before sharing outputs outside your team, consider:

- using redacted exports when appropriate
- reviewing JSON and report artifacts for environment-specific identifiers
- avoiding Secret metadata collection unless you explicitly need it and accept that tradeoff

## Retention And Control

Your local settings, logs, WebView2 data, and generated outputs remain on your machine until you delete them or your operating system removes them. You control where output artifacts are written.

If you share information through GitHub issues, GitHub releases, private vulnerability reports, or other third-party services, retention for that data is governed by those services.

## Security

The repository currently writes desktop settings and startup logs with user-scoped filesystem permissions where supported by the operating system. Even so, you should treat scan artifacts as potentially sensitive and store or share them accordingly.

For security-sensitive disclosures, use the process in [SECURITY.md](SECURITY.md).

## Changes To This Policy

This Privacy Policy may be updated as the project changes. The version published in this repository is the current project policy.

## Contact

For privacy or support questions about this project:

- see [SUPPORT.md](SUPPORT.md) for general support
- see [SECURITY.md](SECURITY.md) for security-sensitive matters
- open a GitHub issue or contact the maintainer through GitHub for project-related questions

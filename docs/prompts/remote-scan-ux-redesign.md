# Prompt 1: Remote-First Scan UX Redesign

You are working in `C:\k8visualizer` on the `K8V` Windows desktop app.

Your job is to redesign and implement the scan setup flow so the product is production-ready for broad customer use on:

- a local Windows desktop
- a jumpbox on the network
- any machine that needs to scan a remote Kubernetes cluster

The current scan wizard is too confusing. It exposes too many low-level knobs, assumes a manually provided kubeconfig path too early, and does not support a direct control-plane endpoint workflow. The app must become much simpler to use without removing useful capability.

## Current Code To Understand First

Read and work from these files before making changes:

- `desktop/frontend/src/views/ScanView.tsx`
- `desktop/frontend/src/App.tsx`
- `desktop/frontend/src/lib/types.ts`
- `desktop/scan_controller.go`
- `internal/appcore/types.go`
- `internal/appcore/preflight.go`
- `internal/appcore/run.go`
- `internal/kube/client.go`

## Product Direction

Assume the common customer workflow is:

- launch the Windows app on a workstation or jumpbox
- connect to a remote cluster
- scan it with the least amount of setup possible
- collect and report everything useful by default

The app is local, but the cluster is usually remote. Design and wording must reflect that clearly.

## Main Outcome

Make the scan experience feel "stupid simple" for first-time operators while preserving advanced functionality behind progressive disclosure.

The default path should be fast, obvious, and hard to misuse.

## Core UX Goals

1. Make remote scanning the primary mental model.
2. Allow scanning with either:
   - an existing Kubernetes login on the machine
   - a kubeconfig file
   - pasted kubeconfig content
   - a direct control-plane endpoint by IP or hostname
3. Remove or hide options that feel like CI, developer, or internal plumbing.
4. Keep power-user features available, but move them behind an `Advanced options` section.
5. Add lightweight in-app help so users do not need outside docs just to understand the form.

## Important Constraint

Do not fake or imply that an IP or hostname alone is enough to scan a cluster.

A direct endpoint workflow still needs authentication and TLS inputs. The UI must explain this clearly and guide the user into a valid connection method.

## Required UX Changes

Replace the current 4-step "Access / Scope / Outputs / Review" wizard with a simpler remote-first flow.

Recommended information architecture:

### 1. Connect

Provide a connection method selector with plain-English choices:

- `Use current Kubernetes login`
- `Use kubeconfig file`
- `Paste kubeconfig`
- `Connect to API endpoint`

Expected behavior:

- `Use current Kubernetes login`
  - Default recommended path
  - Use normal loading rules, `KUBECONFIG`, and current kubectl context
  - Show detected context if available
  - Allow optional context override
- `Use kubeconfig file`
  - Browse or paste path
  - Auto-load available contexts if possible
- `Paste kubeconfig`
  - Multiline text input
  - Parse and validate it in memory
- `Connect to API endpoint`
  - Allow API server URL or host/IP with sane normalization
  - Support explicit auth input

For direct endpoint mode, support at minimum:

- endpoint / control plane host
- bearer token auth
- CA bundle or CA file input
- optional `Skip TLS verification`

If practical within the existing architecture, also support:

- client certificate + key input or file paths

Do not add username/password auth unless there is a strong reason in this codebase.

### 2. Scope

Keep only the scope settings most users actually need:

- `All namespaces` as the default
- optional namespace list when the user chooses scoped mode
- cluster label
- environment label

Cluster and environment should be:

- auto-populated when possible from kubeconfig/context/endpoint
- editable by the user

### 3. Output

Keep the default reporting path simple:

- output directory
- generate the standard artifact set by default

The defaults should be suitable for normal customer use.

Use sensible defaults for:

- HTML report
- JSON bundle
- enriched JSON
- executive summary
- runbook
- CSV exports if they are already part of the default desktop workflow and do not create confusion

### 4. Advanced Options

Move non-essential or expert settings behind a clearly labeled expander, drawer, or modal.

Advanced options can include:

- compare baseline
- scoring profile
- recovery target
- timeout
- TLS skip
- include secret metadata
- redacted exports

Do not expose these in the main happy path unless they are truly necessary.

## Options To Remove From Normal Production UI

These should not be in the main operator flow:

- dry-run fixture mode
- minimum score

If needed, keep them only in:

- developer mode
- debug builds
- hidden advanced/debug affordances not visible to normal customers

## Help And Guidance

Add inline help without making the screen noisy.

Acceptable patterns:

- small `?` help buttons
- hover or click tooltips
- a compact `Help` drawer
- a `What do these mean?` link near the connection area

At minimum, add help for:

- connection method selection
- kubeconfig behavior when path is left blank
- direct endpoint requirements
- namespace scope
- secret metadata
- TLS verification
- compare baseline

## Backend Requirements

Refactor the request and kube loading model so the frontend is not forced to pretend everything is a kubeconfig path.

Introduce a clearer access model in the desktop request path. You may evolve `ScanRequest` and related code as needed.

Support:

- existing default kube loading rules
- `KUBECONFIG`
- explicit kubeconfig file path
- inline kubeconfig content
- direct endpoint config construction

The same access resolution path must be used by both:

- preflight
- scan execution

Avoid duplicate connection logic.

## Security And Data Handling Requirements

Be careful with credentials:

- do not log bearer tokens, certificate data, or inline kubeconfig content
- do not echo secrets into review cards or banners
- avoid persisting inline credentials into saved desktop settings unless the user explicitly asks for that behavior
- keep startup logs and error messages helpful but sanitized

## Simplicity Rules

Design decisions should favor:

- fewer fields shown up front
- clearer wording
- remote-cluster mental model
- progressive disclosure for advanced settings

Do not remove useful underlying capabilities. Re-home them.

## Suggested Implementation Areas

You will likely need changes in:

- `desktop/frontend/src/views/ScanView.tsx`
- `desktop/frontend/src/App.tsx`
- `desktop/frontend/src/lib/types.ts`
- `internal/appcore/types.go`
- `internal/appcore/preflight.go`
- `internal/appcore/run.go`
- `internal/kube/client.go`

You may add focused backend bindings if needed, such as:

- listing available contexts
- validating connection methods
- parsing inline kubeconfig safely

## Acceptance Criteria

The work is complete only if all of the following are true:

1. A first-time user can scan a remote cluster using the default kubectl login path without being forced to browse for a kubeconfig file.
2. A user can scan using an explicit kubeconfig file.
3. A user can scan using pasted kubeconfig content.
4. A user can scan using a direct control-plane endpoint plus valid auth material.
5. The main UI is materially simpler than the current wizard.
6. Dry-run and minimum score are no longer part of the normal production scan form.
7. Advanced options still preserve important existing capabilities.
8. Help affordances exist for the most confusing fields.
9. Preflight and scan both work through the same connection-resolution model.
10. Tests are updated or added for the new access paths and UI behavior.

## Validation Expectations

Before finishing:

- run relevant frontend tests
- run relevant Go tests
- manually verify the local `kind-k8v-test` context still works
- verify a blank kubeconfig path still works with default loading rules
- verify invalid direct endpoint auth produces a clear preflight error

## Deliverable

Implement the redesign, not just a mockup.

Ship working code, tests, and any concise docs updates needed so the desktop app becomes a remote-first operator tool rather than a developer-shaped wizard.

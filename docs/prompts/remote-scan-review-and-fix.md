# Prompt 2: Review, Debug, And Fix The Remote Scan Redesign

You are reviewing the completed implementation from:

- `docs/prompts/remote-scan-ux-redesign.md`

This is not a surface-level summary pass. Treat this as a production-readiness review focused on real operator use.

Your job is to:

1. review the implementation critically
2. find bugs, regressions, confusing flows, and security issues
3. fix what you find
4. leave the scan experience more reliable and less confusing than the first pass

## Review Priorities

### 1. Connection Flows

Verify each supported scan path works or fails clearly:

- current Kubernetes login / default loading rules
- `KUBECONFIG`
- explicit kubeconfig file
- pasted kubeconfig content
- direct endpoint with token auth
- direct endpoint with TLS verification enabled
- direct endpoint with TLS verification disabled

Check for:

- broken parsing
- context resolution bugs
- duplicate logic between preflight and run
- inconsistent behavior between UI and backend

### 2. Operator Simplicity

Review whether the UI is actually simpler in practice.

Look for:

- too many visible fields
- unclear labels
- jargon-heavy wording
- hidden but still critical settings
- overuse of steps, drawers, or nested controls

If anything still feels engineer-shaped rather than operator-shaped, fix it.

### 3. Remote-First Product Fit

Make sure the app now clearly communicates that:

- the app runs locally
- the cluster is often remote
- kubeconfig is optional when machine credentials already work
- endpoint-only access still requires auth

Look for any text that implies:

- local cluster only
- kubeconfig path always required
- IP/hostname alone is enough

Fix all such confusion.

### 4. Security And Privacy

Check carefully that:

- tokens are not rendered back into review cards
- inline kubeconfig content is not logged
- secrets are not emitted in banners or startup logs
- saved settings do not silently persist sensitive credentials unless explicitly intended
- error messages remain useful without leaking secrets

### 5. Accessibility And Help

Review:

- keyboard access
- tooltip/help affordance usability
- focus behavior
- readable validation and error states

If help exists but is clumsy, hard to discover, or inaccessible, improve it.

### 6. Regression Risk

Ensure existing functionality still works:

- namespace-scoped scans
- output selection and artifact generation
- compare baseline
- profile handling
- recovery target behavior
- include secret metadata
- current local `kind-k8v-test` scan path

Do not preserve broken compatibility just because a field used to exist. Preserve user capability.

## Required Validation

Run and verify as much of the relevant stack as possible, including:

- frontend tests
- Go tests
- manual preflight and scan sanity checks
- the local `kind-k8v-test` cluster path

If you find missing tests, add them.

## Required Output Behavior

At the end of this pass:

- the app should feel easier than the first implementation
- confusing text should be reduced further
- any discovered bugs should be fixed
- tests should cover the new access paths and key UI logic

## Deliverable

Do the review, make the fixes, run validation, and leave the code in a production-improved state.

Do not stop at findings only. Apply fixes for anything important you uncover.

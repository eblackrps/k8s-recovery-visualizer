# Release Guide

Use this checklist when preparing a tagged release for `k8s-recovery-visualizer`.

## Versioning

- use semver tags such as `v1.6.0`
- keep source-level version references aligned with the release you are cutting
- bump schema versions only when the JSON contract changes
- use a minor or major release only when the user-visible surface justifies it

## Public Release Scope

This release line publishes exactly four GitHub release assets:

- `k8s-recovery-visualizer-desktop-linux-amd64.tar.gz`
- `k8s-recovery-visualizer-desktop-windows-amd64.zip`
- `checksums.txt`
- `k8s-recovery-visualizer.spdx.json`

Deprecated public release artifacts:

- prebuilt CLI binaries
- macOS desktop package
- GHCR container image

The CLI remains supported for source builds, CI, smoke tests, and contributor workflows.

## Pre-Release Checklist

Run the full validation set:

```bash
make fmt
go build ./...
make vet
make test
make race
make frontend-install
make frontend-build
make frontend-test
make screenshots
make smoke
make schema-samples
make docs-check
make build-gui
```

Also confirm:

- Linux desktop packaging is validated before release
- Windows desktop packaging is validated before release
- screenshot references in the docs resolve correctly
- the desktop app opens existing bundles and refreshes exports
- the release gate is at least as strict as CI before any tag is published
- README, screenshots, changelog, and release notes all match the actual shipped artifacts
- no public-release docs or workflow paths still imply CLI binaries, macOS packages, or GHCR container publishing

## Local Build Notes

Current-host desktop build:

```bash
make build-gui
```

Current-host desktop package:

```bash
make package-gui
```

Host CLI source build:

```bash
make build
```

Contributor cross-platform CLI build:

```bash
make build-cli-cross
```

Windows packaging uses NSIS. If `make package-gui` fails locally with `makensis not found`, install NSIS before retrying.

## GitHub Release Flow

1. Update the changelog, README, screenshots, and public docs.
2. Merge the release branch to `main`.
3. Create and push the release tag.
4. Let [`.github/workflows/release.yml`](../.github/workflows/release.yml) rebuild and publish the supported desktop assets.
5. Verify the GitHub release title, notes, checksums, SBOM, and asset list after the workflow completes.
6. Confirm the release contains only the two desktop packages plus `checksums.txt` and the SPDX SBOM.

## GitHub Release Workflow

Tags matching `v*` trigger [`.github/workflows/release.yml`](../.github/workflows/release.yml).

The release workflow:

1. reruns the CI-grade verification gate
2. builds the Linux desktop package
3. builds the Windows desktop package
4. generates the SPDX SBOM
5. generates release checksums
6. publishes the GitHub release using notes derived from `CHANGELOG.md`

## Asset Verification

Expected release outputs:

- `k8s-recovery-visualizer-desktop-linux-amd64.tar.gz`
- `k8s-recovery-visualizer-desktop-windows-amd64.zip`
- `checksums.txt`
- `k8s-recovery-visualizer.spdx.json`

For contributor workflow details, see [../CONTRIBUTING.md](../CONTRIBUTING.md).

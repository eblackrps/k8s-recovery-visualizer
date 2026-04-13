# Release Guide

Use this checklist when preparing a tagged release for `k8s-recovery-visualizer`.

## Versioning

- use semver tags such as `v1.5.2`
- keep source-level version references aligned with the release you are cutting
- bump schema versions only when the JSON contract changes
- only use a minor or major release when the user-visible surface justifies it

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

Confirm:

- CLI dry-run output still matches the published schemas
- report palette still reflects the shared tokens in `internal/theme`
- screenshot references in the docs resolve correctly
- the desktop app opens existing bundles and refreshes exports
- the release gate is at least as strict as CI before any tag is published
- README, screenshots, and release notes all match the actual shipped artifacts

## Local Packaging

Current-host desktop build:

```bash
make build-gui
```

Current-host desktop package:

```bash
make package-gui
```

Cross-platform CLI binaries:

```bash
make release-cli
```

Windows packaging uses NSIS. If `make package-gui` fails locally with `makensis not found`, install NSIS before retrying.

## GitHub Release Flow

1. Update the changelog, README, screenshots, and public docs.
2. Merge the release branch to `main`.
3. Create and push the release tag.
4. Let [`.github/workflows/release.yml`](../.github/workflows/release.yml) rebuild and publish the release assets.
5. Verify the GitHub release title, notes, checksums, SBOM, and desktop packages after the workflow completes.
6. Update repository metadata such as description, topics, homepage, and social preview if needed.

## GitHub Release Workflow

Tags matching `v*` trigger [`.github/workflows/release.yml`](../.github/workflows/release.yml).

The release workflow:

1. reruns the CI-grade verification gate
2. builds cross-platform CLI binaries
3. builds native desktop artifacts on Linux, Windows, and macOS runners
4. generates checksums and an SPDX SBOM
5. publishes the GitHub release using notes derived from `CHANGELOG.md`
6. pushes the container image to GHCR with provenance and SBOM metadata

## Release Assets

Expected release outputs include:

- CLI binaries for Linux, macOS, and Windows
- desktop packages for native GUI runners
- checksums file
- SPDX SBOM
- GitHub release notes

For contributor workflow details, see [../CONTRIBUTING.md](../CONTRIBUTING.md).

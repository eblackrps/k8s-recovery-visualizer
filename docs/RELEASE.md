# Release Guide

`v1.4.0` ships both CLI and desktop assets while keeping schema compatibility intact.

## Versioning

- prefer semver tags such as `v1.4.0`
- keep `internal/model.Version` on the next `-dev` value between releases
- bump schema versions only when the JSON contract changes

## Pre-Release Checklist

Run the full validation set:

```bash
make fmt
make vet
make test
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

For contributor workflow details, see [CONTRIBUTING.md](../CONTRIBUTING.md).

# Desktop Build Assets

This directory contains the Wails packaging assets for the `desktop/` application.

- `appicon.png`: shared source icon for generated platform assets.
- `darwin/Info.plist`: macOS app metadata kept for source builds and contributor workflow.
- `darwin/Info.dev.plist`: macOS metadata used during `wails dev`.
- `windows/info.json`: Windows executable metadata.
- `windows/installer/`: NSIS installer templates used for packaged Windows releases.

Primary packaging commands live in the root [Makefile](../../Makefile) and in [docs/RELEASE.md](../../docs/RELEASE.md).

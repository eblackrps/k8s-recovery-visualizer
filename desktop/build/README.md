# Desktop Build Assets

This directory contains the Wails packaging assets for the `desktop/` application.

- `appicon.png`: shared source icon for generated platform assets.
- `darwin/Info.plist`: macOS app metadata retained for source builds and contributor workflow only.
- `darwin/Info.dev.plist`: macOS metadata used during `wails dev`.
- `windows/info.json`: Windows executable metadata.
- `windows/installer/`: NSIS installer templates used for packaged Windows releases.

Public GitHub releases currently publish Linux amd64 and Windows amd64 desktop packages only. These build assets still include contributor-oriented metadata for source builds on other platforms.

Primary packaging commands live in the root [Makefile](../../Makefile) and in [docs/RELEASE.md](../../docs/RELEASE.md).

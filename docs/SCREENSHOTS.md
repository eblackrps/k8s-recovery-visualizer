# Desktop Screenshot Workflow

The repository keeps a deterministic desktop screenshot set in `images/` for the main README and supporting docs.

## Current Screenshot Set

- `images/gui-dashboard.png`
- `images/gui-scan-setup.png`
- `images/gui-live-run.png`
- `images/gui-results-findings.png`
- `images/gui-compare.png`

These five files are the maintained public screenshot set for the README and documentation. Refresh them whenever the desktop UI or landing-page gallery changes.

Deprecated screenshots that are no longer part of the maintained desktop gallery should be removed from docs and repo references instead of carried forward indefinitely.

## Generate The GUI Screenshots

Install frontend dependencies:

```bash
make frontend-install
```

Install the Playwright Chromium browser:

```bash
cd desktop/frontend
npx playwright install chromium
```

Generate screenshots from the deterministic fixture build:

```bash
make screenshots
```

The screenshot script:

- builds the frontend
- serves the production build locally
- opens fixture-backed routes for each required screen
- uses a fixed locale, timezone, and reduced-motion profile for deterministic captures
- writes the PNG files into `images/`

No live cluster access is required.

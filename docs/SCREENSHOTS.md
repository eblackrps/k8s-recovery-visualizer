# Desktop Screenshot Workflow

The repository keeps a deterministic desktop screenshot set in `images/` for the main README and supporting docs. The current six-image set reflects the guided operator-console desktop UX shipped in the `v1.9.2` release line.

## Current Screenshot Set

- `images/gui-dashboard.png`
- `images/gui-dashboard.png` now uses the first-run onboarding state (`?view=home&firstRun=1`)
- `images/gui-scan-setup.png` (`?view=scan&scanConnection=api_endpoint&scanStage=connect` guided API endpoint assistant state)
- `images/gui-live-run.png`
- `images/gui-scan-complete.png` (`?view=complete` quieter post-run handoff state)
- `images/gui-results-findings.png`
- `images/gui-compare.png`

These six files are the maintained public screenshot set for the README and documentation. Refresh them whenever the desktop UI, results IA, or landing-page gallery changes.

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

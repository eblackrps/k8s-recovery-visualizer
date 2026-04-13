# Screenshot Workflow

The repository keeps a deterministic screenshot set in `images/` for the main README and supporting docs.

## Current Screenshot Set

- `images/gui-dashboard.png`
- `images/gui-scan-wizard.png`
- `images/gui-live-run.png`
- `images/gui-results-findings.png`
- `images/gui-compare.png`
- `images/report-summary.png`
- `images/report-dr-score.png`

These files are the supported screenshot set for the README and documentation. Refresh them whenever the desktop UI or landing-page gallery changes.

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

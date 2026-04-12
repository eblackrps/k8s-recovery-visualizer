# Screenshot Workflow

The repository keeps the original report screenshots and adds deterministic GUI screenshots beside them in `images/`.

## Report Screenshots Kept In Place

- `images/report-dr-score.png`
- `images/report-summary.png`
- `images/sample-report.png`

## GUI Screenshots

- `images/gui-dashboard.png`
- `images/gui-scan-wizard.png`
- `images/gui-live-run.png`
- `images/gui-results-findings.png`
- `images/gui-compare.png`

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
- writes the PNG files into `images/`

No live cluster access is required.

# Schemas

Published artifact contracts:

- Scan bundle: [`../schemas/recovery-scan-3.1.0.schema.json`](../schemas/recovery-scan-3.1.0.schema.json)
- Enriched bundle: [`../schemas/recovery-enriched-1.2.0.schema.json`](../schemas/recovery-enriched-1.2.0.schema.json)
- Sample scan payload: [`../schemas/examples/recovery-scan-3.1.0.sample.json`](../schemas/examples/recovery-scan-3.1.0.sample.json)
- Sample unverified scan payload: [`../schemas/examples/recovery-scan-3.1.0.unverified.sample.json`](../schemas/examples/recovery-scan-3.1.0.unverified.sample.json)
- Sample enriched payload: [`../schemas/examples/recovery-enriched-1.2.0.sample.json`](../schemas/examples/recovery-enriched-1.2.0.sample.json)

## Versioning Rules

- Additive fields require a schema minor version bump.
- Breaking field removals or incompatible type changes require a major version bump.
- Required-field changes require a major version bump.
- The schema version emitted in JSON must match the published schema file.
- Older schema files stay published so downstream automation can pin to a contract explicitly.

Historical schema files remain in `schemas/` and `schemas/examples/` for compatibility and migration review, even when newer contracts become the default.

## Validation

Use the bundled validator:

```bash
go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.1.0.schema.json -input ./out/recovery-scan.json
go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.2.0.schema.json -input ./out/recovery-enriched.json
```

## 3.1.0 Migration Notes

- Findings now carry prioritization metadata: `title`, `impact`, `effort`, `ownerHint`, `priorityScore`, and `rank`.
- Backup restore simulation now includes namespace `readiness`, readiness counts, estimated data at risk, and top blocking reasons.
- Backup inventory now exports a `drillPlan` for operator-run restore exercises.
- Comparison summaries now include domain, severity, and inventory deltas plus `findingsRegressed`, `findingsImproved`, and `persistentFindingCount`.
- Trend history points now export per-domain scores and finding counts.

CI runs these checks on smoke-test artifacts and committed samples so contract drift is caught before release.

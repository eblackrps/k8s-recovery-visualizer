# Schemas

Published artifact contracts:

- Scan bundle: [`../schemas/recovery-scan-2.2.0.schema.json`](../schemas/recovery-scan-2.2.0.schema.json)
- Enriched bundle: [`../schemas/recovery-enriched-1.1.0.schema.json`](../schemas/recovery-enriched-1.1.0.schema.json)
- Sample scan payload: [`../schemas/examples/recovery-scan-2.2.0.sample.json`](../schemas/examples/recovery-scan-2.2.0.sample.json)
- Sample enriched payload: [`../schemas/examples/recovery-enriched-1.1.0.sample.json`](../schemas/examples/recovery-enriched-1.1.0.sample.json)

## Versioning rules

- Additive fields require a schema minor version bump.
- Breaking field removals or incompatible type changes require a major version bump.
- Required-field changes require a major version bump.
- The schema version emitted in JSON must match the published schema file.
- Older schema files stay published so downstream automation can pin to a contract explicitly.

## Validation

Use the bundled validator:

```bash
go run ./cmd/schema-validate -schema ./schemas/recovery-scan-2.2.0.schema.json -input ./out/recovery-scan.json
go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.1.0.schema.json -input ./out/recovery-enriched.json
```

CI runs these checks on smoke-test artifacts and committed samples so contract drift is caught before release.

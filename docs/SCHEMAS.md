# Schemas

Published artifact contracts:

- Scan bundle: [`../schemas/recovery-scan-2.1.0.schema.json`](../schemas/recovery-scan-2.1.0.schema.json)
- Enriched bundle: [`../schemas/recovery-enriched-1.1.0.schema.json`](../schemas/recovery-enriched-1.1.0.schema.json)

## Versioning rules

- Additive fields require a schema minor version bump.
- Breaking field removals or incompatible type changes require a major version bump.
- The schema version emitted in JSON must match the published schema file.

## Validation

Use the bundled validator:

```bash
go run ./cmd/schema-validate -schema ./schemas/recovery-scan-2.1.0.schema.json -input ./out/recovery-scan.json
go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.1.0.schema.json -input ./out/recovery-enriched.json
```

CI runs these checks on smoke-test artifacts so contract drift is caught before release.

# Risks and Challenges

Probability/impact are initial qualitative ratings and must be reviewed each milestone.

| Risk | P/I | Mitigation | Detection | Recovery |
|---|---|---|---|---|
| Double booking from races | M/Critical | Partial GiST exclusion constraint; transaction tests | Constraint/conflict metrics and invariant query | Stop affected run sales, reconcile/audit, contact passengers |
| Incorrect route ordering | M/High | Unique positions, monotonic-distance validation, versioned approval | Seed/migration tests; admin audit | Disable route/run; correct new version and revalidate |
| Fare miscalculation | M/High | Integer minor units, versioned rules/snapshots, examples | Golden tests and reconciliation | Freeze sales/rule; issue corrected refunds manually/future flow |
| Stale availability | High/Medium | Label snapshot; constraint at commit; short cache | Conflict-rate increase | Refresh/suggest another seat |
| Time-zone mistakes | M/High | UTC instants, explicit Asia/Colombo service date, boundary tests | Schedule anomaly alerts/tests | Correct future runs; audited passenger communication |
| Booking-reference exposure | M/High | Entropy + second factor, rate limits, redaction | Failed lookup/anomaly metrics | Rotate access token/reference, notify/investigate |
| Migration failure | M/High | CI upgrade tests, expand/contract, backups | Migration/readiness gate | Halt rollout; forward-fix or tested rollback/restore |
| Seed inconsistency | M/Medium | Deterministic IDs, idempotent seed, constraints | Seed-twice CI and checksums | Reset disposable env or corrective seed migration |
| Database contention | M/High | Short transactions, indexes, pool limits/timeouts | Lock/pool/p95 metrics | Shed load, tune/index, scale DB; partition only with evidence |
| Incorrect hold expiry | M/High | UTC deadline, conditional idempotent worker, audits | Expiry-lag and overdue-held query | Rerun worker; release/correct through audited transitions |
| Network retry duplicates | High/High | Required idempotency key + request hash/response | Duplicate-key/conflict metrics | Return stored result; reconcile anomalous duplicates |
| Admin configuration error | M/High | RBAC/MFA, validation, preview, maker-checker/versioning | Audit/config canaries | Disable/revert via new version; assess bookings |
| Frontend/backend drift | M/Medium | Normative OpenAPI, generated client, contract CI | Generation diff/E2E failures | Pin compatible release; regenerate/fix contract |
| Scope expansion | High/Medium | Phase gates and explicit out-of-scope list | Milestone burn-up/review | Defer extras; restore core acceptance priorities |

Operational owners, due dates and numeric risk scores are assigned when a team is staffed. Critical invariant breaches trigger an incident regardless of transaction volume.

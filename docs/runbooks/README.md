# Operational Runbooks

These procedures are for incidents and release failures. Adapt commands and approval requirements to the environment; never treat local destructive commands as production-safe defaults.

- [Database recovery](database-recovery.md)
- [Migration failure](migration-failure.md)
- [Booking conflicts](booking-conflicts.md)
- [Service unavailable](service-unavailable.md)
- [Release rollback](release-rollback.md)

Every incident should preserve evidence, avoid passenger data in shared channels, identify an owner, record timestamps in UTC, and produce corrective actions after recovery.

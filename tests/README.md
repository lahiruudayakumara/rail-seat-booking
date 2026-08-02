# Cross-service tests

Package-level Go and React tests remain beside their source. This directory contains tests that cross application or process boundaries.

- `integration/booking-flow.sh` exercises the running API from discovery through cancellation. It requires `curl`, `jq`, and migrated/seeded services.
- `load/booking-contention.js` uses [k6](https://k6.io/) to send 12 overlapping booking attempts and requires exactly one success. Run it only against a disposable database because the winning booking remains confirmed.

Start the stack before running integration tests:

```bash
docker compose up --build -d
make smoke
make integration
```

The k6 test requires IDs and an unexpired fare quote created for the same segment and seat. See the environment variables validated in the test's `setup` function.

# API examples

These `.http` files use the VS Code REST Client `{{$dotenv NAME}}` variables. Copy `.env.example` to `.env`, start the services, and run `catalog.http` to obtain the current train-run and seat IDs. Put response IDs and management tokens in your ignored local `.env` before running each dependent request.

Recommended order:

1. `health.http`
2. `catalog.http`
3. `booking-flow.http`: availability → quote → hold → booking → payment → lookup/cancellation
4. `waitlist-flow.http`: use a train segment with no matching availability

JetBrains HTTP Client users can provide the same variable names through a private HTTP client environment file and replace the `@… = {{$dotenv …}}` declarations with their environment-variable syntax.

All committed identities and passenger details are demonstration values. `.env` is ignored by Git; never paste production credentials or real passenger information into committed examples.

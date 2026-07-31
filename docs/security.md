# Security Design

## Identity and access

Authentication is an explicit initial placeholder, not a claim of security completion. Public search/quote/create routes use abuse controls. Future OIDC/OAuth 2.1 access tokens identify passengers and staff; authorization is deny-by-default with owner, support and `ADMIN` policies checked in services. Admin actions require MFA, short sessions and audited least privilege. Never accept a role supplied by the client.

Guest management uses a random booking reference **plus** a separate signed access token or verified contact challenge. References use high entropy, are case-normalized, rate-limited and never reveal existence through distinguishable auth errors. UUIDs reduce sequential enumeration but are not authorization.

## Controls

| Area | Design |
|---|---|
| Passenger data | Collect minimum; encrypt in transit/at rest; restrict fields/roles; define retention/anonymization; no analytics payloads. |
| Input | Body limits, content-type checks, strict JSON/Zod/OpenAPI validation, normalization and output encoding. |
| SQL injection | sqlc/pgx bound parameters; no concatenated user SQL; allowlist dynamic sort/filter identifiers. |
| Abuse | Gateway/app rate limits by trusted client/account/reference; generic lookup responses; quotas, anomaly metrics and optional CAPTCHA after risk signals. |
| CORS/CSRF | Exact origins and minimal headers/methods. Bearer headers are less CSRF-prone; cookie auth requires SameSite, CSRF token and origin checks. |
| Headers | HSTS in production, CSP, `nosniff`, Referrer-Policy, Permissions-Policy and frame denial/`frame-ancestors`. |
| Transport | TLS 1.2+ externally and encrypted DB links; redirect HTTP; validate certificates. |
| Logging | Redact names/contact, bodies, tokens, cookies, references and idempotency keys; restrict and retain logs by policy. |
| Audit | Append-only booking/admin events with actor, request ID and before/after safe metadata; alert on gaps/tampering. |
| Database | Separate migration/runtime/read-only roles; no public endpoint; network allowlist; runtime cannot alter schema/audit history. |
| Containers | Multi-stage minimal pinned images, non-root/read-only FS where possible, dropped capabilities, resource limits and image/SBOM scanning. |
| Backups | Encrypted, access-logged, retention-limited, separate account/region where needed; restore tests; deletion policy applies. |

Secrets must **never be committed**. `.env.example` contains safe placeholders only. Local `.env` is ignored; production uses a secret manager, scoped workload identity, rotation and startup injection—never image build arguments. Immediately revoke/rotate and history-clean any exposed value.

## Verification and response

CI runs secret, SAST, dependency/license, Go/npm vulnerability and container scans; scheduled DAST and penetration testing precede production. Patch SLAs follow severity. Threat-model booking races, IDOR, reference brute force, mass reservation, input exhaustion, admin compromise and supply chain. Incident procedure: contain credentials/access, preserve audit evidence, assess passenger impact, restore/rotate, notify under applicable law and complete a blameless review.

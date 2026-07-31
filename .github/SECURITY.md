# Security Policy

## Reporting a vulnerability

Do not report security vulnerabilities in public issues or discussions. Use GitHub's **Security → Report a vulnerability** flow so maintainers can investigate privately.

Include the affected commit or version, impact, reproduction steps, and a minimal proof of concept. Remove real passenger data, credentials, booking references, tokens, and production logs.

Maintainers should acknowledge a report within three business days, provide an initial assessment when enough information is available, and coordinate disclosure after a fix is ready. Timelines may vary with severity and complexity.

## Supported versions

Security fixes target the latest commit on `main`. The `dev` branch is pre-release and may change without compatibility guarantees.

## Scope priorities

High-priority areas include double booking, authorization bypass, personal-data exposure, SQL injection, idempotency bypass, fare manipulation, secret disclosure, and vulnerable build or deployment dependencies.

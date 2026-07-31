# ADR 0009: Use Docker Compose for local development

**Status:** Accepted

## Context

New developers need reproducible API, web and PostgreSQL startup with health/migration ordering.

## Decision

Provide `docker compose up --build` as the expected one-command environment, with named DB volume and explicit migration/seed jobs.

## Consequences

Setup becomes consistent and production images are exercised early, but requires Docker resources and file-watching/platform tuning. Native commands remain useful for fast iteration.

## Alternatives Considered

Manual installs are lightweight but drift-prone. Kubernetes is excessive locally. Dev Containers/Nix improve reproducibility but add another adoption layer; they may complement Compose later.

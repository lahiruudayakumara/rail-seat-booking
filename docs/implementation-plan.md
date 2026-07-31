# Implementation Plan

Each phase ends with a demonstrable acceptance gate; core correctness precedes optional features.

## 1. Foundation

Create monorepo/toolchain, Compose, Go/Chi API and React/Vite shells, PostgreSQL, health/readiness, environment validation, Make targets and CI foundation. Gate: one-command setup reaches healthy URLs and clean checks.

## 2. Core domain and schema

Implement routes/stations ordering, train/run, coach/seat configuration, Goose migrations, sqlc queries and idempotent demo seed (one train, multiple runs, configurable three reserved/five unreserved coaches, classes/rules). Gate: schema/seed tests and no hardcoded topology.

## 3. Availability and fares

Implement endpoint membership/positions, overlap-aware availability, effective fare rules, integer arithmetic/snapshot quotes and OpenAPI endpoints/client. Gate: adjacent/overlap query cases and fare golden tests.

## 4. Booking correctness

Implement booking/idempotency transaction, exclusion constraint mapping, reference generation, cancellation/audit, integration and 10+ contender concurrency tests. Gate: exactly one overlapping winner; replay never duplicates; cancellation releases inventory.

## 5. Frontend booking flow

Build journey/run selection, accessible seat map, quote/review, passenger form, confirmation/manage flow and 409 refresh/preservation behavior. Gate: Playwright happy path, conflict path, keyboard/mobile checks.

## 6. Quality and readiness

Complete E2E/migration/load/security tests, structured logs/metrics, graceful failure/error UX, docs/contract validation, accessibility and threat review. Gate: required CI, runbooks, restore evidence and submission checklist.

## 7. Optional extras

Only after Phase 6: holds/expiry, payment adapter, waitlist, admin dashboard, revenue/occupancy analytics, real-time availability and notification/outbox integration. Each requires its own ADR/threat model and cannot weaken database inventory integrity.

# ADR 0008: Use React, TypeScript and Vite

**Status:** Accepted

## Context

The booking flow needs accessible interactive forms, server-state caching and a productive component ecosystem.

## Decision

Use React/TypeScript/Vite with TanStack Query, React Hook Form, Zod, Tailwind CSS and shadcn/ui.

## Consequences

The stack provides fast builds and clear state/form tools, while requiring discipline around dependency weight, accessibility and avoiding duplicated server state. It is an SPA initially, so SEO/server rendering is limited.

## Alternatives Considered

Next.js adds SSR/full-stack conventions not needed for this API-centric application. Vue/Svelte offer excellent alternatives but change team/tool assumptions. A server-rendered Go UI simplifies deployment but makes the rich seat interaction less ergonomic.

# Project Spec : AI Assistants Catalog

## Objective
- Build a fullstack web app for managing, browsing, and running AI assistants with a mockable LLM provider

## Tech Stack
- Backend: Go, PostgreSQL, REST API, JWT, Makefile
- Frontend: React 19, TypeScript, Vite, TanStack Query, shadcn/ui, pnpm
- Infra: Docker Compose, OpenAPI, k6

## Commands
- Run all: `docker-compose up --build`
- Backend build: `make -C backend build` (must pass before handoff)
- Backend test: `make -C backend test` (must pass after each completed backend feature)
- Backend lint: `make -C backend lint` (must pass after each completed backend feature)
- Frontend install: `pnpm -C frontend install`
- Frontend build: `pnpm -C frontend build` (must pass before handoff)
- Frontend test: `pnpm -C frontend test` (must pass after each completed frontend feature)
- Frontend lint: `pnpm -C frontend lint` (must pass after each completed frontend feature)

## Project Structure
- `api.yaml` - OpenAPI contract and source of truth for API DTOs
- `backend/` - Go service, migrations, tests, LLM providers
- `frontend/` - React SPA, API client, routes, UI components
- `load-tests/` - k6 load testing scripts
- `docker-compose.yaml` - Local app, database, and service orchestration
- `README.md` - Setup, decisions, testing, and architecture notes

## Boundaries
- Always: Keep backend/frontend API behavior aligned with `api.yaml`, in case of conflict between definition and user request let user know
- Always: Preserve role checks and do not expose `systemPrompt` to regular users
- Always: Keep code strictly typed; do not use `any`, implicit `any`, unchecked casts, type assertions to bypass errors, or similar type escapes unless the user explicitly approves the exception
- Ask first: Database schema changes after migrations are established
- Ask first: Adding large dependencies or changing the chosen stack
- Never: Commit secrets, real API keys, `.env`, `node_modules/`, or generated build output

# Project Spec : AI Assistants Catalog

## Objective
- Build a fullstack web app for managing, browsing, and running AI assistants with a mockable LLM provider

## Tech Stack
- Backend: Go, PostgreSQL, REST API, JWT, Makefile
- Frontend: React 19, TypeScript, Vite, TanStack Query, shadcn/ui, pnpm
- Infra: Docker Compose, OpenAPI, k6

## Commands
- Run all: `docker-compose up --build`
- Backend build: `make -C backend build`
- Backend test: `make -C backend test`
- Backend lint: `make -C backend lint`
- Frontend install: `pnpm -C frontend install`
- Frontend build: `pnpm -C frontend build`
- Frontend test: `pnpm -C frontend test`
- Frontend lint: `pnpm -C frontend lint`

## Project Structure
- `api.yaml` - OpenAPI contract and source of truth for API DTOs
- `backend/` - Go service, migrations, tests, LLM providers
- `frontend/` - React SPA, API client, routes, UI components
- `load-tests/` - k6 load testing scripts
- `docker-compose.yaml` - Local app, database, and service orchestration
- `README.md` - Setup, decisions, testing, and architecture notes

## Boundaries
- Always: Keep backend/frontend API behavior aligned with `api.yaml`
- Always: Preserve role checks and do not expose `systemPrompt` to regular users
- Ask first: Database schema changes after migrations are established
- Ask first: Adding large dependencies or changing the chosen stack
- Never: Commit secrets, real API keys, `.env`, `node_modules/`, or generated build output

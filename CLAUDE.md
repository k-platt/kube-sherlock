# CLAUDE.md

Operating manual for agents working in this repo. Read it fully before changing anything.

## Product

Kube Sherlock is an AI-assisted Kubernetes troubleshooting tool. A user pastes an error or asks a
natural-language question ("what is the health of my pods in kube-system?"); the Go backend reads
live, **read-only** cluster data via client-go and asks Google Gemini to explain it. It ships as:

- a **Go backend** (`backend/`): CLI (`analyze`) plus an HTTP API (`server`) built on gin;
- a **Next.js 15 frontend** (repo root, `src/`) that calls the backend API.

## Architecture

```
backend/                     Go module "kube-sherlock" (go 1.21)
  main.go                    calls cmd.Execute()
  cmd/                       cobra commands: root.go (config/viper/zap init),
                             server.go (HTTP server + graceful shutdown), analyze.go (CLI)
  internal/config/           Config struct built from viper; global logger accessor
  internal/api/              router.go (gin routes, CORS middleware, service wiring),
                             handlers.go (request/response types + handlers)
  internal/ai/               Gemini client: TroubleshootError, SuggestResources,
                             SummarizeResourceData, QueryWithMCP (tool calling)
  internal/kubernetes/       client-go wrapper: GatherResources, GetPodLogs (read-only)
  internal/mcp/              "MCP" tool registry (pod health, deployments, services,
                             events, pod logs) that the AI service can invoke
  .kube-sherlock.yaml.example  sample config file (~/.kube-sherlock.yaml)
  kube-sherlock              committed build artifact — do not rebuild/commit it

src/                         Next.js app (TypeScript, strict)
  app/                       layout.tsx, page.tsx, globals.css (App Router)
  components/kube-sherlock.tsx  the main UI (forms, query modes, markdown rendering)
  components/ui/             shadcn/ui primitives (generated; edit sparingly)
  hooks/                     use-toast, use-mobile
  lib/api-client.ts          typed fetch client for the Go backend (NEXT_PUBLIC_API_URL,
                             default http://localhost:8080)
  lib/*-backend.ts, mcp-query.ts  thin wrappers used by the UI, one per backend endpoint
  ai/                        legacy Firebase Genkit flows (not used by the UI any more)
```

HTTP API (`backend/internal/api/router.go`): `GET /health`, `POST /api/troubleshoot`,
`/api/suggest-resources`, `/api/summarize`, `/api/gather-resources`, `/api/query`.

Where things go:
- New endpoint → handler + request/response types in `internal/api/handlers.go`, route in
  `router.go`, business logic in the relevant `internal/*` service, client method in
  `src/lib/api-client.ts`.
- New cluster tool → register in `internal/mcp/service.go` `registerTools()`, dispatch in
  `ExecuteTool`, cluster access via `internal/kubernetes`.
- Keep request/response JSON field names in Go structs and `src/lib/api-client.ts` in sync.

## Commands

All from the repo root unless noted.

| Task | Command |
|---|---|
| **Full quality gate (must pass before you finish)** | `scripts/check` |
| Install frontend deps | `npm ci` |
| Frontend dev server (port 9002) | `npm run dev` |
| Frontend type check | `npm run typecheck` |
| Frontend production build | `npm run build` |
| Backend build | `cd backend && go build -o /tmp/kube-sherlock .` |
| Backend server | `cd backend && go run . server --port 8080 --gemini-api-key ...` |
| Backend CLI | `cd backend && go run . analyze "<error text>"` |
| Backend format / vet | `cd backend && gofmt -l . && go vet ./...` |
| All Go tests | `cd backend && go test ./...` |
| Single Go test | `cd backend && go test ./internal/api -run '^TestName$' -v` |
| Security scan | `scripts/audit` |

`scripts/check` runs: `npm ci`, `go mod download`, gofmt check, `go vet`, `tsc --noEmit`,
`next build`, `go test ./...`. It does **not** run `npm run lint` (no ESLint config exists yet;
`next lint` would start an interactive setup) and there is no frontend test runner yet.
Do not set up ESLint, Prettier or a JS test runner unless your issue asks for it.

## Coding conventions (as practiced here)

Go:
- gofmt formatting; `go vet` clean. Doc comment on every exported type/function
  (`// Name does ...`).
- Services are structs built by `NewXxx(..., logger *zap.Logger)` constructors; dependencies are
  passed in, not looked up globally (except `config.GetConfig()`/`GetLogger()` in `cmd/`).
- Errors: wrap with context using `fmt.Errorf("failed to <verb>: %w", err)` and return; log with
  structured zap fields (`zap.Error(err)`, `zap.String(...)`) at the point of handling.
- HTTP handlers: `c.ShouldBindJSON(&req)` with `binding:"required"` tags → 400 with the bind
  error; service failure → log the real error, return 500 with a **generic** message
  (`gin.H{"error": "Failed to ..."}`), never the raw internal error; missing optional dependency
  (e.g. no cluster) → 503. Every request takes `c.Request.Context()` down to the services.
- A nil `k8sService` means "no cluster available" and must be handled, not panicked on.
- CLI fatal config errors: `fmt.Fprintf(os.Stderr, ...)` + `os.Exit(1)` (in `cmd/` only).

TypeScript / React:
- `strict` TS, `@/` path alias for `src/`. Function components, react-hook-form + zod for forms,
  shadcn/ui components from `@/components/ui`, Tailwind for styling, `useToast` for user errors.
- Backend calls go through `kubeSherlockAPI` in `src/lib/api-client.ts`, which throws
  `Error("API request failed: <status> <body>")` on non-2xx; callers catch and show a toast.

## Testing rules

- Every behavior change needs tests. Bug fixes start with a failing test that reproduces the bug.
- Go tests live next to the code (`foo_test.go`, same package), use the standard `testing`
  package; table-driven where it fits. For handlers use `httptest` + `gin.SetMode(gin.TestMode)`.
  For Kubernetes use `k8s.io/client-go/kubernetes/fake` (refactor behind an interface if needed);
  never a real cluster.
- Tests must not need network access, a real Kubernetes cluster, a kubeconfig, or a real Gemini
  API key. Stub the AI client and cluster.
- Never delete, skip, or weaken existing tests or assertions to make a change pass.
- `scripts/check` must pass before you finish.

## Security

- **Secrets**: the Gemini API key comes from `--gemini-api-key`, the `~/.kube-sherlock.yaml`
  config (`gemini.api_key`), or the environment — never hard-code keys, never commit `.env*`,
  kubeconfigs, `*.pem` or real config files. Use obvious placeholders in examples and tests
  (`test-key`). Never log API keys, tokens, kubeconfig contents, or `Authorization` headers.
  Do not add credentials to code that the frontend bundles (`NEXT_PUBLIC_*` is public).
- **Input validation**: all API input is untrusted. Bind into typed structs with `binding` tags;
  validate namespaces, pod/container names, label selectors and resource types before calling
  the cluster; bound numeric inputs (e.g. log line counts) and request sizes. Treat LLM output as
  untrusted too: the AI may only invoke tools from the MCP registry, with arguments validated the
  same way as user input.
- **Dangerous operations**: the backend must stay **read-only** against the cluster (get/list,
  logs). Never add create/update/patch/delete/exec/port-forward capability, or let the AI trigger
  one, unless an issue explicitly asks for it — and then it needs an explicit user confirmation
  step and tests. Pod logs and resource data can contain secrets: never return `Secret` objects
  and do not widen what is gathered without an issue asking for it. Do not loosen the server bind
  address (default `localhost`) or CORS policy without an issue asking for it.
- **Dependencies**: do not add new dependencies unless the issue requires it; prefer the standard
  library and what is already in `go.mod` / `package.json`. If you must add one, pick a maintained,
  widely-used package, pin it via the lockfile (`go.sum`, `package-lock.json` — commit both), and
  keep `scripts/audit` free of new findings. Never use `npm install --force`/`--legacy-peer-deps`
  to paper over conflicts.

## Boundaries

- Never edit `.github/`, `.factory/`, `scripts/check` or `scripts/audit`.
- Do not commit build artifacts (`.next/`, `.bin/`, rebuilt `backend/kube-sherlock`).
- Keep changes scoped to the issue: no drive-by refactors, reformatting of untouched files,
  dependency upgrades, or edits to generated `src/components/ui/*` unless required.
- Do not change public API shapes (routes, JSON fields, CLI flags) unless the issue asks.

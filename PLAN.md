# Duo Streak Widget — Implementation Plan (Cloud Run + Cloudflare)

This document is an improved, actionable plan for building a small service that returns an 88×31 SVG badge showing a Duolingo streak for a username. It expands the original plan to include template variations, testing, theme support, automation for deployment, and troubleshooting notes for Nix/Go environments.

## Architecture Overview

User Browser → Cloudflare (edge cache) → Cloud Run service (Go) → In-memory cache (app-level) → Duolingo API

- Cloud Run: stateless Go service that generates an SVG badge at `GET /badge/{username}.svg`.
- Cloudflare: free plan provides global caching and HTTP/edge rules to reduce hits to Cloud Run.
- In-memory cache (within the app): TTL-based cache (1–4 hours typical) to avoid frequent Duolingo calls on cache miss.

## Goals / Success Criteria

- Badge endpoint: `/badge/{username}.svg` returns an 88×31 SVG with streak number.
- Cache hits served from Cloudflare edge or from app cache; Duolingo API is called only on cache miss + after TTL.
- Reasonable default TTLs and headers so browsers + Cloudflare cache effectively.
- Rate limiting to protect the upstream Duolingo endpoint.

Additional success items:

- Well-documented template system that supports multiple variants (e.g. `duoText`, `duoIcon`, `compact`).
- Simple theming support via query params or path variant (e.g., `?theme=dark`, or `/badge/{username}.svg?variant=compact`).
- Automated tests — unit tests for generator, integration tests for the HTTP path, and a small E2E smoke test in CI.

## Phases (high level)

1. Go backend — core logic (HTTP server, API client, cache, badge generation, rate limiting).
2. Containerize — Dockerfile, `.dockerignore`, and local dev run.
3. Infrastructure — Terraform to deploy to Cloud Run and Artifact Registry.
4. Cloudflare — DNS, Cache Rules and Page Rules to cache `/badge/*.svg` at edge.
5. CI/CD — GitHub Actions to build, test, push, and deploy.
6. Observability — logs, metrics, uptime.

7. Templates & Theming — small library of SVG templates and a testing mechanism for visual snapshots.

Tip: If you plan to use Cloudflare Workers for on-edge logic, only perform cheap, idempotent transformations (strip query params, rewrite headers). Keep Duolingo calls to the origin.

---

## Phase 1 — Go Backend (developer tasks)

Contract

- Input: HTTP GET /badge/{username}.svg
- Output: `Content-Type: image/svg+xml` body with 88×31 SVG or a lightweight error badge.
- Support additional optional query parameters:
  - `variant` — selects one of the embedded SVG templates
  - `theme` — selects light/dark color palette or a CSS class for the SVG
  - `width`/`height` — optional scale constraints (server should clamp/deny large values)
- Error modes: upstream failure -> show fallback badge (N/A); not-found -> friendly badge.

Files to create (suggested structure)

```text
src/
├─ main.go                    # router and entry point
├─ internal/
│  ├─ duolingo/client.go      # Duolingo API calls
│  ├─ cache/cache.go          # in-memory TTL cache (thread-safe)
│  ├─ badge/generator.go      # SVG generation
│  └─ ratelimit/limiter.go    # per-username rate limiter
```

Key behaviors

- Cache key: `streak:{username}`. TTL default: 2 hours (configurable via env var).
- HTTP headers from service: `Cache-Control: public, max-age=300, s-maxage=900` (adjustable).
- ETag or simple hash may be returned to help conditional requests.
- Rate limiting: per-username token bucket or fixed-window (e.g., 1 Duolingo fetch per 5 minutes per username).

Implementation suggestions & best practices:

- Use `singleflight` (golang.org/x/sync/singleflight) to deduplicate concurrent upstream calls for the same username. This reduces load when a cache expires and many requests arrive at once.
- Keep the generator deterministic for snapshot tests: pass numeric format only (no truncation/abbreviation in test mode).
- Add an in-memory `map[string]struct{ Value int; Expires time.Time }` behind a read/write mutex or use an LRU TTL cache library.

Edge cases

- Missing username / invalid characters -> return 400 or sanitized fallback.
- Duolingo returns 404 -> show "user not found" badge and cache short (5–15 minutes).
- Large number of concurrent requests for same user -> use singleflight or similar to coalesce upstream calls.

Badge-specific edge cases:

- Username contains path separators or control characters — sanitize and return 400 for invalid characters to prevent cache poisoning.
- Upstream rate-limit errors — return a transient placeholder (N/A) with short cache TTL (e.g., 60s) and log with detail.
- Internationalized usernames — limit accepted characters and normalize (NFKC) before using cache keys.

## Phase 2 — Containerization

- Multi-stage `Dockerfile` (build in golang image, copy binary to small base like `gcr.io/distroless/static` or `alpine`).
- `.dockerignore` to exclude `.git`, `node_modules`, etc.
- For local dev: `docker-compose` or `skaffold`/`telepresence` optional.

Local dev ergonomics:

- Add a `Makefile` target like `make dev` to run `nix develop -c go run ./src` with the correct env vars.
- Add `docker-compose.yml` with a local Redis option if you choose to use Redis instead of in-memory cache for scale testing.

## Phase 3 — GCP Infra (Terraform) (optional: can skip if you prefer manual deploy)

Minimal resources:

- Artifact Registry (docker repo)
- Cloud Run service (revision settings: concurrency, CPU, memory)

Notes:

- Cloud Run public URL used behind Cloudflare.
- No need for load balancer; Cloud Run URL is fine to front with Cloudflare.

## Phase 4 — Cloudflare configuration

Essential steps:

1. Create a DNS `CNAME` pointing your domain to the Cloud Run custom domain or use Cloudflare's proxy in front of the Cloud Run URL.
   - If mapping an apex domain, configure Cloud Run custom domain + Cloudflare DNS records as needed.
2. Add Cache Rule / Page Rule for the badge endpoint:
   - Pattern: `*yourdomain.com/badge/*.svg`
   - Settings: Cache Level: Standard, Edge Cache TTL: 15 minutes or 1 hour (choose based on how often you want updates), Browser Cache TTL: 5 minutes.
3. Optionally set a Cloudflare Worker or Transform Rule to strip query params that are not cache-relevant.
4. Optionally configure Rate Limiting rules (Cloudflare has a limited free quota) for abusive traffic.

Recommended headers from the service (Cloud Run):

```http
Cache-Control: public, max-age=300, s-maxage=900
Vary: Accept
Content-Type: image/svg+xml; charset=utf-8
```

- `s-maxage` helps CDN caches; `max-age` helps downstream (browsers).

## Phase 5 — CI/CD

   1. `lint` and `go test`.
   2. `docker build` and push to Artifact Registry.
   3. Add a small "deploy" job that pushes a tagged image and performs `gcloud run deploy` or `terraform apply` depending on your infra setup. Staging and production branches are useful here.
   4. `gcloud run deploy` (or use `terraform apply`).

Secrets: store GCP credentials / service account in GitHub Secrets.

## Phase 6 — Observability

- Log structured events: cache hit/miss, upstream call, error, username.
- Monitor Cloud Run metrics (requests, error rate, latencies).
- Optional uptime check (UptimeRobot or Cloudflare Healthchecks).

---

## Observability, Metrics, & Ops

- Instrument request latencies, upstream retries, cache hit/miss rates, and error counts.
- Expose a metrics endpoint for Prometheus scraping (or push to Cloud Monitoring).
- Add structured logs: username, upstream status, cache hit flag.

---

## Security & Privacy

---

## Templates & Theming

- Keep templates in `src/templates/*.svg`, embed via `//go:embed`.
- Standardize placeholder: `{{ .Streak }}` — tests should assert exact numeric output (no truncation).
- Provide multiple pre-validated variants (`duoText`, `duoIcon`, `compact`) with a `variant` query param.
- Themes (`light`/`dark`) should be implemented by toggling CSS class or `data-theme` attribute; keep the engine simple and deterministic.

---

## Testing & CI updates

- Unit tests — generator and template render assertions in `internal/badge`.
- Integration tests — test the HTTP handler with mock Duolingo client and app cache.
- E2E smoke — run container in CI and assert the badge endpoint returns an SVG and correct headers.

Sample GitHub Actions workflow:

```yaml
name: CI
on: [push, pull_request]
jobs:
   test:
      runs-on: ubuntu-latest
      steps:
         - uses: actions/checkout@v3
         - name: Setup Go
            uses: actions/setup-go@v4
            with:
               go-version: 1.20
         - name: Install deps
            run: go mod download
         - name: Run unit tests
            run: go test ./... -v
         - name: E2E smoke
            run: |
               docker build -t duo-streak-widget:e2e .
               docker run --rm -p 8080:8080 duo-streak-widget:e2e &
               sleep 2
               curl -f http://localhost:8080/badge/testuser.svg | grep -q "<svg"
```

## Badge design and API

- Endpoint: `GET /badge/{username}.svg` (optionally allow query params like `?theme=dark`)
- SVG: 88×31px, use scalable vector font or simple shapes to keep response small.
- Fallbacks: `N/A` or flame emoji if upstream fails.

Example cache policy and behavior

- Cloudflare edge TTL: 15 minutes
- Service `s-maxage`/`max-age`: `s-maxage=900, max-age=300`
- App cache TTL (in-memory): 2 hours

This yields:

- Edge caches serve most requests instantly.
- On edge miss, Cloud Run serves from its in-memory cache if available.
- Duolingo API is contacted only on app cache miss or after TTL.

## Security and Privacy

- Do not store or expose any Duolingo credentials. The service only reads public user data.
- Sanitize and rate-limit user-supplied path segments to prevent injection or cache poisoning.

Additional checklist:

- Use environment variables for credentials. Keep Cloud Run service account permissions minimal.
- Pin third-party dependencies to a known-good set in `go.mod` and `go.sum`.
- Consider creating a separate Cloud Run service account only for fetching the Duolingo endpoints.

---

## Visual tests and snapshotting

- Use `testdata/` to store example SVG outputs for a couple of fixed streak values.
- Optionally add a small script to render those SVGs in a simple HTML file for quick manual visual review (CI can also archive artifacts).

---

## Troubleshooting & Nix notes

- If `nix develop -c go test` fails with `go: download go1.x: toolchain not available`, try running `nix develop` and then `go test .` inside the shell.
- Keep CI pinned to a stable Go version so local shells and CI match.

## Cost considerations

- Cloud Run free tier (2M requests/month) should be enough for learning.
- Cloudflare free tier provides global caching without CDN egress charges.
- Use small memory and CPU limits for Cloud Run to minimize cost.

## Local dev quick commands (suggested)

```bash
# run locally
go run ./src

# build container
docker build -t duo-streak-widget:dev .

docker run -p 8080:8080 duo-streak-widget:dev

# test badge
curl -v http://localhost:8080/badge/someusername.svg
```

## Project file layout suggestion

```
/ (repo root)
├── PLAN.md                # This file
├── src/
│   ├── cmd/               # server entry
│   ├── internal/
│   │  ├── badge/          # generator + templates
│   │  ├── cache/          # TTL cache
│   │  └── duolingo/       # client + mocks for tests
├── Dockerfile
├── .dockerignore
├── iac/                   # terraform and deployment config
└── .github/workflows      # CI/CD
```

## Learning outcomes

By implementing this yourself you'll learn:

- HTTP caching strategies and cache headers
- Server-side TTL caching and coalescing upstream calls
- Cloud Run deployment and Docker container builds
- Basic Terraform (optional) and Cloudflare configuration
- Observability basics and lightweight rate limiting

---

If you'd like, I can:

- Stop here and let you implement the code yourself (I won't add more changes), or
- Bootstrapping: add skeleton files and a minimal `main.go` to help you get started quickly.

You asked to do the programming yourself — I created this plan and left the repo untouched except for this file. Ready to continue how you prefer.

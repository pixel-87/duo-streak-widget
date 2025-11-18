# Duo Streak Widget — Hexagonal Architecture Plan

This plan keeps the service small and readable while applying Hexagonal Architecture (a.k.a. Ports & Adapters). The goal is to separate core streak logic from delivery details so future changes (new badge templates, alternate upstream APIs, CLI tools) stay easy.

## Guiding Principles
- Favor tiny packages with obvious names over elaborate frameworks.
- Keep business logic in plain Go structs; interfaces live at the boundaries only.
- Stick to one happy-path per function; panics and hidden magic are out.
- Testing should run fast (`go test ./...`) without external services.

## Hexagonal Architecture at a Glance
```
┌────────────┐      ┌──────────────┐      ┌──────────────┐
│ HTTP Port  │◀────▶│ Application  │◀────▶│ Domain Model │
└────────────┘      │  (Use Cases) │      └──────────────┘
      ▲             └──────────────┘              ▲
      │                     ▲                     │
      │                     │                     │
┌────────────┐      ┌──────────────┐      ┌──────────────┐
│ Cache Adap │◀────▶│ Ports (IFCs) │◀────▶│ Duolingo API │
└────────────┘      └──────────────┘      └──────────────┘
```
- **Domain**: streak value, badge text, simple formatting helpers.
- **Application layer**: orchestrates fetch → cache → template render; exposes ports.
- **Adapters**: HTTP handler, Duolingo client, cache, badge renderer.

## Proposed Project Layout
```text
src/
├─ cmd/server/main.go            # wires adapters + application
├─ internal/
│  ├─ domain/                    # streak entity, badge model
│  │   └─ streak.go
│  ├─ app/                       # use cases (FetchBadge)
│  │   └─ service.go
│  ├─ ports/                     # interfaces shared between app & adapters
│  │   ├─ cache_port.go          # Get/Set
│  │   ├─ duolingo_port.go       # FetchStreak
│  │   └─ render_port.go         # RenderBadge
│  └─ adapters/
│      ├─ http/handler.go        # maps HTTP → app service
│      ├─ duolingo/client.go     # real HTTP client
│      ├─ cache/memory.go        # TTL map (can swap later)
│      └─ badge/svg.go           # template renderer using //go:embed
└─ templates/duo.svg             # SVG skeleton(s)
```
Everything still compiles into one binary; the folders simply reflect responsibilities.

## Request Flow (Happy Path)
1. HTTP adapter validates the username and variant query, then calls `app.Service.FetchBadge(ctx, username, variant)`.
2. Application service checks the cache port: return hit if fresh.
3. On miss, service calls the Duolingo port to load the streak, updates cache, and prepares `domain.Badge`.
4. Service invokes the badge renderer port to produce SVG bytes.
5. HTTP adapter sets headers (`Content-Type`, cache hints) and writes the SVG.

## Layer Responsibilities
### Domain
- `Streak` struct (username, count, updatedAt) with helpers like `ClampToZero()`.
- `Badge` struct (streak string, variant name).
- No networking, no logging.

### Application (Use Cases)
- `Service` holds ports:
  ```go
  type Service struct {
      Cache    ports.Cache
      Duo      ports.Duolingo
      Renderer ports.BadgeRenderer
      TTL      time.Duration
  }
  ```
- Methods: `FetchBadge(ctx, username, variant string) ([]byte, error)`.
- Contains simple policies: TTL duration, negative streak guard, default variant.

### Ports & Adapters
- **Cache adapter**: start with an in-memory TTL map + `sync.RWMutex`. Later swap to Redis by implementing the same interface.
- **Duolingo adapter**: thin HTTP client wrapping `net/http` with timeout + JSON parsing. Provide a fake adapter in tests.
- **Badge adapter**: embed SVG templates via `//go:embed`, parse once, render by passing `domain.Badge` via `text/template`.
- **HTTP adapter**: minimal `net/http` handlers; no frameworks needed. Converts errors into SVG error badges or plain text (tbd).

## Implementation Steps
1. **Scaffold** the layout above and wire `cmd/server/main.go` to construct adapters + service.
2. **Build adapters** one by one (cache, Duolingo, badge renderer) keeping each under ~80 LOC.
3. **Implement application service** with cache-first logic and fallback error badges.
4. **Add HTTP endpoints** and ensure they only depend on the `ports` package.
5. **Write tests**:
   - Domain: pure Go tests (`domain/streak_test.go`).
   - Application: use fake adapters to test cache hit/miss behavior.
   - Adapters: focus on template rendering (string contains) and Duolingo JSON parsing.

## Testing Strategy
- `go test ./...` should run under a second; rely on fake adapters, not network calls.
- Use table-driven tests for application service (hit vs miss, error badge, variant selection).
- Add one golden file test for the SVG renderer (`testdata/duo.svg.golden`).

## Deployment & Infra (lightweight)
- Keep current Cloud Run + Cloudflare idea, but only after core service works.
- Dockerfile: two stages (build + run) with minimal flags.
- Terraform / GitHub Actions remain optional add-ons once functionality is stable.


This keeps Hexagonal Architecture benefits (clean seams for cache/HTTP) while preserving readable, straightforward code.

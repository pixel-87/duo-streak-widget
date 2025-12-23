duo-streak-widget is a tiny Go service that renders 88×31 SVG badges for Duolingo streaks and GitHub contribution streaks so you can drop them into READMEs and personal sites.

<a href="https://github.com/pixel-87/duo-streak-widget" rel="ed.thomas.dev" target="_blank"><img src="https://api.pixel-87.uk/api/duolingo/button?username=Edward527516" alt="Ed's Duolingo" title="Ed's Duolingo" /></a>

```html
<a href="https://github.com/pixel-87/duo-streak-widget"
     rel="ed.thomas.dev"
     target="_blank">
  <img src="https://api.pixel-87.uk/api/duolingo/button?username=<YOUR_USERNAME>"
           alt="Ed's Duolingo"
           title="Ed's Duolingo"
  />
</a>
```

## Endpoints
- `GET /api/duolingo/button?username=<user>&variant=<optional>` – returns a Duolingo streak badge for the requested user.
- `GET /api/github/button?username=<user>&variant=<optional>` – returns a GitHub contribution streak badge (supports unauthenticated requests but works best with `GITHUB_TOKEN`).

Both endpoints accept an optional `variant` query parameter, defaulting to `default`, and they respond with `image/svg+xml` plus cache-friendly headers.

The Duolingo and GitHub services cache every username for roughly four hours, and the SVG helpers in `api` wrap failures in a tiny error badge so embedders always get a renderable asset.

## Running locally
### Nix / NixOS
- `nix develop`
- `go run ./src/cmd/api`

### Standard Go workflow
- Install Go 1.25.2
- `go run ./src/cmd/api`

### Quick preview
```bash
curl "http://localhost:8080/api/duolingo/button?username=yourname"
curl "http://localhost:8080/api/github/button?username=yourname"
```

## Testing & linting
- Unit tests: `cd src && go test ./...`
- Lint: `golangci-lint run ./...`
- Formatting: `go fmt ./...`
- Vet: `go vet ./...`

These commands are mirrored in `.github/workflows/ci.yml` so CI enforces the same signal.

## Deployment
1. Build the Nix image and push it to GCR via `./deploy.sh` (requires `gcloud auth configure-docker`).
2. Apply the Terraform plan under `terraform/` to create a Cloud Run service that exposes the container publicly.

## Environment variables
- `PORT` – HTTP port (default `8080`).
- `GITHUB_TOKEN` – optional but recommended for authenticated GraphQL queries; it prevents hitting anonymous rate limits.

## License
The project is released under GPLv3+ as declared in `nix/default.nix`.





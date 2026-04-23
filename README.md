# Polybase

Self-hosted user database with OIDC authentication.

## Components

- `polybase/`: CLI frontend
- `polybased/`: Web interface (Go + HTMX)
- `internal/`: Core backend logic with tests

## Usage

To develop or to build polybase, you must have Go 1.24+ and Bun installed.
We are using `just` as a command runner.

Build:
```bash
just build
```

Publish:
```bash
just publish
```

Development:
```bash
just dev            # basic backend
just dev-frontend   # frontend
just dev-rw         # test high packet loss
just dev-hivemind   # if you have hivemind installed (start dev, dev-frontend and dev-air)
just migrate  # initialize database
just clean    # remove artifacts
```

The server now uses OIDC for sign-in. Set the `oidc` section in `polybase.cfg`
with your provider `client_id`, `client_secret`, `issuer_url`, and
`redirect_uri` before starting the backend.

## Icons

Icons are selected from the [Tabler icon set](https://tabler.io/icons), a MIT-licensed icon set.

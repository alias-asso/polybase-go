# Polybase

Self-hosted user database with LDAP authentication.

## Components

- `polybase/`: CLI frontend
- `polybased/`: Web interface (Go + HTMX)
- `internal/`: Core backend logic with tests

## Requirements

- Go 1.21+
- SQLite
- Tailwind CSS
- Templ

## Nix Users

```shell
nix develop
just build
```

## Other Users

Install dependencies:

- `go install github.com/cosmtrek/air@latest`
- `go install github.com/a-h/templ/cmd/templ@v0.2.793`
- `npm install -g tailwindcss`

Build:

```shell
just build
```

Development:

```shell
just dev    # hot reload
just setup  # initialize database
just clean  # remove artifacts
```

## LDAP Development

Start GLAuth development LDAP server:

```shell
glauth -c glauth.cfg
```

Test accounts:

- `paul:paul*`
- `ionys:ionys*`
- `lydia:lydia*`

## Icons

Icons are selected from the [Tabler icon set](https://tabler.io/icons), a MIT-licensed icon set.

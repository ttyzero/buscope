# Agents guide for buscope

Visual inspector for [ttybus](https://github.com/ttyzero/ttybus). Discovers
channels via `LIST`, subscribes, counts traffic, and tails the selected
channel. Companion to gitwing; same `$TTYTHEME`.

## Stack

- Go 1.25, module `github.com/ttyzero/buscope`
- Bubble Tea v2 + Lip Gloss v2
- Theme: `github.com/ttyzero/commons/theme`
- Dial: `github.com/ttyzero/commons/connect`

```sh
make test
make build
```

## Layout

- `cmd/buscope` — flags
- `internal/tap` — per-channel counts, sparkline buckets, message ring
- `internal/ui` — Bubble Tea model + view

## Invariants

- Discover channels from peer `LIST` plus `--watch` / default `files`.
- Always watch `theme` and apply `THEME` / `BORDERS` to this pane too.
- Never wildcard-sub (protocol has no wildcards).
- Honor `$TTYTHEME` then `$CLITHEME` then OSC 11.
- Do not paint a panel background.

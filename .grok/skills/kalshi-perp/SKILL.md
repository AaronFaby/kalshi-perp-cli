---
name: kalshi-perp
description: >
  Work on or with the Kalshi perpetual-futures (margin/perps) CLI in this repo.
  Covers building/running kalshi-perp, RSA-PSS auth, fixed-point prices, demo vs prod,
  REST commands, and WebSocket streaming. Use when editing this CLI, placing or
  debugging perps trades, integrating the margin API, or when the user runs /kalshi-perp.
---

# Kalshi Perp CLI

## Prefer the CLI

For authenticated margin API calls, use the CLI — not ad-hoc curl:

```bash
go run ./cmd/kalshi-perp --format json <command>
# or after build:
./kalshi-perp --format json <command>
```

Drop to raw HTTP only when debugging request signing.

## Defaults

- **Env:** `demo` unless the user explicitly wants production (`--env prod`).
- **Format:** `--format json` when an agent must parse output.
- **Secrets:** never commit API keys or private key PEMs; use env or local config only.

## Auth (checklist)

Every REST request and the WebSocket handshake need:

| Header | Value |
|---|---|
| `KALSHI-ACCESS-KEY` | API key ID |
| `KALSHI-ACCESS-TIMESTAMP` | Unix ms as string |
| `KALSHI-ACCESS-SIGNATURE` | base64(RSA-PSS-SHA256(`timestamp + METHOD + pathWithoutQuery`)) |

Sign path includes `/trade-api/v2/...` and **strips query parameters**. See [references/conventions.md](references/conventions.md).

## Money & size

- Order `price` and `count` are **fixed-point decimal strings**, never JSON numbers.
- Intra-exchange transfer `amount` is **centicents** (integer). Prefer CLI flags that convert dollars when available.

## Extending the CLI

1. Add typed method under `internal/api/`.
2. Wire a cobra command under `internal/cli/`.
3. Update [references/endpoints.md](references/endpoints.md).
4. Add httptest coverage for the new path.
5. Keep FCM and FIX out of scope unless the user asks.

## Safety

- Confirm before production (`--env prod`) order create/cancel/transfer. Mutating commands require `--confirm-prod` against production; `--dry-run` does not.
- Prefer demo for examples and integration smoke tests.

## References

- [endpoints.md](references/endpoints.md) — command ↔ path map
- [conventions.md](references/conventions.md) — auth, URLs, fixed-point, WS channels
- Vendored specs: `docs/openapi/perps_openapi.yaml`, `docs/asyncapi/perps_asyncapi.yaml`
- User docs: `README.md`

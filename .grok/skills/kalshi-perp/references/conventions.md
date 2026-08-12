# Conventions

## Base URLs

| Env | REST | WebSocket |
|---|---|---|
| demo | `https://external-api.demo.kalshi.co/trade-api/v2` | `wss://external-api-margin-ws.demo.kalshi.co/trade-api/ws/v2/margin` |
| prod | `https://external-api.kalshi.com/trade-api/v2` | `wss://external-api-margin-ws.kalshi.com/trade-api/ws/v2/margin` |

CLI: `--env demo|prod` (default `demo`), overrides `--base-url` / `--ws-url`.

Mutating commands against production (`--env prod` or the production REST host) require `--confirm-prod` and print a warning naming the host. `--dry-run` skips that check.

## Config resolution (highest wins)

1. Flags: `--api-key`, `--private-key`, `--env`
2. Env: `KALSHI_API_KEY`, `KALSHI_PRIVATE_KEY_PATH` or `KALSHI_PRIVATE_KEY`, `KALSHI_ENV`
3. File: `~/.config/kalshi-perp/config.yaml` (`$XDG_CONFIG_HOME`)

## RSA-PSS signing

```
message = timestamp_ms + METHOD + path_without_query
signature = base64(RSA_PSS_SHA256(private_key, message))
```

- Salt length = digest length (SHA-256).
- Example path: `/trade-api/v2/margin/orders` (never include `?limit=…`).
- Client builds request URL from base + relative path; sign path is the full URL path.

## Fixed-point

| Type | Example | Notes |
|---|---|---|
| `FixedPointDollars` | `"0.5600"` | Prices / money as strings |
| `FixedPointCount` | `"10.00"` | Contract qty; 0–2 decimals on request |

Never serialize money/qty as `float64` in API JSON.

## Transfer amount

`POST /portfolio/intra_exchange_instance_transfer` uses `amount` in **centicents** (1 dollar = 10_000 centicents). CLI accepts exactly one of `--amount-centicents` or `--amount-dollars`. `source`/`destination` are `event_contract` or `margined`.

## Pagination

- Query: `limit`, `cursor`
- CLI: `--limit` (default 100, always sent), `--cursor`, `--all` (opt-in auto-follow; stops on a repeated cursor or after 100 pages)

## Order create required fields

`ticker`, `client_order_id`, `side` (`bid`|`ask`), `count`, `price`, `time_in_force`, `self_trade_prevention_type`

TIF: `fill_or_kill` | `good_till_canceled` | `immediate_or_cancel`  
STP: `taker_at_cross` | `maker`

## WebSocket

- Auth headers on HTTP upgrade (same three as REST).
- Server pings every ~10s (`heartbeat`); client must pong.
- Timestamps are Unix ms (`*_ms`), not RFC3339.
- Channels: `orderbook_delta`, `ticker`, `trade`, `fill`, `user_orders`, `order_group_updates`.
- `orderbook_delta` requires `--ticker`.

## Output

`--format table|json|jsonl` — agents should use `json` or `jsonl`.

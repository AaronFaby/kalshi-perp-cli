# kalshi-perp

Feature-complete CLI for the [Kalshi perpetual futures (margin) API](https://docs.kalshi.com/margin).

- **REST**: markets, orders, positions, fills, balance, risk, funding, fees, order groups, subaccounts, transfers
- **WebSocket**: `orderbook_delta`, `ticker`, `trade`, `fill`, `user_orders`, `order_group_updates`
- **Auth**: RSA-PSS request signing (same model as event contracts)
- **Ship**: single static Go binary

> Default environment is **demo**. Production is rolling out member-by-member.

## Install

```bash
# From source
go install github.com/aaronfaby/kalshi-perp-cli/cmd/kalshi-perp@latest

# Or build locally
git clone https://github.com/aaronfaby/kalshi-perp-cli
cd kalshi-perp-cli
make build
./kalshi-perp version
```

## Quick start (demo)

1. Create an API key at [demo.kalshi.co](https://demo.kalshi.co/) → Account → API Keys.
2. Save the private key PEM and note the key ID.

```bash
export KALSHI_ENV=demo
export KALSHI_API_KEY='your-key-id'
export KALSHI_PRIVATE_KEY_PATH='/path/to/key.pem'

# Or write a config file
kalshi-perp config init
# edit ~/.config/kalshi-perp/config.yaml

kalshi-perp auth whoami
kalshi-perp markets list --format json
kalshi-perp balance
```

### Place an order

Prices and sizes are **fixed-point decimal strings**, not cents.

```bash
kalshi-perp orders create \
  --ticker YOUR_TICKER \
  --side bid \
  --count 1.00 \
  --price 0.5000 \
  --tif good_till_canceled \
  --stp taker_at_cross \
  --format json

kalshi-perp orders list --status resting
kalshi-perp orders cancel <order_id>
```

Dry-run prints the JSON body without sending:

```bash
kalshi-perp orders create --ticker X --side bid --count 1 --price 0.5 \
  --tif good_till_canceled --stp maker --dry-run --format json
```

### Stream market data

```bash
kalshi-perp stream \
  --channels ticker,trade,orderbook_delta \
  --ticker YOUR_TICKER \
  --reconnect
```

## Configuration

Resolution order (highest wins):

1. Flags (`--api-key`, `--private-key`, `--env`, …)
2. Environment variables
3. Config file (`~/.config/kalshi-perp/config.yaml` or `$XDG_CONFIG_HOME/kalshi-perp/config.yaml`)

| Variable | Description |
|---|---|
| `KALSHI_ENV` | `demo` or `prod` |
| `KALSHI_API_KEY` | API key ID |
| `KALSHI_PRIVATE_KEY_PATH` | Path to PEM private key |
| `KALSHI_PRIVATE_KEY` | PEM contents (alternative to path) |
| `KALSHI_BASE_URL` | Override REST base URL |
| `KALSHI_WS_URL` | Override WebSocket URL |

### Endpoints

| Env | REST | WebSocket |
|---|---|---|
| demo | `https://external-api.demo.kalshi.co/trade-api/v2` | `wss://external-api-margin-ws.demo.kalshi.co/trade-api/ws/v2/margin` |
| prod | `https://external-api.kalshi.com/trade-api/v2` | `wss://external-api-margin-ws.kalshi.com/trade-api/ws/v2/margin` |

## Command overview

| Group | Examples |
|---|---|
| `auth` | `whoami` |
| `account` | `limits` |
| `exchange` | `status`, `enabled` |
| `markets` | `list`, `get`, `orderbook`, `candles`, `trades` |
| `orders` | `list`, `get`, `create`, `cancel`, `amend`, `decrease` |
| `positions` | `list` |
| `fills` | `list` |
| `balance` | (root) |
| `risk` | `get`, `parameters`, `notional-limit` |
| `fees` | `tiers` |
| `funding` | `estimate`, `rates`, `history` |
| `transfer` | `exchange` (event ↔ margin; amount in centicents) |
| `subaccounts` | `create`, `transfer` |
| `groups` | `list`, `create`, `get`, `delete`, `reset`, `trigger`, `limit` |
| `stream` | WebSocket JSONL |
| `config` | `init`, `path`, `show` |

Global flags: `--env`, `--api-key`, `--private-key`, `--format table|json|jsonl`, `--timeout`, `--verbose`.

### Transfers

`transfer exchange` uses **centicents** (`1 USD = 10_000`). Prefer:

```bash
kalshi-perp transfer exchange \
  --source <instance> --destination <instance> \
  --amount-dollars 10.00
```

This endpoint may be disabled until production rollout. Check `kalshi-perp exchange enabled` first.

## Development

```bash
make test
make build
make vendor-specs   # refresh OpenAPI/AsyncAPI snapshots under docs/
```

Vendored specs:

- `docs/openapi/perps_openapi.yaml`
- `docs/asyncapi/perps_asyncapi.yaml`

Agent skill for this repo: `.grok/skills/kalshi-perp/` (`/kalshi-perp`).

## Out of scope (v1)

- FIX gateway
- FCM-only admin endpoints (`/margin/fcm/*`)
- Event-contract non-margin market APIs (except the shared transfer path)

## License

MIT

# Command ↔ API map

Non-FCM surface only. Paths are relative to `/trade-api/v2`.

## Account & exchange

| Command | Method | Path |
|---|---|---|
| `account limits` | GET | `/account/limits/perps` |
| `exchange status` | GET | `/margin/exchange/status` |
| `exchange enabled` | GET | `/margin/enabled` |
| `auth whoami` | GET | `/margin/enabled` + `/account/limits/perps` |

## Markets

| Command | Method | Path |
|---|---|---|
| `markets list` | GET | `/margin/markets` |
| `markets get <ticker>` | GET | `/margin/markets/{ticker}` |
| `markets orderbook <ticker>` | GET | `/margin/markets/{ticker}/orderbook` |
| `markets candles <ticker>` | GET | `/margin/markets/{ticker}/candlesticks` |
| `markets trades <ticker>` | GET | `/margin/trades` |

## Orders

| Command | Method | Path |
|---|---|---|
| `orders list` | GET | `/margin/orders` |
| `orders get <order_id>` | GET | `/margin/orders/{order_id}` |
| `orders create` | POST | `/margin/orders` |
| `orders cancel <order_id>` | DELETE | `/margin/orders/{order_id}` |
| `orders amend <order_id>` | POST | `/margin/orders/{order_id}/amend` |
| `orders decrease <order_id>` | POST | `/margin/orders/{order_id}/decrease` |

## Portfolio

| Command | Method | Path |
|---|---|---|
| `balance` | GET | `/margin/balance` |
| `positions list` | GET | `/margin/positions` |
| `fills list` | GET | `/margin/fills` |
| `transfer exchange` | POST | `/portfolio/intra_exchange_instance_transfer` |
| `subaccounts create` | POST | `/portfolio/margin/subaccounts` |
| `subaccounts transfer` | POST | `/portfolio/margin/subaccounts/transfer` |

## Risk, fees, funding

| Command | Method | Path |
|---|---|---|
| `risk get` | GET | `/margin/risk` |
| `risk parameters` | GET | `/margin/risk_parameters` |
| `risk notional-limit` | GET | `/margin/notional_risk_limit` |
| `fees tiers` | GET | `/margin/fee_tiers` |
| `funding estimate` | GET | `/margin/funding_rates/estimate` |
| `funding rates` | GET | `/margin/funding_rates/historical` |
| `funding history` | GET | `/margin/funding_history` |

## Order groups

| Command | Method | Path |
|---|---|---|
| `groups list` | GET | `/margin/order_groups` |
| `groups create` | POST | `/margin/order_groups/create` |
| `groups get <id>` | GET | `/margin/order_groups/{id}` |
| `groups delete <id>` | DELETE | `/margin/order_groups/{id}` |
| `groups reset <id>` | PUT | `/margin/order_groups/{id}/reset` |
| `groups trigger <id>` | PUT | `/margin/order_groups/{id}/trigger` |
| `groups limit <id>` | PUT | `/margin/order_groups/{id}/limit` |

## WebSocket

| Command | Channels |
|---|---|
| `stream` | `orderbook_delta`, `ticker`, `trade`, `fill`, `user_orders`, `order_group_updates` |

WS URL path: `/trade-api/ws/v2/margin`

## Out of scope

- FCM: `/margin/fcm/*`
- FIX gateway

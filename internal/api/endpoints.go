package api

import (
	"context"
	"fmt"
	"net/url"
)

// --- Account ---

func (c *Client) GetPerpsAccountAPILimits(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.Get(ctx, "/account/limits/perps", nil, &out)
	return out, err
}

// --- Exchange ---

func (c *Client) GetMarginExchangeStatus(ctx context.Context) (*ExchangeStatus, error) {
	var out ExchangeStatus
	err := c.Get(ctx, "/margin/exchange/status", nil, &out)
	return &out, err
}

func (c *Client) GetMarginEnabled(ctx context.Context) (*GetMarginEnabledResponse, error) {
	var out GetMarginEnabledResponse
	err := c.Get(ctx, "/margin/enabled", nil, &out)
	return &out, err
}

// --- Markets ---

func (c *Client) GetMarginMarkets(ctx context.Context, status string) (*GetMarginMarketsResponse, error) {
	q := Q()
	SetStr(q, "status", status)
	var out GetMarginMarketsResponse
	err := c.Get(ctx, "/margin/markets", q, &out)
	return &out, err
}

func (c *Client) GetMarginMarket(ctx context.Context, ticker string) (*GetMarginMarketResponse, error) {
	var out GetMarginMarketResponse
	err := c.Get(ctx, "/margin/markets/"+url.PathEscape(ticker), nil, &out)
	return &out, err
}

func (c *Client) GetMarginMarketOrderbook(ctx context.Context, ticker string) (*GetMarginMarketOrderbookResponse, error) {
	var out GetMarginMarketOrderbookResponse
	err := c.Get(ctx, "/margin/markets/"+url.PathEscape(ticker)+"/orderbook", nil, &out)
	return &out, err
}

type CandlesParams struct {
	StartTs                   int64
	EndTs                     int64
	PeriodInterval            int // 1, 60, 1440
	IncludeLatestBeforeStart  bool
}

func (c *Client) GetMarginMarketCandlesticks(ctx context.Context, ticker string, p CandlesParams) (*GetMarginMarketCandlesticksResponse, error) {
	q := Q()
	q.Set("start_ts", fmt.Sprintf("%d", p.StartTs))
	q.Set("end_ts", fmt.Sprintf("%d", p.EndTs))
	q.Set("period_interval", fmt.Sprintf("%d", p.PeriodInterval))
	if p.IncludeLatestBeforeStart {
		q.Set("include_latest_before_start", "true")
	}
	var out GetMarginMarketCandlesticksResponse
	err := c.Get(ctx, "/margin/markets/"+url.PathEscape(ticker)+"/candlesticks", q, &out)
	return &out, err
}

type TradesParams struct {
	Ticker string
	Limit  *int
	Cursor string
	MinTs  *int64
	MaxTs  *int64
}

func (c *Client) GetMarginTrades(ctx context.Context, p TradesParams) (*GetMarginTradesResponse, error) {
	q := Q()
	q.Set("ticker", p.Ticker)
	SetInt(q, "limit", p.Limit)
	SetStr(q, "cursor", p.Cursor)
	SetInt64(q, "min_ts", p.MinTs)
	SetInt64(q, "max_ts", p.MaxTs)
	var out GetMarginTradesResponse
	err := c.Get(ctx, "/margin/trades", q, &out)
	return &out, err
}

// --- Orders ---

type OrdersParams struct {
	Ticker     string
	Status     string
	Limit      *int
	Cursor     string
	MinTs      *int64
	MaxTs      *int64
	Subaccount *int
}

func (c *Client) GetMarginOrders(ctx context.Context, p OrdersParams) (*GetMarginOrdersResponse, error) {
	q := Q()
	SetStr(q, "ticker", p.Ticker)
	SetStr(q, "status", p.Status)
	SetInt(q, "limit", p.Limit)
	SetStr(q, "cursor", p.Cursor)
	SetInt64(q, "min_ts", p.MinTs)
	SetInt64(q, "max_ts", p.MaxTs)
	SetInt(q, "subaccount", p.Subaccount)
	var out GetMarginOrdersResponse
	err := c.Get(ctx, "/margin/orders", q, &out)
	return &out, err
}

func (c *Client) GetMarginOrder(ctx context.Context, orderID string) (*GetMarginOrderResponse, error) {
	var out GetMarginOrderResponse
	err := c.Get(ctx, "/margin/orders/"+url.PathEscape(orderID), nil, &out)
	return &out, err
}

func (c *Client) CreateMarginOrder(ctx context.Context, req CreateMarginOrderRequest) (*CreateMarginOrderResponse, error) {
	var out CreateMarginOrderResponse
	err := c.Post(ctx, "/margin/orders", req, &out)
	return &out, err
}

func (c *Client) CancelMarginOrder(ctx context.Context, orderID string) (map[string]any, error) {
	var out map[string]any
	err := c.Delete(ctx, "/margin/orders/"+url.PathEscape(orderID), &out)
	return out, err
}

func (c *Client) AmendMarginOrder(ctx context.Context, orderID string, req AmendMarginOrderRequest) (map[string]any, error) {
	var out map[string]any
	err := c.Post(ctx, "/margin/orders/"+url.PathEscape(orderID)+"/amend", req, &out)
	return out, err
}

func (c *Client) DecreaseMarginOrder(ctx context.Context, orderID string, req DecreaseMarginOrderRequest) (map[string]any, error) {
	var out map[string]any
	err := c.Post(ctx, "/margin/orders/"+url.PathEscape(orderID)+"/decrease", req, &out)
	return out, err
}

// --- Portfolio ---

func (c *Client) GetMarginBalance(ctx context.Context) (*GetMarginBalanceResponse, error) {
	var out GetMarginBalanceResponse
	err := c.Get(ctx, "/margin/balance", nil, &out)
	return &out, err
}

type PositionsParams struct {
	Subaccount *int
	Ticker     string
}

func (c *Client) GetMarginPositions(ctx context.Context, p PositionsParams) (*GetMarginPositionsResponse, error) {
	q := Q()
	SetInt(q, "subaccount", p.Subaccount)
	SetStr(q, "ticker", p.Ticker)
	var out GetMarginPositionsResponse
	err := c.Get(ctx, "/margin/positions", q, &out)
	return &out, err
}

type FillsParams struct {
	Subaccount *int
	Limit      *int
	Cursor     string
	MinTs      *int64
	MaxTs      *int64
}

func (c *Client) GetMarginFills(ctx context.Context, p FillsParams) (*GetMarginFillsResponse, error) {
	q := Q()
	SetInt(q, "subaccount", p.Subaccount)
	SetInt(q, "limit", p.Limit)
	SetStr(q, "cursor", p.Cursor)
	SetInt64(q, "min_ts", p.MinTs)
	SetInt64(q, "max_ts", p.MaxTs)
	var out GetMarginFillsResponse
	err := c.Get(ctx, "/margin/fills", q, &out)
	return &out, err
}

func (c *Client) IntraExchangeInstanceTransfer(ctx context.Context, req IntraExchangeInstanceTransferRequest) (map[string]any, error) {
	var out map[string]any
	err := c.Post(ctx, "/portfolio/intra_exchange_instance_transfer", req, &out)
	return out, err
}

func (c *Client) CreateMarginSubaccount(ctx context.Context) (*CreateSubaccountResponse, error) {
	var out CreateSubaccountResponse
	err := c.Post(ctx, "/portfolio/margin/subaccounts", map[string]any{}, &out)
	return &out, err
}

func (c *Client) TransferBetweenSubaccounts(ctx context.Context, req ApplySubaccountTransferRequest) (map[string]any, error) {
	var out map[string]any
	err := c.Post(ctx, "/portfolio/margin/subaccounts/transfer", req, &out)
	return out, err
}

// --- Risk / fees / funding ---

func (c *Client) GetMarginRisk(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.Get(ctx, "/margin/risk", nil, &out)
	return out, err
}

func (c *Client) GetMarginRiskParameters(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.Get(ctx, "/margin/risk_parameters", nil, &out)
	return out, err
}

func (c *Client) GetMarginNotionalRiskLimit(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.Get(ctx, "/margin/notional_risk_limit", nil, &out)
	return out, err
}

func (c *Client) GetMarginFeeTiers(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.Get(ctx, "/margin/fee_tiers", nil, &out)
	return out, err
}

func (c *Client) GetMarginFundingRateEstimate(ctx context.Context, ticker string) (map[string]any, error) {
	q := Q()
	q.Set("ticker", ticker)
	var out map[string]any
	err := c.Get(ctx, "/margin/funding_rates/estimate", q, &out)
	return out, err
}

type FundingRatesParams struct {
	Ticker  string
	StartTs *int64
	EndTs   *int64
}

func (c *Client) GetMarginHistoricalFundingRates(ctx context.Context, p FundingRatesParams) (map[string]any, error) {
	q := Q()
	SetStr(q, "ticker", p.Ticker)
	SetInt64(q, "start_ts", p.StartTs)
	SetInt64(q, "end_ts", p.EndTs)
	var out map[string]any
	err := c.Get(ctx, "/margin/funding_rates/historical", q, &out)
	return out, err
}

type FundingHistoryParams struct {
	Ticker     string
	StartDate  string // YYYY-MM-DD
	EndDate    string
	Subaccount *int
}

func (c *Client) GetMarginFundingHistory(ctx context.Context, p FundingHistoryParams) (map[string]any, error) {
	q := Q()
	SetStr(q, "ticker", p.Ticker)
	q.Set("start_date", p.StartDate)
	q.Set("end_date", p.EndDate)
	SetInt(q, "subaccount", p.Subaccount)
	var out map[string]any
	err := c.Get(ctx, "/margin/funding_history", q, &out)
	return out, err
}

// --- Order groups ---

func (c *Client) GetMarginOrderGroups(ctx context.Context) (*GetOrderGroupsResponse, error) {
	var out GetOrderGroupsResponse
	err := c.Get(ctx, "/margin/order_groups", nil, &out)
	return &out, err
}

func (c *Client) CreateMarginOrderGroup(ctx context.Context, req CreateOrderGroupRequest) (*CreateOrderGroupResponse, error) {
	var out CreateOrderGroupResponse
	err := c.Post(ctx, "/margin/order_groups/create", req, &out)
	return &out, err
}

func (c *Client) GetMarginOrderGroup(ctx context.Context, id string) (*GetOrderGroupResponse, error) {
	var out GetOrderGroupResponse
	err := c.Get(ctx, "/margin/order_groups/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (c *Client) DeleteMarginOrderGroup(ctx context.Context, id string) error {
	return c.Delete(ctx, "/margin/order_groups/"+url.PathEscape(id), nil)
}

func (c *Client) ResetMarginOrderGroup(ctx context.Context, id string) error {
	return c.Put(ctx, "/margin/order_groups/"+url.PathEscape(id)+"/reset", map[string]any{}, nil)
}

func (c *Client) TriggerMarginOrderGroup(ctx context.Context, id string) error {
	return c.Put(ctx, "/margin/order_groups/"+url.PathEscape(id)+"/trigger", map[string]any{}, nil)
}

func (c *Client) UpdateMarginOrderGroupLimit(ctx context.Context, id string, req UpdateOrderGroupLimitRequest) error {
	return c.Put(ctx, "/margin/order_groups/"+url.PathEscape(id)+"/limit", req, nil)
}

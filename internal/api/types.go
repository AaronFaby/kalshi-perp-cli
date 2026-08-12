package api

// ErrorResponse is the standard Kalshi error body.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e ErrorResponse) Error() string {
	if e.Details != "" {
		return e.Code + ": " + e.Message + " (" + e.Details + ")"
	}
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	return e.Message
}

// Fixed-point strings — keep as string, never float64.
type FixedPointCount = string
type FixedPointDollars = string

// --- Exchange ---

type ExchangeStatus struct {
	ExchangeActive              bool    `json:"exchange_active"`
	TradingActive               bool    `json:"trading_active"`
	ExchangeEstimatedResumeTime *string `json:"exchange_estimated_resume_time,omitempty"`
}

type GetMarginEnabledResponse struct {
	Enabled bool `json:"enabled"`
}

// --- Markets ---

type TickerPrice struct {
	Price FixedPointDollars `json:"price"`
	TsMs  int64             `json:"ts_ms"`
}

type MarginMarket struct {
	Ticker                           string             `json:"ticker"`
	Title                            string             `json:"title"`
	Status                           string             `json:"status"`
	ContractSize                     string             `json:"contract_size"`
	TickSize                         FixedPointDollars  `json:"tick_size"`
	FractionalTradingEnabled         bool               `json:"fractional_trading_enabled"`
	ExchangeIndex                    int                `json:"exchange_index"`
	Schedule                         any                `json:"schedule,omitempty"`
	LeverageEstimate                 *float64           `json:"leverage_estimate,omitempty"`
	LeverageEstimates                map[string]float64 `json:"leverage_estimates,omitempty"`
	LongLeverageEstimates            map[string]float64 `json:"long_leverage_estimates,omitempty"`
	ShortLeverageEstimates           map[string]float64 `json:"short_leverage_estimates,omitempty"`
	Price                            FixedPointDollars  `json:"price,omitempty"`
	Volume                           FixedPointCount    `json:"volume,omitempty"`
	VolumeNotionalValueDollars       FixedPointDollars  `json:"volume_notional_value_dollars,omitempty"`
	OpenInterest                     FixedPointCount    `json:"open_interest,omitempty"`
	OpenInterestNotionalValueDollars FixedPointDollars  `json:"open_interest_notional_value_dollars,omitempty"`
	Volume24h                        FixedPointCount    `json:"volume_24h,omitempty"`
	Volume24hNotionalValueDollars    FixedPointDollars  `json:"volume_24h_notional_value_dollars,omitempty"`
	Bid                              FixedPointDollars  `json:"bid,omitempty"`
	Ask                              FixedPointDollars  `json:"ask,omitempty"`
	SettlementMarkPrice              *TickerPrice       `json:"settlement_mark_price,omitempty"`
	LiquidationMarkPrice             *TickerPrice       `json:"liquidation_mark_price,omitempty"`
	ReferencePrice                   *TickerPrice       `json:"reference_price,omitempty"`
}

type GetMarginMarketsResponse struct {
	Markets []MarginMarket `json:"markets"`
	Cursor  string         `json:"cursor,omitempty"`
}

type GetMarginMarketResponse struct {
	Market MarginMarket `json:"market"`
}

// Orderbook: price levels as [price, count] string pairs.
type PriceLevel = []string

type GetMarginMarketOrderbookResponse struct {
	Orderbook struct {
		Bids []PriceLevel `json:"bids"`
		Asks []PriceLevel `json:"asks"`
	} `json:"orderbook"`
}

type BidAskDistributionHistorical struct {
	Open  FixedPointDollars `json:"open"`
	Low   FixedPointDollars `json:"low"`
	High  FixedPointDollars `json:"high"`
	Close FixedPointDollars `json:"close"`
}

// Price fields are nullable when the candle has no trades.
type PriceDistributionHistorical struct {
	Open     *FixedPointDollars `json:"open"`
	Low      *FixedPointDollars `json:"low"`
	High     *FixedPointDollars `json:"high"`
	Close    *FixedPointDollars `json:"close"`
	Mean     *FixedPointDollars `json:"mean"`
	Previous *FixedPointDollars `json:"previous"`
}

type Candlestick struct {
	EndPeriodTs                      int64                        `json:"end_period_ts"`
	Bid                              BidAskDistributionHistorical `json:"bid"`
	Ask                              BidAskDistributionHistorical `json:"ask"`
	Price                            PriceDistributionHistorical  `json:"price"`
	Volume                           FixedPointCount              `json:"volume"`
	VolumeNotionalValueDollars       FixedPointDollars            `json:"volume_notional_value_dollars"`
	OpenInterest                     FixedPointCount              `json:"open_interest"`
	OpenInterestNotionalValueDollars FixedPointDollars            `json:"open_interest_notional_value_dollars"`
}

type GetMarginMarketCandlesticksResponse struct {
	Ticker       string        `json:"ticker,omitempty"`
	Candlesticks []Candlestick `json:"candlesticks"`
}

type MarginTrade struct {
	TradeID     string            `json:"trade_id"`
	Ticker      string            `json:"ticker"`
	Count       FixedPointCount   `json:"count"`
	Price       FixedPointDollars `json:"price"`
	TakerSide   string            `json:"taker_side"`
	Ts          int64             `json:"ts,omitempty"`
	CreatedTime string            `json:"created_time,omitempty"`
}

type GetMarginTradesResponse struct {
	Trades []MarginTrade `json:"trades"`
	Cursor string        `json:"cursor,omitempty"`
}

// --- Orders ---

type CreateMarginOrderRequest struct {
	Ticker                  string            `json:"ticker"`
	ClientOrderID           string            `json:"client_order_id"`
	Side                    string            `json:"side"`
	Count                   FixedPointCount   `json:"count"`
	Price                   FixedPointDollars `json:"price"`
	TimeInForce             string            `json:"time_in_force"`
	SelfTradePreventionType string            `json:"self_trade_prevention_type"`
	ExpirationTime          *int64            `json:"expiration_time,omitempty"`
	PostOnly                *bool             `json:"post_only,omitempty"`
	CancelOrderOnPause      *bool             `json:"cancel_order_on_pause,omitempty"`
	ReduceOnly              *bool             `json:"reduce_only,omitempty"`
	Subaccount              *int              `json:"subaccount,omitempty"`
	OrderGroupID            string            `json:"order_group_id,omitempty"`
}

type CreateMarginOrderResponse struct {
	OrderID          string            `json:"order_id"`
	ClientOrderID    string            `json:"client_order_id,omitempty"`
	FillCount        FixedPointCount   `json:"fill_count"`
	RemainingCount   FixedPointCount   `json:"remaining_count"`
	AverageFillPrice FixedPointDollars `json:"average_fill_price,omitempty"`
	AverageFeePaid   FixedPointDollars `json:"average_fee_paid,omitempty"`
}

type MarginOrder struct {
	OrderID                 string            `json:"order_id"`
	UserID                  string            `json:"user_id,omitempty"`
	ClientOrderID           string            `json:"client_order_id"`
	Ticker                  string            `json:"ticker"`
	Side                    string            `json:"side"`
	Price                   FixedPointDollars `json:"price"`
	FillCount               FixedPointCount   `json:"fill_count"`
	RemainingCount          FixedPointCount   `json:"remaining_count"`
	LastUpdateReason        string            `json:"last_update_reason,omitempty"`
	ExpirationTime          *string           `json:"expiration_time,omitempty"`
	CreatedTime             *string           `json:"created_time,omitempty"`
	LastUpdateTime          *string           `json:"last_update_time,omitempty"`
	SelfTradePreventionType *string           `json:"self_trade_prevention_type,omitempty"`
	CancelOrderOnPause      bool              `json:"cancel_order_on_pause,omitempty"`
	OrderGroupID            string            `json:"order_group_id,omitempty"`
	OrderSource             string            `json:"order_source,omitempty"`
	OrderReason             string            `json:"order_reason,omitempty"`
	Subaccount              *int              `json:"subaccount,omitempty"`
	Status                  string            `json:"status,omitempty"`
}

// CancelMarginOrderResponse is returned by DELETE /margin/orders/{order_id}.
type CancelMarginOrderResponse struct {
	OrderID       string          `json:"order_id"`
	ClientOrderID string          `json:"client_order_id,omitempty"`
	ReducedBy     FixedPointCount `json:"reduced_by"`
}

type GetMarginOrdersResponse struct {
	Orders []MarginOrder `json:"orders"`
	Cursor string        `json:"cursor"`
}

type GetMarginOrderResponse struct {
	Order MarginOrder `json:"order"`
}

type AmendMarginOrderRequest struct {
	Ticker               string            `json:"ticker"`
	Side                 string            `json:"side"`
	Price                FixedPointDollars `json:"price"`
	Count                FixedPointCount   `json:"count"`
	ClientOrderID        string            `json:"client_order_id,omitempty"`
	UpdatedClientOrderID string            `json:"updated_client_order_id,omitempty"`
}

type DecreaseMarginOrderRequest struct {
	ReduceBy *FixedPointCount `json:"reduce_by,omitempty"`
	ReduceTo *FixedPointCount `json:"reduce_to,omitempty"`
}

// --- Portfolio ---

type MarginSubaccountBalance struct {
	Subaccount          int               `json:"subaccount"`
	PositionValue       FixedPointDollars `json:"position_value"`
	AccountEquity       FixedPointDollars `json:"account_equity"`
	MaintenanceMargin   FixedPointDollars `json:"maintenance_margin"`
	InitialMargin       FixedPointDollars `json:"initial_margin"`
	RestingOrdersMargin FixedPointDollars `json:"resting_orders_margin"`
	AvailableBalance    FixedPointDollars `json:"available_balance"`
}

type GetMarginBalanceResponse struct {
	SubaccountBalances []MarginSubaccountBalance `json:"subaccount_balances"`
	SettledFunds       FixedPointDollars         `json:"settled_funds"`
}

type MarginPosition struct {
	Subaccount    int                `json:"subaccount"`
	MarketTicker  string             `json:"market_ticker"`
	Position      FixedPointCount    `json:"position"`
	EntryPrice    FixedPointDollars  `json:"entry_price"`
	UnrealizedPnL FixedPointDollars  `json:"unrealized_pnl"`
	MarginUsed    *FixedPointDollars `json:"margin_used,omitempty"`
	Fees          FixedPointDollars  `json:"fees"`
	ROE           *float64           `json:"roe,omitempty"`
	IsPortfolio   bool               `json:"is_portfolio"`
}

type GetMarginPositionsResponse struct {
	Positions []MarginPosition `json:"positions"`
}

type MarginFill struct {
	FillID        string            `json:"fill_id"`
	OrderID       string            `json:"order_id"`
	Ticker        string            `json:"ticker"`
	Side          string            `json:"side"`
	Count         FixedPointCount   `json:"count"`
	Price         FixedPointDollars `json:"price"`
	EntryPrice    FixedPointDollars `json:"entry_price"`
	Fees          FixedPointDollars `json:"fees"`
	RealizedPnL   FixedPointDollars `json:"realized_pnl"`
	OrderSource   string            `json:"order_source,omitempty"`
	IsTaker       bool              `json:"is_taker,omitempty"`
	CreatedTime   string            `json:"created_time,omitempty"`
	Ts            int64             `json:"ts,omitempty"`
	ClientOrderID string            `json:"client_order_id,omitempty"`
	Subaccount    *int              `json:"subaccount,omitempty"`
}

type GetMarginFillsResponse struct {
	Fills  []MarginFill `json:"fills"`
	Cursor string       `json:"cursor,omitempty"`
}

type IntraExchangeInstanceTransferRequest struct {
	Source                   string `json:"source"`
	Destination              string `json:"destination"`
	Amount                   int64  `json:"amount"` // centicents
	SourceExchangeShard      *int   `json:"source_exchange_shard,omitempty"`
	DestinationExchangeShard *int   `json:"destination_exchange_shard,omitempty"`
}

type CreateSubaccountResponse struct {
	SubaccountNumber int `json:"subaccount_number"`
}

type ApplySubaccountTransferRequest struct {
	ClientTransferID string `json:"client_transfer_id"`
	FromSubaccount   int    `json:"from_subaccount"`
	ToSubaccount     int    `json:"to_subaccount"`
	AmountCents      int64  `json:"amount_cents"`
}

// --- Order groups ---

type CreateOrderGroupRequest struct {
	Subaccount       *int             `json:"subaccount,omitempty"`
	ContractsLimit   *int64           `json:"contracts_limit,omitempty"`
	ContractsLimitFp *FixedPointCount `json:"contracts_limit_fp,omitempty"`
	ExchangeIndex    *int             `json:"exchange_index,omitempty"`
}

type CreateOrderGroupResponse struct {
	OrderGroupID  string `json:"order_group_id"`
	Subaccount    int    `json:"subaccount"`
	ExchangeIndex int    `json:"exchange_index"`
}

type OrderGroup struct {
	OrderGroupID        string          `json:"id,omitempty"`
	ContractsLimitFp    FixedPointCount `json:"contracts_limit_fp,omitempty"`
	IsAutoCancelEnabled bool            `json:"is_auto_cancel_enabled,omitempty"`
	ExchangeIndex       int             `json:"exchange_index,omitempty"`
}

type GetOrderGroupsResponse struct {
	OrderGroups []OrderGroup `json:"order_groups"`
}

type GetOrderGroupResponse struct {
	IsAutoCancelEnabled bool            `json:"is_auto_cancel_enabled"`
	ContractsLimitFp    FixedPointCount `json:"contracts_limit_fp,omitempty"`
	Orders              []string        `json:"orders,omitempty"`
	ExchangeIndex       int             `json:"exchange_index,omitempty"`
}

type UpdateOrderGroupLimitRequest struct {
	ContractsLimit   *int64           `json:"contracts_limit,omitempty"`
	ContractsLimitFp *FixedPointCount `json:"contracts_limit_fp,omitempty"`
}

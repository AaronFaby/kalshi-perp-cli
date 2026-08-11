package cli

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/aaronfaby/kalshi-perp-cli/internal/api"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func newBalanceCmd(opts *RootOptions) *cobra.Command {
	var computeAvailableBalance bool
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Get margin balance breakdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginBalance(ctx(), computeAvailableBalance)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(out.SubaccountBalances))
			for _, b := range out.SubaccountBalances {
				rows = append(rows, []string{
					strconv.Itoa(b.Subaccount),
					b.AvailableBalance,
					b.AccountEquity,
					b.PositionValue,
					b.MaintenanceMargin,
					b.InitialMargin,
					b.RestingOrdersMargin,
				})
			}
			return rt.out.PrintTable(
				[]string{"SUB", "AVAILABLE", "EQUITY", "POS_VALUE", "MAINT", "INITIAL", "RESTING"},
				rows, out,
			)
		},
	}
	cmd.Flags().BoolVar(&computeAvailableBalance, "compute-available-balance", false, "Compute available balance (higher API cost)")
	return cmd
}

func newPositionsCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "positions", Short: "Open positions"}
	var ticker string
	var subaccount int
	list := &cobra.Command{
		Use:   "list",
		Short: "List margin positions",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			p := api.PositionsParams{Ticker: ticker}
			if cmd.Flags().Changed("subaccount") {
				p.Subaccount = &subaccount
			}
			out, err := rt.client.GetMarginPositions(ctx(), p)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(out.Positions))
			for _, pos := range out.Positions {
				mu := ""
				if pos.MarginUsed != nil {
					mu = *pos.MarginUsed
				}
				rows = append(rows, []string{
					strconv.Itoa(pos.Subaccount),
					pos.MarketTicker,
					pos.Position,
					pos.EntryPrice,
					pos.UnrealizedPnL,
					mu,
					pos.Fees,
				})
			}
			return rt.out.PrintTable(
				[]string{"SUB", "TICKER", "POSITION", "ENTRY", "U_PNL", "MARGIN", "FEES"},
				rows, out,
			)
		},
	}
	list.Flags().StringVar(&ticker, "ticker", "", "Filter by ticker")
	list.Flags().IntVar(&subaccount, "subaccount", 0, "Subaccount")
	cmd.AddCommand(list)
	return cmd
}

func newFillsCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "fills", Short: "User fills"}
	var limit, subaccount int
	var cursor string
	var minTs, maxTs int64
	var all bool
	list := &cobra.Command{
		Use:   "list",
		Short: "List fills",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			p := api.FillsParams{Cursor: cursor}
			if cmd.Flags().Changed("limit") {
				p.Limit = &limit
			}
			if cmd.Flags().Changed("subaccount") {
				p.Subaccount = &subaccount
			}
			if cmd.Flags().Changed("min-ts") {
				p.MinTs = &minTs
			}
			if cmd.Flags().Changed("max-ts") {
				p.MaxTs = &maxTs
			}
			var allFills []api.MarginFill
			for {
				out, err := rt.client.GetMarginFills(ctx(), p)
				if err != nil {
					return err
				}
				allFills = append(allFills, out.Fills...)
				if !all || out.Cursor == "" {
					rows := make([][]string, 0, len(allFills))
					for _, f := range allFills {
						rows = append(rows, []string{
							f.FillID, f.OrderID, f.Ticker, f.Side, f.Price, f.Count,
							f.EntryPrice, f.Fees, f.RealizedPnL,
						})
					}
					payload := map[string]any{"fills": allFills, "cursor": out.Cursor}
					return rt.out.PrintTable(
						[]string{"FILL_ID", "ORDER_ID", "TICKER", "SIDE", "PRICE", "COUNT", "ENTRY", "FEES", "R_PNL"},
						rows, payload,
					)
				}
				p.Cursor = out.Cursor
			}
		},
	}
	list.Flags().IntVar(&limit, "limit", 100, "Page size")
	list.Flags().StringVar(&cursor, "cursor", "", "Cursor")
	list.Flags().IntVar(&subaccount, "subaccount", 0, "Subaccount")
	list.Flags().Int64Var(&minTs, "min-ts", 0, "Min ts")
	list.Flags().Int64Var(&maxTs, "max-ts", 0, "Max ts")
	list.Flags().BoolVar(&all, "all", false, "Follow cursors")
	cmd.AddCommand(list)
	return cmd
}

func newRiskCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "risk", Short: "Risk metrics and parameters"}
	cmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Account leverage and liquidation data",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginRisk(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "parameters",
		Short: "System risk parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginRiskParameters(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "notional-limit",
		Short: "Notional risk limit",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginNotionalRiskLimit(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})
	return cmd
}

func newFeesCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "fees", Short: "Fee schedule"}
	cmd.AddCommand(&cobra.Command{
		Use:   "tiers",
		Short: "Margin fee tiers by market",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginFeeTiers(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})
	return cmd
}

func newFundingCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "funding", Short: "Funding rates and payments"}

	var estTicker string
	estimate := &cobra.Command{
		Use:   "estimate",
		Short: "Estimated funding rate for current period",
		RunE: func(cmd *cobra.Command, args []string) error {
			if estTicker == "" {
				return fmt.Errorf("--ticker is required")
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginFundingRateEstimate(ctx(), estTicker)
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	estimate.Flags().StringVar(&estTicker, "ticker", "", "Market ticker (required)")
	cmd.AddCommand(estimate)

	var ratesTicker string
	var startTs, endTs int64
	rates := &cobra.Command{
		Use:   "rates",
		Short: "Historical funding rates",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			p := api.FundingRatesParams{Ticker: ratesTicker}
			if cmd.Flags().Changed("start-ts") {
				p.StartTs = &startTs
			}
			if cmd.Flags().Changed("end-ts") {
				p.EndTs = &endTs
			}
			out, err := rt.client.GetMarginHistoricalFundingRates(ctx(), p)
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	rates.Flags().StringVar(&ratesTicker, "ticker", "", "Market ticker (empty = all)")
	rates.Flags().Int64Var(&startTs, "start-ts", 0, "Start unix ts")
	rates.Flags().Int64Var(&endTs, "end-ts", 0, "End unix ts")
	cmd.AddCommand(rates)

	var histTicker, startDate, endDate string
	var histSub int
	history := &cobra.Command{
		Use:   "history",
		Short: "Your funding payment history",
		RunE: func(cmd *cobra.Command, args []string) error {
			if startDate == "" || endDate == "" {
				return fmt.Errorf("--start-date and --end-date are required (YYYY-MM-DD)")
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			p := api.FundingHistoryParams{
				Ticker:    histTicker,
				StartDate: startDate,
				EndDate:   endDate,
			}
			if cmd.Flags().Changed("subaccount") {
				p.Subaccount = &histSub
			}
			out, err := rt.client.GetMarginFundingHistory(ctx(), p)
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	history.Flags().StringVar(&histTicker, "ticker", "", "Market ticker (empty = all)")
	history.Flags().StringVar(&startDate, "start-date", "", "Inclusive UTC start YYYY-MM-DD (required)")
	history.Flags().StringVar(&endDate, "end-date", "", "Inclusive UTC end YYYY-MM-DD (required)")
	history.Flags().IntVar(&histSub, "subaccount", 0, "Subaccount")
	cmd.AddCommand(history)

	return cmd
}

func newTransferCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "transfer", Short: "Fund transfers"}

	var source, dest string
	var amountCenticents int64
	var amountDollars string
	var srcShard, dstShard int
	exchange := &cobra.Command{
		Use:   "exchange",
		Short: "Transfer between event-contract and margin balances (amount in centicents or --amount-dollars)",
		Long:  "POST /portfolio/intra_exchange_instance_transfer. Amount is in centicents (1 USD = 10000). May be unavailable until production rollout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if source == "" || dest == "" {
				return fmt.Errorf("--source and --destination are required (exchange instance names)")
			}
			amount := amountCenticents
			if amountDollars != "" {
				c, err := dollarsToCenticents(amountDollars)
				if err != nil {
					return err
				}
				amount = c
			}
			if amount <= 0 {
				return fmt.Errorf("provide --amount-centicents or --amount-dollars")
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			req := api.IntraExchangeInstanceTransferRequest{
				Source:      source,
				Destination: dest,
				Amount:      amount,
			}
			if cmd.Flags().Changed("source-shard") {
				req.SourceExchangeShard = &srcShard
			}
			if cmd.Flags().Changed("destination-shard") {
				req.DestinationExchangeShard = &dstShard
			}
			out, err := rt.client.IntraExchangeInstanceTransfer(ctx(), req)
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	exchange.Flags().StringVar(&source, "source", "", "Source exchange instance (required)")
	exchange.Flags().StringVar(&dest, "destination", "", "Destination exchange instance (required)")
	exchange.Flags().Int64Var(&amountCenticents, "amount-centicents", 0, "Amount in centicents (1 USD = 10000)")
	exchange.Flags().StringVar(&amountDollars, "amount-dollars", "", "Amount in dollars (converted to centicents)")
	exchange.Flags().IntVar(&srcShard, "source-shard", 0, "Source shard")
	exchange.Flags().IntVar(&dstShard, "destination-shard", 0, "Destination shard")
	cmd.AddCommand(exchange)
	return cmd
}

func newSubaccountsCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "subaccounts", Short: "Margin subaccounts"}

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a new margin subaccount",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.CreateMarginSubaccount(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})

	var from, to int
	var amountCents int64
	var amountDollars, clientTransferID string
	transfer := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer between margin subaccounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("amount-cents") == (amountDollars != "") {
				return fmt.Errorf("provide exactly one of --amount-cents or --amount-dollars")
			}
			amount := amountCents
			if amountDollars != "" {
				c, err := dollarsToUnits(amountDollars, 100)
				if err != nil {
					return err
				}
				amount = c
			}
			if amount <= 0 {
				return fmt.Errorf("amount must be positive")
			}
			if clientTransferID == "" {
				clientTransferID = uuid.NewString()
			} else if _, err := uuid.Parse(clientTransferID); err != nil {
				return fmt.Errorf("invalid --client-transfer-id: %w", err)
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.TransferBetweenSubaccounts(ctx(), api.ApplySubaccountTransferRequest{
				ClientTransferID: clientTransferID,
				FromSubaccount:   from,
				ToSubaccount:     to,
				AmountCents:      amount,
			})
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	transfer.Flags().IntVar(&from, "from", 0, "From subaccount (0=primary)")
	transfer.Flags().IntVar(&to, "to", 0, "To subaccount")
	transfer.Flags().Int64Var(&amountCents, "amount-cents", 0, "Amount in cents")
	transfer.Flags().StringVar(&amountDollars, "amount-dollars", "", "Amount in dollars (converted to cents)")
	transfer.Flags().StringVar(&clientTransferID, "client-transfer-id", "", "Idempotency UUID (auto-generated if omitted)")
	cmd.AddCommand(transfer)

	return cmd
}

// dollarsToCenticents converts a dollar string to centicents (1 USD = 10_000).
func dollarsToCenticents(s string) (int64, error) {
	return dollarsToUnits(s, 10_000)
}

func dollarsToUnits(s string, unitsPerDollar int64) (int64, error) {
	s = strings.TrimSpace(s)
	amount, ok := new(big.Rat).SetString(s)
	if !ok {
		return 0, fmt.Errorf("invalid dollars %q", s)
	}
	if amount.Sign() < 0 {
		return 0, fmt.Errorf("amount must be non-negative")
	}
	amount.Mul(amount, big.NewRat(unitsPerDollar, 1))
	if !amount.IsInt() {
		return 0, fmt.Errorf("amount has too many decimal places")
	}
	if !amount.Num().IsInt64() {
		return 0, fmt.Errorf("amount is too large")
	}
	return amount.Num().Int64(), nil
}

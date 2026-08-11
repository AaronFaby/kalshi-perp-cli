package cli

import (
	"fmt"
	"strconv"

	"github.com/aaronfaby/kalshi-perp-cli/internal/api"
	"github.com/spf13/cobra"
)

func newMarketsCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "markets", Short: "Margin market data"}

	var status string
	list := &cobra.Command{
		Use:   "list",
		Short: "List margin markets",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginMarkets(ctx(), status)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(out.Markets))
			for _, m := range out.Markets {
				rows = append(rows, []string{m.Ticker, m.Status, m.Title, m.TickSize, strconv.Itoa(m.ExchangeIndex)})
			}
			return rt.out.PrintTable([]string{"TICKER", "STATUS", "TITLE", "TICK", "EX_IDX"}, rows, out)
		},
	}
	list.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.AddCommand(list)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <ticker>",
		Short: "Get a margin market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginMarket(ctx(), args[0])
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "orderbook <ticker>",
		Short: "Get market orderbook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginMarketOrderbook(ctx(), args[0])
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})

	var startTs, endTs int64
	var period int
	var includeLatest bool
	candles := &cobra.Command{
		Use:   "candles <ticker>",
		Short: "Get candlesticks (period 1|60|1440 minutes)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if startTs == 0 || endTs == 0 {
				return fmt.Errorf("--start-ts and --end-ts are required")
			}
			if period != 1 && period != 60 && period != 1440 {
				return fmt.Errorf("--period must be 1, 60, or 1440")
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginMarketCandlesticks(ctx(), args[0], api.CandlesParams{
				StartTs: startTs, EndTs: endTs, PeriodInterval: period, IncludeLatestBeforeStart: includeLatest,
			})
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	candles.Flags().Int64Var(&startTs, "start-ts", 0, "Start unix timestamp (required)")
	candles.Flags().Int64Var(&endTs, "end-ts", 0, "End unix timestamp (required)")
	candles.Flags().IntVar(&period, "period", 60, "Period minutes: 1, 60, 1440")
	candles.Flags().BoolVar(&includeLatest, "include-latest-before-start", false, "Prepend latest candle before start")
	cmd.AddCommand(candles)

	var limit int
	var cursor string
	var minTs, maxTs int64
	var limitSet, minSet, maxSet bool
	trades := &cobra.Command{
		Use:   "trades <ticker>",
		Short: "Public trades for a market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			limitSet = cmd.Flags().Changed("limit")
			minSet = cmd.Flags().Changed("min-ts")
			maxSet = cmd.Flags().Changed("max-ts")
			out, err := rt.client.GetMarginTrades(ctx(), api.TradesParams{
				Ticker: args[0],
				Limit:  ptrInt(limit, limitSet),
				Cursor: cursor,
				MinTs:  ptrInt64(minTs, minSet),
				MaxTs:  ptrInt64(maxTs, maxSet),
			})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(out.Trades))
			for _, tr := range out.Trades {
				rows = append(rows, []string{tr.TradeID, tr.Ticker, tr.Side, tr.Price, tr.Count})
			}
			return rt.out.PrintTable([]string{"TRADE_ID", "TICKER", "SIDE", "PRICE", "COUNT"}, rows, out)
		},
	}
	trades.Flags().IntVar(&limit, "limit", 100, "Page size")
	trades.Flags().StringVar(&cursor, "cursor", "", "Pagination cursor")
	trades.Flags().Int64Var(&minTs, "min-ts", 0, "Min unix ts")
	trades.Flags().Int64Var(&maxTs, "max-ts", 0, "Max unix ts")
	cmd.AddCommand(trades)

	return cmd
}

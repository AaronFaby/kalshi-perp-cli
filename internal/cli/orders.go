package cli

import (
	"fmt"

	"github.com/aaronfaby/kalshi-perp-cli/internal/api"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func newOrdersCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "orders", Short: "Order management"}

	var ticker, status, cursor string
	var limit int
	var minTs, maxTs int64
	var subaccount int
	var all bool
	list := &cobra.Command{
		Use:   "list",
		Short: "List margin orders",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			p := api.OrdersParams{
				Ticker: ticker,
				Status: status,
				Cursor: cursor,
			}
			if cmd.Flags().Changed("limit") {
				p.Limit = &limit
			}
			if cmd.Flags().Changed("min-ts") {
				p.MinTs = &minTs
			}
			if cmd.Flags().Changed("max-ts") {
				p.MaxTs = &maxTs
			}
			if cmd.Flags().Changed("subaccount") {
				p.Subaccount = &subaccount
			}

			var allOrders []api.MarginOrder
			for {
				out, err := rt.client.GetMarginOrders(ctx(), p)
				if err != nil {
					return err
				}
				allOrders = append(allOrders, out.Orders...)
				if !all || out.Cursor == "" {
					rows := make([][]string, 0, len(allOrders))
					for _, o := range allOrders {
						rows = append(rows, []string{o.OrderID, o.Ticker, o.Side, o.Price, o.FillCount, o.RemainingCount, o.ClientOrderID})
					}
					payload := map[string]any{"orders": allOrders, "cursor": out.Cursor}
					return rt.out.PrintTable([]string{"ORDER_ID", "TICKER", "SIDE", "PRICE", "FILL", "REMAIN", "CLIENT_ID"}, rows, payload)
				}
				p.Cursor = out.Cursor
			}
		},
	}
	list.Flags().StringVar(&ticker, "ticker", "", "Filter ticker")
	list.Flags().StringVar(&status, "status", "", "Filter status")
	list.Flags().IntVar(&limit, "limit", 100, "Page size")
	list.Flags().StringVar(&cursor, "cursor", "", "Pagination cursor")
	list.Flags().Int64Var(&minTs, "min-ts", 0, "Min unix ts")
	list.Flags().Int64Var(&maxTs, "max-ts", 0, "Max unix ts")
	list.Flags().IntVar(&subaccount, "subaccount", 0, "Subaccount")
	list.Flags().BoolVar(&all, "all", false, "Follow cursors until exhausted")
	cmd.AddCommand(list)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <order_id>",
		Short: "Get one order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginOrder(ctx(), args[0])
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})

	var (
		createTicker, side, count, price, tif, stp, clientOrderID, orderGroupID string
		expirationMs                                                           int64
		postOnly, reduceOnly, cancelOnPause, dryRun                            bool
		createSub                                                              int
	)
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a margin order",
		RunE: func(cmd *cobra.Command, args []string) error {
			if createTicker == "" || side == "" || count == "" || price == "" || tif == "" || stp == "" {
				return fmt.Errorf("--ticker, --side, --count, --price, --tif, and --stp are required")
			}
			if clientOrderID == "" {
				clientOrderID = uuid.NewString()
			}
			req := api.CreateMarginOrderRequest{
				Ticker:                  createTicker,
				ClientOrderID:           clientOrderID,
				Side:                    side,
				Count:                   count,
				Price:                   price,
				TimeInForce:             tif,
				SelfTradePreventionType: stp,
				OrderGroupID:            orderGroupID,
			}
			if cmd.Flags().Changed("expiration-ms") {
				req.ExpirationTime = &expirationMs
			}
			if cmd.Flags().Changed("post-only") {
				req.PostOnly = &postOnly
			}
			if cmd.Flags().Changed("reduce-only") {
				req.ReduceOnly = &reduceOnly
			}
			if cmd.Flags().Changed("cancel-on-pause") {
				req.CancelOrderOnPause = &cancelOnPause
			}
			if cmd.Flags().Changed("subaccount") {
				req.Subaccount = &createSub
			}
			if dryRun {
				rt, err := opts.setup(false)
				if err != nil {
					return err
				}
				return rt.out.Print(req)
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.CreateMarginOrder(ctx(), req)
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	create.Flags().StringVar(&createTicker, "ticker", "", "Market ticker (required)")
	create.Flags().StringVar(&side, "side", "", "bid|ask (required)")
	create.Flags().StringVar(&count, "count", "", "Size as fixed-point string e.g. 10.00 (required)")
	create.Flags().StringVar(&price, "price", "", "Price as fixed-point dollars e.g. 0.5600 (required)")
	create.Flags().StringVar(&tif, "tif", "", "fill_or_kill|good_till_canceled|immediate_or_cancel (required)")
	create.Flags().StringVar(&stp, "stp", "", "taker_at_cross|maker (required)")
	create.Flags().StringVar(&clientOrderID, "client-order-id", "", "Client order id (auto UUID if omitted)")
	create.Flags().Int64Var(&expirationMs, "expiration-ms", 0, "Expiration unix ms")
	create.Flags().BoolVar(&postOnly, "post-only", false, "Post-only")
	create.Flags().BoolVar(&reduceOnly, "reduce-only", false, "Reduce-only")
	create.Flags().BoolVar(&cancelOnPause, "cancel-on-pause", false, "Cancel if exchange pauses")
	create.Flags().IntVar(&createSub, "subaccount", 0, "Subaccount number")
	create.Flags().StringVar(&orderGroupID, "order-group-id", "", "Order group id")
	create.Flags().BoolVar(&dryRun, "dry-run", false, "Print request body without sending")
	cmd.AddCommand(create)

	cmd.AddCommand(&cobra.Command{
		Use:   "cancel <order_id>",
		Short: "Cancel an order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.CancelMarginOrder(ctx(), args[0])
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})

	var amendTicker, amendSide, amendPrice, amendCount, amendClientID, amendNewClientID string
	amend := &cobra.Command{
		Use:   "amend <order_id>",
		Short: "Amend price and/or max fillable count",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if amendTicker == "" || amendSide == "" || amendPrice == "" || amendCount == "" {
				return fmt.Errorf("--ticker, --side, --price, and --count are required")
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.AmendMarginOrder(ctx(), args[0], api.AmendMarginOrderRequest{
				Ticker:               amendTicker,
				Side:                 amendSide,
				Price:                amendPrice,
				Count:                amendCount,
				ClientOrderID:        amendClientID,
				UpdatedClientOrderID: amendNewClientID,
			})
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	amend.Flags().StringVar(&amendTicker, "ticker", "", "Market ticker (required)")
	amend.Flags().StringVar(&amendSide, "side", "", "bid|ask (required)")
	amend.Flags().StringVar(&amendPrice, "price", "", "New price (required)")
	amend.Flags().StringVar(&amendCount, "count", "", "New max fillable count (required)")
	amend.Flags().StringVar(&amendClientID, "client-order-id", "", "Original client order id")
	amend.Flags().StringVar(&amendNewClientID, "updated-client-order-id", "", "New client order id")
	cmd.AddCommand(amend)

	var reduceBy, reduceTo string
	decrease := &cobra.Command{
		Use:   "decrease <order_id>",
		Short: "Decrease remaining size (exactly one of --reduce-by / --reduce-to)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if (reduceBy == "" && reduceTo == "") || (reduceBy != "" && reduceTo != "") {
				return fmt.Errorf("exactly one of --reduce-by or --reduce-to is required")
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			req := api.DecreaseMarginOrderRequest{}
			if reduceBy != "" {
				req.ReduceBy = &reduceBy
			}
			if reduceTo != "" {
				req.ReduceTo = &reduceTo
			}
			out, err := rt.client.DecreaseMarginOrder(ctx(), args[0], req)
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	decrease.Flags().StringVar(&reduceBy, "reduce-by", "", "Contracts to reduce by")
	decrease.Flags().StringVar(&reduceTo, "reduce-to", "", "Contracts to reduce to")
	cmd.AddCommand(decrease)

	return cmd
}

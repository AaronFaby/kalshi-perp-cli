package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aaronfaby/kalshi-perp-cli/internal/api"
	"github.com/aaronfaby/kalshi-perp-cli/internal/ws"
	"github.com/spf13/cobra"
)

func newGroupsCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "groups", Short: "Order groups"}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List order groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginOrderGroups(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})

	var contractsLimit int64
	var contractsLimitFp string
	var exchangeIndex int
	create := &cobra.Command{
		Use:   "create",
		Short: "Create an order group",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("contracts-limit") && contractsLimitFp == "" {
				return fmt.Errorf("provide --contracts-limit or --contracts-limit-fp")
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			req := api.CreateOrderGroupRequest{}
			if cmd.Flags().Changed("contracts-limit") {
				req.ContractsLimit = &contractsLimit
			}
			if contractsLimitFp != "" {
				req.ContractsLimitFp = &contractsLimitFp
			}
			if cmd.Flags().Changed("exchange-index") {
				req.ExchangeIndex = &exchangeIndex
			}
			out, err := rt.client.CreateMarginOrderGroup(ctx(), req)
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	}
	create.Flags().Int64Var(&contractsLimit, "contracts-limit", 0, "Contracts limit (rolling window)")
	create.Flags().StringVar(&contractsLimitFp, "contracts-limit-fp", "", "Contracts limit as fixed-point string")
	create.Flags().IntVar(&exchangeIndex, "exchange-index", 0, "Exchange index binding")
	cmd.AddCommand(create)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <order_group_id>",
		Short: "Get order group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginOrderGroup(ctx(), args[0])
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <order_group_id>",
		Short: "Delete order group (cancels member orders)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			if err := rt.client.DeleteMarginOrderGroup(ctx(), args[0]); err != nil {
				return err
			}
			return rt.out.Print(map[string]string{"status": "deleted", "order_group_id": args[0]})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reset <order_group_id>",
		Short: "Reset matched contracts counter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			if err := rt.client.ResetMarginOrderGroup(ctx(), args[0]); err != nil {
				return err
			}
			return rt.out.Print(map[string]string{"status": "reset", "order_group_id": args[0]})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "trigger <order_group_id>",
		Short: "Trigger group (cancel all orders)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			if err := rt.client.TriggerMarginOrderGroup(ctx(), args[0]); err != nil {
				return err
			}
			return rt.out.Print(map[string]string{"status": "triggered", "order_group_id": args[0]})
		},
	})

	var limitContracts int64
	var limitFp string
	limitCmd := &cobra.Command{
		Use:   "limit <order_group_id>",
		Short: "Update order group contracts limit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("contracts-limit") && limitFp == "" {
				return fmt.Errorf("provide --contracts-limit or --contracts-limit-fp")
			}
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			req := api.UpdateOrderGroupLimitRequest{}
			if cmd.Flags().Changed("contracts-limit") {
				req.ContractsLimit = &limitContracts
			}
			if limitFp != "" {
				req.ContractsLimitFp = &limitFp
			}
			if err := rt.client.UpdateMarginOrderGroupLimit(ctx(), args[0], req); err != nil {
				return err
			}
			return rt.out.Print(map[string]string{"status": "updated", "order_group_id": args[0]})
		},
	}
	limitCmd.Flags().Int64Var(&limitContracts, "contracts-limit", 0, "New contracts limit")
	limitCmd.Flags().StringVar(&limitFp, "contracts-limit-fp", "", "New limit as fixed-point")
	cmd.AddCommand(limitCmd)

	return cmd
}

func newStreamCmd(opts *RootOptions) *cobra.Command {
	var channels string
	var tickers []string
	var reconnect bool

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream margin WebSocket channels as JSONL",
		Long: `Connect to the margin WebSocket API and print messages as JSON lines.

Channels: orderbook_delta, ticker, trade, fill, user_orders, order_group_updates
Timestamps are Unix milliseconds (*_ms fields).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			chList := splitCSV(channels)
			if len(chList) == 0 {
				return fmt.Errorf("provide --channels")
			}
			for _, ch := range chList {
				if !validChannel(ch) {
					return fmt.Errorf("unknown channel %q", ch)
				}
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			client := &ws.Client{
				URL:        rt.cfg.WSURL,
				KeyID:      rt.cfg.APIKey,
				PrivateKey: rt.key,
			}
			if rt.cfg.Verbose {
				client.Logf = func(format string, a ...any) {
					fmt.Fprintf(os.Stderr, format+"\n", a...)
				}
			}

			params := ws.SubscribeParams{Channels: chList, Tickers: tickers}
			emit := func(msg ws.Message) error {
				return rt.out.PrintLine(msg)
			}
			err = client.RunWithReconnect(ctx, params, reconnect, emit)
			if err != nil && ctx.Err() != nil {
				return nil // graceful interrupt
			}
			return err
		},
	}
	cmd.Flags().StringVar(&channels, "channels", "ticker", "Comma-separated channels")
	cmd.Flags().StringArrayVar(&tickers, "ticker", nil, "Market ticker (repeatable)")
	cmd.Flags().BoolVar(&reconnect, "reconnect", false, "Reconnect with backoff on disconnect")
	return cmd
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func validChannel(ch string) bool {
	switch ch {
	case "orderbook_delta", "ticker", "trade", "fill", "user_orders", "order_group_updates":
		return true
	default:
		return false
	}
}

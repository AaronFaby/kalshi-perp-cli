package cli

import (
	"fmt"

	"github.com/aaronfaby/kalshi-perp-cli/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage local configuration",
	}

	var force bool
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Write a sample config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := opts.ConfigPath
			if path == "" {
				path = config.DefaultConfigPath()
			}
			if err := config.WriteSample(path, force); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", path)
			return nil
		},
	}
	initCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config")
	cmd.AddCommand(initCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print resolved config path",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := opts.load()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), cfg.ConfigPath)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show resolved config (secrets redacted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := opts.load()
			if err != nil {
				return err
			}
			key := cfg.APIKey
			if len(key) > 4 {
				key = "…" + key[len(key)-4:]
			}
			fmt.Fprintf(cmd.OutOrStdout(), "env: %s\napi_key: %s\nprivate_key_path: %s\nbase_url: %s\nws_url: %s\nformat: %s\ntimeout_sec: %d\nconfig: %s\n",
				cfg.Env, key, cfg.PrivateKeyPath, cfg.BaseURL, cfg.WSURL, cfg.Format, cfg.TimeoutSec, cfg.ConfigPath)
			return nil
		},
	})

	return cmd
}

func newAuthCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication helpers",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "whoami",
		Short: "Smoke-test auth (enabled + account limits)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			en, err := rt.client.GetMarginEnabled(ctx())
			if err != nil {
				return err
			}
			limits, err := rt.client.GetPerpsAccountAPILimits(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(map[string]any{
				"env":     rt.cfg.Env,
				"enabled": en,
				"limits":  limits,
			})
		},
	})
	return cmd
}

func newAccountCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "account", Short: "Account endpoints"}
	cmd.AddCommand(&cobra.Command{
		Use:   "limits",
		Short: "Get perps API tier limits",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetPerpsAccountAPILimits(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})
	return cmd
}

func newExchangeCmd(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "exchange", Short: "Exchange status"}
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Margin exchange status",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginExchangeStatus(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "enabled",
		Short: "Whether margin is enabled for this account",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := opts.setup(true)
			if err != nil {
				return err
			}
			out, err := rt.client.GetMarginEnabled(ctx())
			if err != nil {
				return err
			}
			return rt.out.Print(out)
		},
	})
	return cmd
}

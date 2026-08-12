package cli

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/aaronfaby/kalshi-perp-cli/internal/api"
	"github.com/aaronfaby/kalshi-perp-cli/internal/auth"
	"github.com/aaronfaby/kalshi-perp-cli/internal/config"
	"github.com/aaronfaby/kalshi-perp-cli/internal/output"
	"github.com/spf13/cobra"
)

// Version is set via ldflags.
var Version = "dev"

// RootOptions holds global flags.
type RootOptions struct {
	Env            string
	APIKey         string
	PrivateKeyPath string
	BaseURL        string
	WSURL          string
	Format         string
	TimeoutSec     int
	ConfigPath     string
	Verbose        bool
}

// NewRoot builds the kalshi-perp command tree.
func NewRoot() *cobra.Command {
	opts := &RootOptions{}

	root := &cobra.Command{
		Use:           "kalshi-perp",
		Short:         "CLI for the Kalshi perpetual futures (margin) API",
		Long:          "Feature-complete CLI for Kalshi perps: REST trading, portfolio, risk, funding, and WebSocket streaming.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&opts.Env, "env", "", "Environment: demo|prod (default demo)")
	root.PersistentFlags().StringVar(&opts.APIKey, "api-key", "", "API key ID")
	root.PersistentFlags().StringVar(&opts.PrivateKeyPath, "private-key", "", "Path to RSA private key PEM")
	root.PersistentFlags().StringVar(&opts.BaseURL, "base-url", "", "Override REST base URL")
	root.PersistentFlags().StringVar(&opts.WSURL, "ws-url", "", "Override WebSocket URL")
	root.PersistentFlags().StringVar(&opts.Format, "format", "", "Output format: table|json|jsonl")
	root.PersistentFlags().IntVar(&opts.TimeoutSec, "timeout", 0, "HTTP timeout seconds")
	root.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "Config file path")
	root.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Log request method/path/status to stderr")
	root.PersistentFlags().Bool(confirmProdFlag, false, "Required to send mutating requests when targeting production")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newConfigCmd(opts))
	root.AddCommand(newAuthCmd(opts))
	root.AddCommand(newAccountCmd(opts))
	root.AddCommand(newExchangeCmd(opts))
	root.AddCommand(newMarketsCmd(opts))
	root.AddCommand(newOrdersCmd(opts))
	root.AddCommand(newPositionsCmd(opts))
	root.AddCommand(newFillsCmd(opts))
	root.AddCommand(newBalanceCmd(opts))
	root.AddCommand(newRiskCmd(opts))
	root.AddCommand(newFeesCmd(opts))
	root.AddCommand(newFundingCmd(opts))
	root.AddCommand(newTransferCmd(opts))
	root.AddCommand(newSubaccountsCmd(opts))
	root.AddCommand(newGroupsCmd(opts))
	root.AddCommand(newStreamCmd(opts))

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
}

type runtime struct {
	cfg    config.Config
	client *api.Client
	out    *output.Printer
	key    *rsa.PrivateKey
}

func (opts *RootOptions) load() (config.Config, error) {
	return config.Load(opts.ConfigPath, config.Config{
		Env:            opts.Env,
		APIKey:         opts.APIKey,
		PrivateKeyPath: opts.PrivateKeyPath,
		BaseURL:        opts.BaseURL,
		WSURL:          opts.WSURL,
		Format:         opts.Format,
		TimeoutSec:     opts.TimeoutSec,
		Verbose:        opts.Verbose,
		ConfigPath:     opts.ConfigPath,
	})
}

func (opts *RootOptions) setup(needAuth bool) (*runtime, error) {
	cfg, err := opts.load()
	if err != nil {
		return nil, err
	}
	rt := &runtime{
		cfg: cfg,
		out: output.New(cfg.Format),
	}
	if !needAuth {
		return rt, nil
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key required (flag --api-key, env KALSHI_API_KEY, or config)")
	}
	var key *rsa.PrivateKey
	if cfg.PrivateKeyPEM != "" {
		key, err = auth.ParsePrivateKey([]byte(cfg.PrivateKeyPEM))
	} else if cfg.PrivateKeyPath != "" {
		key, err = auth.LoadPrivateKey(cfg.PrivateKeyPath)
	} else {
		return nil, fmt.Errorf("private key required (--private-key, KALSHI_PRIVATE_KEY_PATH, or KALSHI_PRIVATE_KEY)")
	}
	if err != nil {
		return nil, err
	}
	rt.key = key
	rt.client = api.New(cfg.BaseURL, cfg.APIKey, key, time.Duration(cfg.TimeoutSec)*time.Second)
	rt.client.Verbose = cfg.Verbose
	rt.client.Logf = func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
	return rt, nil
}

func ctx() context.Context {
	return context.Background()
}

func exitErr(err error) error {
	if err == nil {
		return nil
	}
	return err
}

func ptrInt(v int, set bool) *int {
	if !set {
		return nil
	}
	return &v
}

func ptrInt64(v int64, set bool) *int64 {
	if !set {
		return nil
	}
	return &v
}

func boolPtr(v bool) *bool { return &v }

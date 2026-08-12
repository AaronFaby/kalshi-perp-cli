package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const confirmProdFlag = "confirm-prod"

func printDryRun(opts *RootOptions, v any) error {
	rt, err := opts.setup(false)
	if err != nil {
		return err
	}
	return rt.out.Print(v)
}

// requireProdConfirm blocks mutating requests against production unless --confirm-prod is set.
func (opts *RootOptions) requireProdConfirm(cmd *cobra.Command) error {
	cfg, err := opts.load()
	if err != nil {
		return err
	}
	if !cfg.IsProduction() {
		return nil
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "warning: targeting production %s\n", cfg.BaseURL)
	ok, err := cmd.Flags().GetBool(confirmProdFlag)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("refusing production mutation against %s; pass --%s", cfg.BaseURL, confirmProdFlag)
	}
	return nil
}

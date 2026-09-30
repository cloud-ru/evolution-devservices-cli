package cmd

import (
	"github.com/spf13/cobra"
)

// newConfigCmd creates the `eds config` command.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show the effective configuration",
		Long: `config prints the configuration values the CLI is using
right now, after applying defaults, the on-disk config and environment
variables. Useful to debug "why is my CLI pointing at the wrong host?".`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}

			return ctx.Printer.Print(map[string]any{
				"project_id":    ctx.Cfg.ProjectID,
				"repo_api_url":  ctx.Cfg.APIURL,
				"api_key":       maskKey(ctx.Cfg.APIKey),
				"repo_git_host": ctx.Cfg.GitHost,
				"wf_api_url":    ctx.Cfg.WorkflowAPIURL,
			})
		},
	}
	return cmd
}

// maskKey returns the API key with most characters hidden.
// Empty strings are returned unchanged.
func maskKey(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + "…" + s[len(s)-4:]
}

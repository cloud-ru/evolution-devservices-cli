// Package cmd implements the `eds` command-line interface.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the root `eds` command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "eds",
		Short: "Evolution DevServices CLI - manage cloud.ru developer tools products",
		Long: `eds is a command-line interface for cloud.ru developer tools
products: Repo (git repositories, "eds repo") and Workflow Studio
(deploy pipelines, "eds wf").

It is designed to be safely driven by automation and AI agents via the
--json flag and the EDS_* environment variables.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags. Repo and Workflow Studio are separate products with
	// separate credentials, so their flags/env vars are namespaced
	// per-product (repo-*, wf-*); --project is shared.
	root.PersistentFlags().String("project", "",
		"Project ID, shared across products (env EDS_PROJECT_ID)")

	root.PersistentFlags().String("repo-api-url", "",
		"Repo product API base URL (default https://devtools.api.cloud.ru/repo/api/v1; env EDS_REPO_API_URL)")
	root.PersistentFlags().String("api-key", "",
		"Repo product API key (env EDS_API_KEY; overrides saved config)")
	root.PersistentFlags().String("repo-git-host", "",
		"Repo product git smart-HTTP host (default repo.cloud.ru; env EDS_REPO_GIT_HOST)")
	root.PersistentFlags().String("wf-api-url", "",
		"Workflow Studio product API base URL (env EDS_WF_API_URL)")

	root.PersistentFlags().Bool("json", false,
		"force JSON output")
	root.PersistentFlags().Bool("yaml", false,
		"force YAML output")
	root.PersistentFlags().BoolP("quiet", "q", false,
		"suppress non-essential output")

	root.AddCommand(newLoginCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newRepoCmd())
	root.AddCommand(newWFCmd())
	root.AddCommand(newVersionCmd())

	return root
}

// Execute runs the root command.
func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
		os.Exit(1)
	}
}

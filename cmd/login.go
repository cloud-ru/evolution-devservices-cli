package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cloud-ru/evolution-devservices-cli/internal/config"
)

// newLoginCmd creates the `eds login` command.
func newLoginCmd() *cobra.Command {
	var (
		apiKey      string
		repoAPIURL  string
		project     string
		repoGitHost string
		wfAPIURL    string
		fromStdin   bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save credentials to the local config",
		Long: `login stores credentials and (optionally) API URLs, project id
and git host in the user's local config (~/.config/eds/config.json
by default, configurable via EDS_CONFIG or XDG_CONFIG_HOME).

It is safe to call this from a script: the file is written with 0600
permissions.`,
		Example: `  # interactive (you'll be prompted for the key)
  eds login --project 3232b2d0-1063-41e6-b2fa-13df767f4a0a

  # non-interactive (preferred for automation / agents)
  eds login --api-key $EDS_API_KEY --project <project-id>

  # dev environment
  eds login --repo-api-url https://devtools.dev.api.internal.cloud.ru/repo/api/v1 \
             --api-key <KEY> --project <project-id>`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("cannot load config: %w", err)
			}

			// Read from stdin if requested.
			if fromStdin {
				if _, err := fmt.Scanln(&apiKey); err != nil {
					return fmt.Errorf("read api key from stdin: %w", err)
				}
			}

			if apiKey == "" && !fromStdin {
				return fmt.Errorf("--api-key or --stdin is required")
			}
			if apiKey != "" {
				cfg.APIKey = apiKey
			}

			if repoAPIURL != "" {
				cfg.APIURL = repoAPIURL
			}
			if project != "" {
				cfg.ProjectID = project
			}
			if repoGitHost != "" {
				cfg.GitHost = repoGitHost
			}
			if wfAPIURL != "" {
				cfg.WorkflowAPIURL = wfAPIURL
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Credentials saved.")
			return nil
		},
	}

	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (shared across products)")
	cmd.Flags().StringVar(&repoAPIURL, "repo-api-url", "", "Repo product API base URL")
	cmd.Flags().StringVar(&project, "project", "", "default project id (shared across products)")
	cmd.Flags().StringVar(&repoGitHost, "repo-git-host", "", "Repo product git smart-HTTP host")
	cmd.Flags().StringVar(&wfAPIURL, "wf-api-url", "", "Workflow Studio product API base URL")
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "read the API key from stdin")

	return cmd
}

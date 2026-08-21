package cmd

import (
	"context"
	"fmt"

	workflowclient "github.com/cloud-ru/evolution-devservices-cli/internal/workflow_client"
	"github.com/spf13/cobra"

	"github.com/cloud-ru/evolution-devservices-cli/internal/config"
	"github.com/cloud-ru/evolution-devservices-cli/internal/output"
	"github.com/cloud-ru/evolution-devservices-cli/internal/repoapi"
)

// runtimeContext holds the resolved configuration, API clients and printer
// for a single command invocation.
type runtimeContext struct {
	Cfg            *config.Config
	API            *repoapi.Client
	Printer        *output.Printer
	Quiet          bool
	ProjectID      string
	WorkflowClient *workflowclient.APIClient
}

// resolveContext loads config, applies flag/env overrides and constructs
// the API client and printer.
func resolveContext(cmd *cobra.Command) (*runtimeContext, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load config: %w", err)
	}

	// Apply flag overrides on top of config (flags > env > file > defaults).
	if v, _ := cmd.Flags().GetString("project"); v != "" {
		cfg.ProjectID = v
	}
	if v, _ := cmd.Flags().GetString("repo-api-url"); v != "" {
		cfg.APIURL = v
	}
	if v, _ := cmd.Flags().GetString("api-key"); v != "" {
		cfg.APIKey = v
	}
	if v, _ := cmd.Flags().GetString("repo-git-host"); v != "" {
		cfg.GitHost = v
	}
	if v, _ := cmd.Flags().GetString("wf-api-url"); v != "" {
		cfg.WorkflowAPIURL = v
	}

	wantJSON, _ := cmd.Flags().GetBool("json")
	quiet, _ := cmd.Flags().GetBool("quiet")

	format := output.FormatAuto
	if wantJSON {
		format = output.FormatJSON
	}

	workflowClientCfg := workflowclient.NewConfiguration()
	workflowClientCfg.AddDefaultHeader("X-API-KEY", cfg.APIKey)
	workflowClientCfg.Servers[0].URL = cfg.WorkflowAPIURL

	rt := &runtimeContext{
		Cfg:            cfg,
		API:            repoapi.New(cfg.APIURL, cfg.APIKey, cfg.ProjectID),
		Printer:        output.New(format),
		Quiet:          quiet,
		ProjectID:      cfg.ProjectID,
		WorkflowClient: workflowclient.NewAPIClient(workflowClientCfg),
	}

	return rt, nil
}

// requireAPIKey returns a friendly error if the API key is missing.
func (r *runtimeContext) requireAPIKey() error {
	if r.Cfg.APIKey == "" {
		return fmt.Errorf("API key is not set. Run `eds login --api-key <KEY>` or set EDS_API_KEY")
	}
	return nil
}

// ensureWorkflowAuth validates that the API key is set before a Workflow
// Studio call. The same API key is used for both products.
func (r *runtimeContext) ensureWorkflowAuth(_ context.Context) error {
	if r.Cfg.APIKey == "" {
		return fmt.Errorf("API key is not set. Run `eds login --api-key <KEY>` or set EDS_API_KEY")
	}
	return nil
}

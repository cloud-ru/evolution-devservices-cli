package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	workflowclient "github.com/cloud-ru/evolution-devservices-cli/internal/workflow_client"
	"github.com/spf13/cobra"
)

// newAppCmd creates the parent `eds wf app` command and all its subcommands.
// An "application" (Workflow Studio) wires a repository+branch to a deploy
// pipeline; creating a deployment for it runs that pipeline and publishes
// the result.
func newAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage Workflow Studio applications (deployed services)",
	}
	cmd.AddCommand(newAppCreateCmd())
	cmd.AddCommand(newAppListCmd())
	cmd.AddCommand(newAppShowCmd())
	cmd.AddCommand(newAppUpdateCmd())
	cmd.AddCommand(newAppDeleteCmd())
	cmd.AddCommand(newAppDeployCmd())
	cmd.AddCommand(newAppDeploymentsCmd())
	return cmd
}

func newAppCreateCmd() *cobra.Command {
	var (
		repository    string
		repositoryURL string
		branch        string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Workflow Studio application from a repository",
		Args:  cobra.ExactArgs(1),
		Long: `create wires a repository + branch to a deploy pipeline.

The repository can be an existing "eds repo" repository (--repository,
accepts either its id or its name) or an external git URL (--repository-url).
Use "eds wf app deploy" afterwards to actually run the pipeline and publish.`,
		Example: `  eds repo create my-site && eds wf app create my-site --repository my-site --branch main
  eds wf app create my-site --repository-url https://github.com/user/my-site --branch main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			if repository == "" && repositoryURL == "" {
				return fmt.Errorf("either --repository or --repository-url is required")
			}

			var repositoryID string
			if repository != "" {
				if !looksLikeUUID(repository) {
					if err := ctx.requireAPIKey(); err != nil {
						return err
					}
				}
				repositoryID, err = resolveRepoID(cmd.Context(), ctx, repository)
				if err != nil {
					return err
				}
			}

			app, _, err := ctx.WorkflowClient.ServicesAPI.ProjectProjectIdApplicationPost(cmd.Context(), ctx.ProjectID).
				Request(workflowclient.GitSbercloudTechDsworksServicesPipelineSrcInternalApplicationRequestApplicationCreate{
					Branch:        branch,
					Name:          &args[0],
					RepositoryId:  &repositoryID,
					RepositoryUrl: &repositoryURL,
					SpaceId:       nil,
				}).Execute()
			if err != nil {
				return fmt.Errorf("workflow client application post: %w", err)
			}

			return ctx.Printer.PrintJSON(app)
		},
	}

	cmd.Flags().StringVar(&repository, "repository", "", "id or name of an existing `eds repo` repository")
	cmd.Flags().StringVar(&repositoryURL, "repository-url", "", "external git repository URL (alternative to --repository)")
	cmd.Flags().StringVar(&branch, "branch", "main", "branch to deploy from")
	return cmd
}

func newAppListCmd() *cobra.Command {
	var (
		search string
		sort   string
		limit  int
		offset int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Workflow Studio applications in the configured project",
		Example: `  eds wf app list
  eds wf app list --search my-site
  eds wf app list --json | jq '.applications[].name'`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			resp, _, err := ctx.WorkflowClient.ServicesAPI.ListApplications(cmd.Context(), ctx.ProjectID).
				Sort(sort).
				Limit(limit).
				Offset(offset).
				Search(search).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client application list: %w", err)
			}

			return ctx.Printer.PrintJSON(resp)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "search term (default: all)")
	cmd.Flags().StringVar(&sort, "sort", "created_at_desc", "sort order: created_at_asc, created_at_desc")
	cmd.Flags().IntVar(&limit, "limit", 50, "page size")
	cmd.Flags().IntVar(&offset, "offset", 0, "offset")
	return cmd
}

func newAppShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <application-id>",
		Short: "Show application details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			app, _, err := ctx.WorkflowClient.ServicesAPI.ProjectProjectIdApplicationApplicationIdGet(cmd.Context(), ctx.ProjectID, args[0]).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client get app: %w", err)
			}

			return ctx.Printer.PrintJSON(app)
		},
	}
	return cmd
}

func newAppUpdateCmd() *cobra.Command {
	var (
		name   string
		branch string
	)

	cmd := &cobra.Command{
		Use:   "update <application-id>",
		Short: "Update an application's name or branch",
		Args:  cobra.ExactArgs(1),
		Example: `  eds wf app update my-app-id --branch release
  eds wf app update my-app-id --name new-name --branch main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}
			if branch == "" {
				return fmt.Errorf("--branch is required")
			}

			_, err = ctx.WorkflowClient.ServicesAPI.ProjectProjectIdApplicationApplicationIdPatch(cmd.Context(), ctx.ProjectID, args[0]).
				Request(workflowclient.GitSbercloudTechDsworksServicesPipelineSrcInternalApplicationRequestApplicationUpdate{
					Branch:  branch,
					Name:    &name,
					SpaceId: "",
				}).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client update application: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new application name")
	cmd.Flags().StringVar(&branch, "branch", "", "new branch to deploy from")
	return cmd
}

func newAppDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <application-id>",
		Short: "Delete an application and its deployments (irreversible)",
		Args:  cobra.ExactArgs(1),
		Example: `  eds wf app delete my-app-id --force
  eds wf app delete my-app-id        # prompts for confirmation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			if !force {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"Are you sure you want to delete application %q? Type its id to confirm: ", args[0])
				reader := bufio.NewReader(os.Stdin)
				confirm, _ := reader.ReadString('\n')
				if strings.TrimSpace(confirm) != args[0] {
					return fmt.Errorf("confirmation failed; aborting")
				}
			}

			_, err = ctx.WorkflowClient.ServicesAPI.ProjectProjectIdApplicationApplicationIdDelete(cmd.Context(), ctx.ProjectID, args[0]).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client delete application: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	return cmd
}

func newAppDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy <application-id>",
		Short: "Run the application's pipeline and publish it",
		Args:  cobra.ExactArgs(1),
		Long: `deploy triggers a new run of the application's deploy pipeline.
Use the returned run_id with "eds wf run show" to
follow progress, and check the "url" field once it succeeds.`,
		Example: `  eds wf app deploy my-app-id --json | jq -r '.run_id'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			dep, _, err := ctx.WorkflowClient.ServicesAPI.ProjectProjectIdApplicationApplicationIdDeploymentPost(cmd.Context(), ctx.ProjectID, args[0]).
				Request(map[string]any{}).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client deploy application: %w", err)
			}

			return ctx.Printer.PrintJSON(dep)
		},
	}
	return cmd
}

func newAppDeploymentsCmd() *cobra.Command {
	var (
		search string
		sort   string
		limit  int
		offset int
	)

	cmd := &cobra.Command{
		Use:   "deployments <application-id>",
		Short: "List deployments (publish history) for an application",
		Args:  cobra.ExactArgs(1),
		Example: `  eds wf app deployments my-app-id
  eds wf app deployments my-app-id --sort created_at_desc --limit 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			resp, _, err := ctx.WorkflowClient.ServicesAPI.ProjectProjectIdApplicationApplicationIdDeploymentListGet(cmd.Context(), ctx.ProjectID, args[0]).
				Limit(limit).
				Offset(offset).
				Search(search).
				Sort(sort).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client list application: %w", err)
			}

			return ctx.Printer.PrintJSON(resp)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "search term (default: all)")
	cmd.Flags().StringVar(&sort, "sort", "created_at_desc", "sort order: created_at_asc, created_at_desc")
	cmd.Flags().IntVar(&limit, "limit", 50, "page size")
	cmd.Flags().IntVar(&offset, "offset", 0, "offset")
	return cmd
}

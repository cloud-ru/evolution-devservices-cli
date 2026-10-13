package cmd

import (
	"fmt"

	workflowclient "github.com/cloud-ru/evolution-devservices-cli/internal/workflow_client"
	"github.com/spf13/cobra"
)

// newSpaceCmd creates the parent `eds wf space` command.
func newSpaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Manage Workflow Studio application spaces",
	}
	cmd.AddCommand(newSpaceListCmd())
	cmd.AddCommand(newSpaceCreateCmd())
	cmd.AddCommand(newSpaceShowCmd())
	cmd.AddCommand(newSpaceUpdateCmd())
	cmd.AddCommand(newSpaceDeleteCmd())
	return cmd
}

func newSpaceListCmd() *cobra.Command {
	var (
		search string
		sort   string
		limit  int
		offset int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List application spaces in the configured project",
		Example: `  eds wf space list
  eds wf space list --search default`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			resp, _, err := ctx.WorkflowClient.ApplicationSpaceAPI.ListApplicationSpaces(cmd.Context(), ctx.ProjectID).
				Search(search).
				Sort(sort).
				Limit(limit).
				Offset(offset).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client space list: %w", err)
			}

			return ctx.Printer.Print(resp)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "search term (default: all)")
	cmd.Flags().StringVar(&sort, "sort", "", "sort order: created_at_asc, created_at_desc")
	cmd.Flags().IntVar(&limit, "limit", 50, "page size")
	cmd.Flags().IntVar(&offset, "offset", 0, "offset")
	return cmd
}

func newSpaceCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create an application space",
		Example: `  eds wf space create --name my-space`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			resp, _, err := ctx.WorkflowClient.ApplicationSpaceAPI.CreateApplicationSpace(cmd.Context(), ctx.ProjectID).
				Request(workflowclient.GitSbercloudTechDsworksServicesPipelineSrcInternalApplicationRequestApplicationSpaceCreate{
					Name: name,
				}).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client space create: %w", err)
			}

			return ctx.Printer.Print(resp)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "space name (required)")
	return cmd
}

func newSpaceShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show <space_id>",
		Short:   "Show an application space",
		Args:    cobra.ExactArgs(1),
		Example: `  eds wf space show my-space-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			resp, _, err := ctx.WorkflowClient.ApplicationSpaceAPI.GetApplicationSpace(cmd.Context(), ctx.ProjectID, args[0]).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client space show: %w", err)
			}

			return ctx.Printer.Print(resp)
		},
	}

	return cmd
}

func newSpaceUpdateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:     "update <space_id>",
		Short:   "Update an application space",
		Args:    cobra.ExactArgs(1),
		Example: `  eds wf space update my-space-id --name new-name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			resp, _, err := ctx.WorkflowClient.ApplicationSpaceAPI.UpdateApplicationSpace(cmd.Context(), ctx.ProjectID, args[0]).
				Request(workflowclient.GitSbercloudTechDsworksServicesPipelineSrcInternalApplicationRequestApplicationSpaceUpdate{
					Name: name,
				}).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client space update: %w", err)
			}

			return ctx.Printer.Print(resp)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new space name (required)")
	return cmd
}

func newSpaceDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <space_id>",
		Short:   "Delete an application space",
		Args:    cobra.ExactArgs(1),
		Example: `  eds wf space delete my-space-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			if _, err := ctx.WorkflowClient.ApplicationSpaceAPI.DeleteApplicationSpace(cmd.Context(), ctx.ProjectID, args[0]).
				Execute(); err != nil {
				return fmt.Errorf("workflow client space delete: %w", err)
			}

			return nil
		},
	}

	return cmd
}

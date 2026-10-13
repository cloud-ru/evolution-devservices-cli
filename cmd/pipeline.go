package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newPipelineCmd creates the parent `eds wf pipeline` command.
func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Manage Workflow Studio pipelines",
	}
	cmd.AddCommand(newPipelineListCmd())
	cmd.AddCommand(newPipelineShowCmd())
	return cmd
}

func newPipelineListCmd() *cobra.Command {
	var (
		search string
		sort   string
		limit  int
		offset int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipelines in the configured project",
		Example: `  eds wf pipeline list
  eds wf pipeline list --search my-pipeline --limit 10`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			resp, _, err := ctx.WorkflowClient.PipelinesAPI.ListPipelines(cmd.Context(), ctx.ProjectID).
				Search(search).
				Sort(sort).
				Limit(limit).
				Offset(offset).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client pipeline list: %w", err)
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

func newPipelineShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show <pipeline-id>",
		Short:   "Show pipeline details",
		Args:    cobra.ExactArgs(1),
		Example: `  eds wf pipeline show my-pipeline-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			pipe, _, err := ctx.WorkflowClient.PipelinesAPI.GetPipeline(cmd.Context(), ctx.ProjectID, args[0]).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client get pipeline: %w", err)
			}

			return ctx.Printer.Print(pipe)
		},
	}
	return cmd
}

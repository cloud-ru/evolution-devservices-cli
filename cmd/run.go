package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newRunCmd creates the parent `eds wf run` command and all its subcommands.
// A run is a single execution of a Workflow Studio pipeline (e.g. the
// pipeline behind an application's deploy).
func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Inspect and control Workflow Studio pipeline runs",
	}
	cmd.AddCommand(newRunListCmd())
	cmd.AddCommand(newRunShowCmd())
	cmd.AddCommand(newRunStopCmd())
	return cmd
}

func newRunListCmd() *cobra.Command {
	var (
		pipelineID string
		sort       string
		limit      int
		offset     int
		withConfig bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Workflow Studio pipeline runs",
		Example: `  eds wf run list
  eds wf run list --pipeline-id my-pipeline-id --limit 10`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			req := ctx.WorkflowClient.WorkflowsAPI.ListRuns(cmd.Context(), ctx.ProjectID).
				PipelineId(pipelineID).
				Sort(sort).
				Limit(limit).
				Offset(offset)
			if cmd.Flags().Changed("with-config") {
				req = req.WithConfig(withConfig)
			}

			resp, _, err := req.Execute()
			if err != nil {
				return fmt.Errorf("workflow client run list: %w", err)
			}

			return ctx.Printer.Print(resp)
		},
	}

	cmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "filter by pipeline ID")
	cmd.Flags().StringVar(&sort, "sort", "", "sort order")
	cmd.Flags().IntVar(&limit, "limit", 50, "page size")
	cmd.Flags().IntVar(&offset, "offset", 0, "offset")
	cmd.Flags().BoolVar(&withConfig, "with-config", false, "include run config in the response")
	return cmd
}

func newRunShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <run-id>",
		Short: "Show a run's status, stages and jobs",
		Args:  cobra.ExactArgs(1),
		Example: `  eds wf run show my-run-id
  eds wf run show my-run-id --json | jq '.status'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			run, _, err := ctx.WorkflowClient.WorkflowsAPI.GetRun(cmd.Context(), ctx.ProjectID, args[0]).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client get run: %w", err)
			}

			return ctx.Printer.Print(run)
		},
	}
	return cmd
}

func newRunStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <run-id>",
		Short: "Stop a running run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			_, _, err = ctx.WorkflowClient.WorkflowsAPI.StopRun(cmd.Context(), ctx.ProjectID, args[0]).Execute()
			if err != nil {
				return fmt.Errorf("stop run: %w", err)
			}

			return ctx.Printer.Print("run has been stopped")
		},
	}
	return cmd
}

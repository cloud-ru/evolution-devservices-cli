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
	cmd.AddCommand(newRunShowCmd())
	cmd.AddCommand(newRunStopCmd())
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

			err = ctx.Printer.PrintJSON(run)
			if err != nil {
				return fmt.Errorf("print json run: %w", err)
			}

			return nil
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

			fmt.Fprintf(cmd.OutOrStdout(), "Stopped run %s\n", args[0])
			return nil
		},
	}
	return cmd
}

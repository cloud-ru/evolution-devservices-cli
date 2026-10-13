package cmd

import (
	"fmt"
	"time"

	workflowclient "github.com/cloud-ru/evolution-devservices-cli/internal/workflow_client"
	"github.com/spf13/cobra"
)

// newJobCmd creates the parent `eds wf job` command and all its subcommands.
// A job is a single unit of work within a run's stage (e.g. build, push,
// publish).
func newJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Inspect and control Workflow Studio jobs",
	}

	cmd.AddCommand(newJobListCmd())
	cmd.AddCommand(newJobShowCmd())
	cmd.AddCommand(newJobStopCmd())
	cmd.AddCommand(newJobLogsCmd())

	return cmd
}

func newJobListCmd() *cobra.Command {
	var (
		runID  string
		limit  int
		offset int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs for a run",
		Example: `  eds wf job list --run-id my-run-id
  eds wf job list --run-id my-run-id --limit 20`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}
			if runID == "" {
				return fmt.Errorf("--run-id is required")
			}

			resp, _, err := ctx.WorkflowClient.PipelinesAPI.ListJobs(cmd.Context(), ctx.ProjectID).
				RunId(runID).
				Limit(limit).
				Offset(offset).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client job list: %w", err)
			}

			return ctx.Printer.Print(resp)
		},
	}

	cmd.Flags().StringVar(&runID, "run-id", "", "run ID to list jobs for (required)")
	cmd.Flags().IntVar(&limit, "limit", 50, "page size")
	cmd.Flags().IntVar(&offset, "offset", 0, "offset")
	return cmd
}

func newJobShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show <job-id>",
		Short:   "Show job details",
		Args:    cobra.ExactArgs(1),
		Example: `  eds wf job show my-job-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			job, _, err := ctx.WorkflowClient.WorkflowsAPI.GetJob(cmd.Context(), ctx.ProjectID, args[0]).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client get job: %w", err)
			}

			return ctx.Printer.Print(job)
		},
	}
	return cmd
}

func newJobStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <job-id>",
		Short: "Stop a running job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			_, _, err = ctx.WorkflowClient.WorkflowsAPI.StopJob(cmd.Context(), ctx.ProjectID, args[0]).Execute()
			if err != nil {
				return fmt.Errorf("stop job: %w", err)
			}

			return ctx.Printer.Print("job has been stopped")
		},
	}
	return cmd
}

func newJobLogsCmd() *cobra.Command {
	var (
		limit      int
		beforeTime string
	)

	cmd := &cobra.Command{
		Use:   "logs <job-id>",
		Short: "Get logs for a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}

			resp, _, err := ctx.WorkflowClient.WorkflowsAPI.
				GetJobLogs(cmd.Context(), ctx.ProjectID, args[0]).
				Request(workflowclient.GitSbercloudTechDsworksServicesPipelineSrcInternalApplicationRequestJobLogList{
					BeforeTime: &beforeTime,
					Limit:      &limit,
				}).Execute()
			if err != nil {
				return fmt.Errorf("workflow client logs: %w", err)
			}

			return ctx.Printer.Print(resp)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "page size, default is 10")
	cmd.Flags().StringVar(&beforeTime, "before", time.Now().Format(time.RFC3339), "RFC3339 time, default is now")

	return cmd
}

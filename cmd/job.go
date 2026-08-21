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

	cmd.AddCommand(newJobLogsCmd())

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

			err = ctx.Printer.PrintJSON(resp)
			if err != nil {
				return fmt.Errorf("print json: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "page size, default is 10")
	cmd.Flags().StringVar(&beforeTime, "before", time.Now().Format(time.RFC3339), "RFC3339 time, default is now")

	return cmd
}

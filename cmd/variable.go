package cmd

import (
	"fmt"

	workflowclient "github.com/cloud-ru/evolution-devservices-cli/internal/workflow_client"
	"github.com/spf13/cobra"
)

// newVariableCmd creates the parent `eds wf variable` command.
func newVariableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage Workflow Studio pipeline variables",
	}
	cmd.AddCommand(newVariableListCmd())
	cmd.AddCommand(newVariableCreateCmd())
	cmd.AddCommand(newVariableUpdateCmd())
	cmd.AddCommand(newVariableDeleteCmd())
	return cmd
}

func newVariableListCmd() *cobra.Command {
	var pipelineID string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List variables for a pipeline",
		Example: `  eds wf variable list --pipeline-id my-pipeline-id`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}
			if pipelineID == "" {
				return fmt.Errorf("--pipeline-id is required")
			}

			resp, _, err := ctx.WorkflowClient.PipelinesAPI.ListVariables(cmd.Context(), ctx.ProjectID, pipelineID).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client variable list: %w", err)
			}

			for i := range resp.Variables {
				resp.Variables[i].Value = new("Value is hidden. Use Web-UI to view it")
			}

			return ctx.Printer.Print(resp)
		},
	}

	cmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "pipeline ID to list variables for (required)")
	return cmd
}

func newVariableCreateCmd() *cobra.Command {
	var pipelineID string

	cmd := &cobra.Command{
		Use:     "create <name> <value>",
		Short:   "Create a variable for a pipeline",
		Args:    cobra.ExactArgs(2),
		Example: `  eds wf variable create MY_VAR my_value --pipeline-id my-pipeline-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}
			if pipelineID == "" {
				return fmt.Errorf("--pipeline-id is required")
			}

			_, err = ctx.WorkflowClient.PipelinesAPI.CreateVariable(cmd.Context(), ctx.ProjectID, pipelineID).
				Request(workflowclient.GitSbercloudTechDsworksServicesPipelineSrcInternalApplicationRequestVariableCreate{
					Name:  args[0],
					Value: args[1],
				}).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client variable create: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "pipeline ID (required)")
	return cmd
}

func newVariableUpdateCmd() *cobra.Command {
	var pipelineID string

	cmd := &cobra.Command{
		Use:     "update <name> <value>",
		Short:   "Update a pipeline variable",
		Args:    cobra.ExactArgs(2),
		Example: `  eds wf variable update MY_VAR new_value --pipeline-id my-pipeline-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}
			if pipelineID == "" {
				return fmt.Errorf("--pipeline-id is required")
			}

			_, err = ctx.WorkflowClient.PipelinesAPI.UpdateVariable(cmd.Context(), ctx.ProjectID, pipelineID).
				Name(args[0]).
				Request(workflowclient.GitSbercloudTechDsworksServicesPipelineSrcInternalApplicationRequestVariableUpdate{
					Value: args[1],
				}).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client variable update: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "pipeline ID (required)")
	return cmd
}

func newVariableDeleteCmd() *cobra.Command {
	var pipelineID string

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete a pipeline variable",
		Args:    cobra.ExactArgs(1),
		Example: `  eds wf variable delete MY_VAR --pipeline-id my-pipeline-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.ensureWorkflowAuth(cmd.Context()); err != nil {
				return err
			}
			if pipelineID == "" {
				return fmt.Errorf("--pipeline-id is required")
			}

			_, err = ctx.WorkflowClient.PipelinesAPI.DeleteVariable(cmd.Context(), ctx.ProjectID, pipelineID).
				Name(args[0]).
				Execute()
			if err != nil {
				return fmt.Errorf("workflow client variable delete: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "pipeline ID (required)")
	return cmd
}

package cmd

import (
	"github.com/spf13/cobra"
)

// newWFCmd creates the parent `eds wf` command for the Workflow Studio
// product: applications (services wired to a deploy pipeline), their runs,
// and the jobs within a run.
func newWFCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wf",
		Short: "Manage Workflow Studio applications, runs and jobs",
	}
	cmd.AddCommand(newAppCmd())
	cmd.AddCommand(newPipelineCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newSpaceCmd())
	cmd.AddCommand(newVariableCmd())
	cmd.AddCommand(newJobCmd())
	return cmd
}

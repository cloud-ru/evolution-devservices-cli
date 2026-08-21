package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/cloud-ru/evolution-devservices-cli/internal/output"
	"github.com/cloud-ru/evolution-devservices-cli/internal/repoapi"
)

// newRepoCmd creates the parent `eds repo` command and all its subcommands.
func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage git repositories",
	}
	cmd.AddCommand(newRepoListCmd())
	cmd.AddCommand(newRepoCreateCmd())
	cmd.AddCommand(newRepoShowCmd())
	cmd.AddCommand(newRepoDeleteCmd())
	cmd.AddCommand(newRepoCloneCmd())
	return cmd
}

func newRepoListCmd() *cobra.Command {
	var (
		search string
		sort   string
		limit  int
		offset int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List git repositories in the configured project",
		Example: `  eds repo list
  eds repo list --search demo
  eds repo list --sort updated_at_desc --limit 50
  eds repo list --json | jq '.repositories[].name'`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.requireAPIKey(); err != nil {
				return err
			}

			opts := repoapi.ListOptions{
				Search: search,
				Sort:   repoapi.SortOrder(sort),
				Limit:  limit,
				Offset: offset,
			}
			// If the user did not provide --search, omit the parameter
			// entirely. Sending a single space is interpreted by the
			// server as a literal match and yields an empty list.
			if opts.Limit <= 0 {
				opts.Limit = 50
			}

			resp, err := ctx.API.ListRepositories(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("api list repositories: %w", err)
			}

			if ctx.Printer.Format == 1 /* FormatJSON */ {
				return ctx.Printer.PrintJSON(resp)
			}

			headers := []string{"ID", "NAME", "TYPE", "VISIBILITY", "UPDATED"}
			rows := make([][]string, 0, len(resp.Repositories))
			for _, r := range resp.Repositories {
				rows = append(rows, []string{
					r.ID,
					r.Name,
					r.Type,
					r.Visibility,
					output.HumanTime(r.UpdatedAt),
				})
			}
			ctx.Printer.Table(headers, rows)
			fmt.Fprintf(cmd.OutOrStdout(), "Showing %d of %d\n", len(resp.Repositories), resp.Total)
			return nil
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "search term (default: all)")
	cmd.Flags().StringVar(&sort, "sort", "updated_at_desc",
		"sort order: name_asc, name_desc, updated_at_asc, updated_at_desc")
	cmd.Flags().IntVar(&limit, "limit", 50, "page size")
	cmd.Flags().IntVar(&offset, "offset", 0, "offset")
	return cmd
}

func newRepoCreateCmd() *cobra.Command {
	var (
		description string
		visibility  string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new git repository",
		Args:  cobra.ExactArgs(1),
		Example: `  eds repo create demo-service --description "my new repo"
  eds repo create demo-service --visibility private`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.requireAPIKey(); err != nil {
				return err
			}

			req := repoapi.CreateRepositoryRequest{
				Name:            args[0],
				Description:     description,
				Type:            string(repoapi.TypeGit),
				VisibilityLevel: visibility,
			}

			r, err := ctx.API.CreateRepository(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("api create repository: %w", err)
			}

			if !ctx.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(),
					"Created repository %q (id=%s)\n", r.Name, r.ID)
			}
			return ctx.Printer.PrintJSON(r)
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "repository description")
	cmd.Flags().StringVar(&visibility, "visibility", "private",
		"visibility: private or shadow")
	return cmd
}

func newRepoShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <repository-id>",
		Short: "Show repository details",
		Args:  cobra.ExactArgs(1),
		Example: `  eds repo show 3232b2d0-1063-41e6-b2fa-13df767f4a0a
  eds repo show my-repo --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.requireAPIKey(); err != nil {
				return err
			}

			id, err := resolveRepoID(cmd.Context(), ctx, args[0])
			if err != nil {
				return err
			}

			info, err := ctx.API.GetRepository(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("api get repository: %w", err)
			}

			if ctx.Printer.Format == 1 /* FormatJSON */ {
				return ctx.Printer.PrintJSON(info)
			}
			ctx.Printer.KeyValue([][2]string{
				{"id", info.ID},
				{"name", info.Name},
				{"description", info.Description},
				{"type", info.Type},
				{"default_branch", info.DefaultBranch},
				{"branches", fmt.Sprintf("%d", info.BranchesCount)},
				{"commits", fmt.Sprintf("%d", info.CommitsCount)},
				{"size", output.HumanSize(info.Size)},
				{"created_at", output.HumanTime(info.CreatedAt)},
				{"updated_at", output.HumanTime(info.UpdatedAt)},
				{"clone_https", info.Clone.HTTPS},
				{"clone_ssh", info.Clone.SSH},
			})
			return nil
		},
	}
	return cmd
}

func newRepoDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <repository-id>",
		Short: "Delete a repository (irreversible)",
		Args:  cobra.ExactArgs(1),
		Example: `  eds repo delete my-old-repo --force
  eds repo delete my-old-repo        # prompts for confirmation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.requireAPIKey(); err != nil {
				return err
			}

			id, err := resolveRepoID(cmd.Context(), ctx, args[0])
			if err != nil {
				return err
			}

			if !force {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"Are you sure you want to delete repository %q? Type its name to confirm: ", args[0])
				reader := bufio.NewReader(os.Stdin)
				confirm, _ := reader.ReadString('\n')
				if strings.TrimSpace(confirm) != args[0] {
					return fmt.Errorf("confirmation failed; aborting")
				}
			}

			if err := ctx.API.DeleteRepository(cmd.Context(), id); err != nil {
				return fmt.Errorf("api delete repository: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted repository %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	return cmd
}

func newRepoCloneCmd() *cobra.Command {
	var (
		target string
		ssh    bool
	)

	cmd := &cobra.Command{
		Use:   "clone <repository-id-or-name> [directory]",
		Short: "Clone a repository using the local git CLI",
		Long: `clone looks up the repository via the API and then runs
"git clone" against the smart-HTTP (or SSH) URL it returns.

This command does not upload files - use git add/commit/push after
the clone, as you would with any other git server.`,
		Args: cobra.RangeArgs(1, 2),
		Example: `  eds repo clone my-repo
  eds repo clone my-repo ./work/my-repo
  eds repo clone my-repo --ssh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := resolveContext(cmd)
			if err != nil {
				return err
			}
			if err := ctx.requireAPIKey(); err != nil {
				return err
			}

			// Try to resolve the argument as either an id or a name.
			id, err := resolveRepoID(cmd.Context(), ctx, args[0])
			if err != nil {
				return err
			}

			info, err := ctx.API.GetRepository(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("api get repository: %w", err)
			}

			cloneURL := info.Clone.HTTPS
			if ssh && info.Clone.SSH != "" {
				cloneURL = info.Clone.SSH
			}
			if cloneURL == "" {
				return fmt.Errorf("repository has no clone url")
			}

			dst := target
			if dst == "" && len(args) == 2 {
				dst = args[1]
			}

			gitArgs := []string{"clone", cloneURL}
			if dst != "" {
				gitArgs = append(gitArgs, dst)
			}

			if !ctx.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Running: git %s\n", strings.Join(gitArgs, " "))
			}

			c := exec.CommandContext(cmd.Context(), "git", gitArgs...)
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.ErrOrStderr()
			c.Stdin = os.Stdin
			return c.Run()
		},
	}

	cmd.Flags().StringVar(&target, "target", "", "target directory (defaults to repo name)")
	cmd.Flags().BoolVar(&ssh, "ssh", false, "clone via SSH instead of HTTPS")
	return cmd
}

// resolveRepoID accepts either a raw repository id (UUID) or a repository
// name and returns the id. Names are resolved against the configured
// project's listing.
func resolveRepoID(ctx context.Context, r *runtimeContext, ref string) (string, error) {
	// Cheap heuristic: ids are typically long UUIDs. The API accepts names
	// in some endpoints, but to be safe we always resolve by listing when
	// the value does not look like a UUID.
	if looksLikeUUID(ref) {
		return ref, nil
	}

	resp, err := r.API.ListRepositories(ctx, repoapi.ListOptions{
		Search: ref,
		Sort:   repoapi.SortNameAsc,
		Limit:  10,
	})
	if err != nil {
		return "", fmt.Errorf("api list repositories: %w", err)
	}
	for _, repo := range resp.Repositories {
		if strings.EqualFold(repo.Name, ref) {
			return repo.ID, nil
		}
	}
	return "", fmt.Errorf("repository %q not found in project %s", ref, r.Cfg.ProjectID)
}

func looksLikeUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

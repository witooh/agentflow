package cli

import (
	"fmt"

	"agentflow/internal/buildinfo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd creates the root cobra command for the agentflow CLI.
// It wires minimal flags and binds them to the provided Viper instance.
func NewRootCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "agentflow",
		Short:         "agentflow CLI",
		Long:          "agentflow CLI — project automation toolkit (skeleton)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If --version is requested, print version info and exit.
			showVersion, _ := cmd.Flags().GetBool("version")
			if showVersion {
				fmt.Fprintln(cmd.OutOrStdout(), buildinfo.Summary())
				return nil
			}

			// If --json is requested at the root, emit a minimal placeholder JSON payload.
			if v.GetBool("json") {
				// Minimal schema for now; subject to change as commands are implemented.
				payload := fmt.Sprintf(`{"name":"agentflow","version":"%s","capabilities":["help","version"],"ok":true}`, buildinfo.Version)
				fmt.Fprintln(cmd.OutOrStdout(), payload)
				return nil
			}

			// Default: show help if no subcommands are provided.
			return cmd.Help()
		},
	}

	// Flags
	cmd.PersistentFlags().String("project-dir", ".", "Project root directory (default: current working directory)")
	cmd.PersistentFlags().Bool("json", false, "Output JSON when supported")
	cmd.Flags().Bool("version", false, "Show version information and exit")

	// Bind flags to Viper
	_ = v.BindPFlag("project_dir", cmd.PersistentFlags().Lookup("project-dir"))
	_ = v.BindPFlag("json", cmd.PersistentFlags().Lookup("json"))

	// Set some sane defaults
	v.SetDefault("project_dir", ".")
	v.SetDefault("json", false)

	// Example of using viper value in a template or pre-run
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		// No-op, but keep for future wiring. Demonstrate value access:
		_ = fmt.Sprintf("project_dir=%s", v.GetString("project_dir"))
	}

	return cmd
}

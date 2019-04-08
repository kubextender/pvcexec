package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func GetVersion() string {
	return fmt.Sprintf("Version=%v (commit %v, built at %v)", version, commit, date)
}

// NewVersionCommand provides the version command.
func NewVersionCommand(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information for pvcexec",
		RunE: func(c *cobra.Command, args []string) error {
			fmt.Fprintln(streams.Out, GetVersion())
			return nil
		},
	}
	return cmd
}
